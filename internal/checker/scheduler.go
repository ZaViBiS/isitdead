package checker

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/ZaViBiS/isitdead/internal/database"
	"github.com/ZaViBiS/isitdead/internal/model"
)

// Scheduler керує циклічними перевірками серверів
type Scheduler struct {
	storage         *database.Storage
	notifier        TransitionNotifier
	regionalChecker RegionalChecker
	localRegion     string
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	mu              sync.Mutex
	monitors        map[uint]monitorControl
}

type monitorControl struct {
	cancel context.CancelFunc
	done   chan struct{}
}

type lastResult struct {
	Status  string
	Latency int64
}

type RegionResult struct {
	Region  string
	Status  string
	Latency int64
}

type RegionalChecker interface {
	CheckRegions(ctx context.Context, checkType, target string, timeoutSeconds int) []RegionResult
}

type TransitionNotifier interface {
	Notify(ctx context.Context, server model.Server, previousState, currentState string, latency int64) error
	NotifySSL(ctx context.Context, server model.Server, status model.SSLCertificateStatus, event string) error
}

// NewScheduler створює новий екземпляр планувальника
func NewScheduler(db *database.Storage) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		storage:     db,
		localRegion: "de",
		ctx:         ctx,
		cancel:      cancel,
		monitors:    make(map[uint]monitorControl),
	}
}

func (s *Scheduler) SetNotifier(notifier TransitionNotifier) {
	s.notifier = notifier
}

func (s *Scheduler) SetLocalRegion(region string) {
	region = strings.TrimSpace(region)
	if region == "" || region == model.CheckRegionGlobal || region == model.CheckRegionAll {
		region = "de"
	}
	s.localRegion = region
}

func (s *Scheduler) SetRegionalChecker(checker RegionalChecker) {
	s.regionalChecker = checker
}

// Start запускає моніторинг для всіх серверів у базі
func (s *Scheduler) Start() error {
	servers, err := s.storage.GetAllServers()
	if err != nil {
		return err
	}

	log.Info().Int("count", len(servers)).Msg("Starting scheduler for servers")

	for _, server := range servers {
		if err := s.storage.EnsureDefaultNotificationPreferences(server.UserID, server.ID); err != nil {
			return err
		}
		s.RunServerMonitor(server)
	}

	s.wg.Add(1)
	go s.runSSLScheduler()

	return nil
}

func (s *Scheduler) runSSLScheduler() {
	defer s.wg.Done()
	s.runSSLChecks()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.runSSLChecks()
		}
	}
}

func (s *Scheduler) runSSLChecks() {
	servers, err := s.storage.GetSSLEnabledServers()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load SSL-enabled servers")
		return
	}
	for _, server := range servers {
		s.RunSSLCheck(server)
	}
}

func (s *Scheduler) RunSSLCheck(server model.Server) {
	previous, _ := s.storage.GetSSLCertificateStatus(server.ID)
	info := InspectSSLCertificate(server.URL, connectionTimeout(server.Timeout))
	lastNotifiedThreshold := retainedSSLNotificationThreshold(previous, info)
	status := model.SSLCertificateStatus{
		ServerID:              server.ID,
		Valid:                 info.Valid,
		SelfSigned:            info.SelfSigned,
		ExpiresAt:             info.ExpiresAt,
		DaysRemaining:         info.DaysRemaining,
		Issuer:                info.Issuer,
		Fingerprint:           info.Fingerprint,
		LastError:             info.Error,
		LastNotifiedThreshold: lastNotifiedThreshold,
		LastCheckedAt:         time.Now().UTC(),
	}
	if err := s.storage.UpsertSSLCertificateStatus(status); err != nil {
		log.Error().Err(err).Uint("server_id", server.ID).Msg("Failed to save SSL certificate status")
		return
	}
	if s.notifier == nil || !status.Valid || status.ExpiresAt == nil {
		return
	}
	event, threshold, ok := sslReminder(status.DaysRemaining, status.LastNotifiedThreshold)
	if !ok {
		return
	}
	if err := s.notifier.NotifySSL(s.ctx, server, status, event); err != nil {
		log.Error().Err(err).Uint("server_id", server.ID).Msg("Failed to process SSL notification")
		return
	}
	status.LastNotifiedThreshold = threshold
	if err := s.storage.UpsertSSLCertificateStatus(status); err != nil {
		log.Error().Err(err).Uint("server_id", server.ID).Msg("Failed to save SSL notification threshold")
	}
}

func retainedSSLNotificationThreshold(previous *model.SSLCertificateStatus, current SSLCertificateInfo) int {
	if previous == nil || previous.ExpiresAt == nil || current.ExpiresAt == nil {
		return 0
	}
	if previous.Fingerprint != current.Fingerprint || !previous.ExpiresAt.Equal(*current.ExpiresAt) {
		return 0
	}
	return previous.LastNotifiedThreshold
}

func sslReminder(daysRemaining, lastNotifiedThreshold int) (event string, threshold int, ok bool) {
	switch {
	case daysRemaining <= 7 && lastNotifiedThreshold != 7:
		return model.NotificationEventSSL7d, 7, true
	case daysRemaining <= 14 && daysRemaining > 7 && lastNotifiedThreshold != 14:
		return model.NotificationEventSSL14d, 14, true
	case daysRemaining <= 30 && daysRemaining > 14 && lastNotifiedThreshold != 30:
		return model.NotificationEventSSL30d, 30, true
	default:
		return "", 0, false
	}
}

