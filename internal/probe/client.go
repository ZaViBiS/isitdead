package probe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ZaViBiS/isitdead/internal/checker"
)

type Client struct {
	targets    []Target
	secret     string
	httpClient *http.Client
}

func NewClient(targets []Target, secret string) *Client {
	normalized := make([]Target, 0, len(targets))
	for i, target := range targets {
		region := normalizeRegionName(target.Region, fmt.Sprintf("probe-%d", i+1))
		url := strings.TrimRight(strings.TrimSpace(target.URL), "/")
		if url == "" {
			continue
		}
		normalized = append(normalized, Target{Region: region, URL: url})
	}

	return &Client{
		targets:    normalized,
		secret:     secret,
		httpClient: &http.Client{},
	}
}

func (c *Client) CheckRegions(ctx context.Context, checkType, target string, timeoutSeconds int) []checker.RegionResult {
	if len(c.targets) == 0 {
		return nil
	}

	results := make([]checker.RegionResult, len(c.targets))
	var wg sync.WaitGroup
	for i, probeTarget := range c.targets {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = c.checkRegion(ctx, probeTarget, checkType, target, timeoutSeconds)
		}()
	}
	wg.Wait()
	return results
}

func (c *Client) checkRegion(ctx context.Context, probeTarget Target, checkType, target string, timeoutSeconds int) checker.RegionResult {
	start := time.Now()
	requestTimeout := time.Duration(timeoutSeconds)*time.Second + 2*time.Second
	if timeoutSeconds <= 0 {
		requestTimeout = 12 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	payload, err := json.Marshal(CheckRequest{
		CheckType: checkType,
		URL:       target,
		Timeout:   timeoutSeconds,
	})
	if err != nil {
		return probeError(probeTarget.Region, start, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, probeTarget.URL+"/api/probe/check", bytes.NewReader(payload))
	if err != nil {
		return probeError(probeTarget.Region, start, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.secret != "" {
		req.Header.Set(SecretHeader, c.secret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return probeError(probeTarget.Region, start, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		status := strings.TrimSpace(string(body))
		if status == "" {
			status = resp.Status
		}
		return checker.RegionResult{
			Region:  probeTarget.Region,
			Status:  fmt.Sprintf("Probe HTTP error: %s", status),
			Latency: time.Since(start).Milliseconds(),
		}
	}

	var response CheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return probeError(probeTarget.Region, start, err)
	}
	if response.Status == "" {
		response.Status = response.Error
	}
	if response.Status == "" {
		response.Status = "Probe returned empty status"
	}

	region := normalizeRegionName(response.Region, probeTarget.Region)
	return checker.RegionResult{
		Region:  region,
		Status:  response.Status,
		Latency: response.Latency,
	}
}

func probeError(region string, start time.Time, err error) checker.RegionResult {
	return checker.RegionResult{
		Region:  region,
		Status:  fmt.Sprintf("Probe request error: %v", err),
		Latency: time.Since(start).Milliseconds(),
	}
}

func normalizeRegionName(region, fallback string) string {
	region = strings.TrimSpace(region)
	if region == "" || strings.EqualFold(region, "global") || strings.EqualFold(region, "all") {
		return fallback
	}
	return region
}