// Stop зупиняє всі процеси моніторингу
func (s *Scheduler) Stop() {
	s.cancel()
	s.mu.Lock()
	for id, monitor := range s.monitors {
		monitor.cancel()
		delete(s.monitors, id)
	}
	s.mu.Unlock()
	s.wg.Wait()
}

// RunServerMonitor запускає окрему горутину для моніторингу конкретного сервера
func (s *Scheduler) RunServerMonitor(srv model.Server) {
	s.StopServerMonitor(srv.ID)

	ctx, cancel := context.WithCancel(s.ctx)
	done := make(chan struct{})

	s.mu.Lock()
	s.monitors[srv.ID] = monitorControl{cancel: cancel, done: done}
	s.wg.Add(1)
	s.mu.Unlock()

	go func() {
		defer func() {
			close(done)
			s.mu.Lock()
			if current, ok := s.monitors[srv.ID]; ok && current.done == done {
				delete(s.monitors, srv.ID)
			}
			s.mu.Unlock()
			s.wg.Done()
		}()

		log.Info().Str("server", srv.Name).Str("url", srv.URL).Int("interval", srv.CheckInterval).Msg("Monitoring started")

		var last lastResult
		last = s.performCheck(srv, last)

		ticker := time.NewTicker(time.Duration(srv.CheckInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Info().Str("server", srv.Name).Msg("Monitoring stopped")
				return
			case <-ticker.C:
				last = s.performCheck(srv, last)
			}
		}
	}()
}

// StopServerMonitor зупиняє моніторинг конкретного сервера.
func (s *Scheduler) StopServerMonitor(serverID uint) {
	s.mu.Lock()
	monitor, ok := s.monitors[serverID]
	if ok {
		delete(s.monitors, serverID)
	}
	s.mu.Unlock()

	if ok {
		monitor.cancel()
	}
}

func (s *Scheduler) performCheck(srv model.Server, last lastResult) lastResult {
	results := s.collectRegionResults(srv)
	aggregated := aggregateRegionResults(results)
	createdAt := time.Now().UTC()

	log.Debug().
		Str("server", srv.URL).
		Str("status", aggregated.Status).
		Int64("latency", aggregated.Latency).
		Int("regions", len(results)).
		Msg("Check completed")

	if err := s.saveCheckResults(srv.ID, createdAt, results); err != nil {
		log.Error().Err(err).Uint("server_id", srv.ID).Msg("Failed to save check results")
	}

	if s.notifier != nil {
		if err := s.notifier.Notify(s.ctx, srv, last.Status, aggregated.Status, aggregated.Latency); err != nil {
			log.Error().Err(err).Uint("server_id", srv.ID).Msg("Failed to process notifications")
		}
	}

	return lastResult{
		Status:  aggregated.Status,
		Latency: aggregated.Latency,
	}
}

func (s *Scheduler) collectRegionResults(srv model.Server) []RegionResult {
	localCh := make(chan RegionResult, 1)
	go func() {
		status, latency := Check(srv.CheckType, srv.URL, srv.Timeout)
		localCh <- RegionResult{
			Region:  s.localRegion,
			Status:  status,
			Latency: latency,
		}
	}()

	results := make([]RegionResult, 0, 1)
	if s.regionalChecker != nil {
		results = append(results, s.regionalChecker.CheckRegions(s.ctx, srv.CheckType, srv.URL, srv.Timeout)...)
	}
	results = append([]RegionResult{<-localCh}, results...)
	return results
}

func (s *Scheduler) saveCheckResults(serverID uint, createdAt time.Time, results []RegionResult) error {
	for _, result := range results {
		if err := s.storage.AddCheckResult(model.CheckResult{
			ServerID:  serverID,
			Region:    result.Region,
			Status:    result.Status,
			Latency:   result.Latency,
			CreatedAt: createdAt,
		}); err != nil {
			return err
		}
	}
	return nil
}

func aggregateRegionResults(results []RegionResult) RegionResult {
	if len(results) == 0 {
		return RegionResult{Region: model.CheckRegionGlobal, Status: "No regions checked"}
	}

	successes := make([]RegionResult, 0, len(results))
	for _, result := range results {
		if isSuccessfulStatus(result.Status) {
			successes = append(successes, result)
		}
	}
	if len(successes) > 0 {
		return RegionResult{
			Region:  model.CheckRegionGlobal,
			Status:  successes[0].Status,
			Latency: averageLatency(successes),
		}
	}

	return RegionResult{
		Region:  model.CheckRegionGlobal,
		Status:  formatAllRegionsFailed(results),
		Latency: maxLatency(results),
	}
}

func isSuccessfulStatus(status string) bool {
	return strings.HasPrefix(status, "2") || status == "Connected"
}

func averageLatency(results []RegionResult) int64 {
	if len(results) == 0 {
		return 0
	}
	var total int64
	for _, result := range results {
		total += result.Latency
	}
	return int64(math.Round(float64(total) / float64(len(results))))
}

func maxLatency(results []RegionResult) int64 {
	var max int64
	for _, result := range results {
		if result.Latency > max {
			max = result.Latency
		}
	}
	return max
}

func formatAllRegionsFailed(results []RegionResult) string {
	var b strings.Builder
	b.WriteString("All regions failed")
	for _, result := range results {
		fmt.Fprintf(&b, "; %s: %s", result.Region, compactStatus(result.Status))
	}
	return b.String()
}

func compactStatus(status string) string {
	status = strings.Join(strings.Fields(status), " ")
	const maxStatusLen = 160
	if len(status) <= maxStatusLen {
		return status
	}
	return status[:maxStatusLen-3] + "..."
}
