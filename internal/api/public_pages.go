package api

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/ZaViBiS/isitdead/internal/model"
	"github.com/gofiber/fiber/v3"
)

func (s *Server) handlePublicStatusPage(c fiber.Ctx) error {
	server, err := s.DB.GetPublicServerBySlug(c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).Type("html").SendString(s.siteNotFoundHTML(c.Path()))
	}

	results, _ := s.DB.GetHistory(server.ID)
	incidents, _ := s.DB.GetIncidents(server.ID, 10)

	c.Set("Cache-Control", "public, max-age=120")
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(s.publicStatusHTML(*server, results, incidents))
}

type publicPageMetrics struct {
	uptime           string
	averageLatency   int64
	maxLatency       int64
	incidentCount    int
	checkCount       int
	lastUpdated      string
	latencyChartHTML string
	availabilityHTML string
	recentChecksHTML string
	incidentsHTML    string
}

func (s *Server) publicStatusHTML(server model.Server, results []model.CheckResult, incidents []model.CheckResult) string {
	title := fmt.Sprintf("%s status: is %s down?", server.Name, server.Name)
	description := fmt.Sprintf("Live uptime monitor for %s. Check if %s is down, view response time, availability, outages, and recent incidents.", server.Name, server.Name)
	canonical := s.publicStatusURL(server.PublicSlug)
	homeURL := s.publicHomeURL()
	healthy := isPublicHealthy(server.Status)
	statusText := "Operational"
	if !healthy {
		statusText = "Incident detected"
	}
	metrics := buildPublicPageMetrics(server, results, incidents)

	return fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s</title>
  <meta name="description" content="%s">
  <link rel="canonical" href="%s">
  <meta property="og:title" content="%s">
  <meta property="og:description" content="%s">
  <meta property="og:type" content="website">
  <meta property="og:url" content="%s">
  <script type="application/ld+json">{"@context":"https://schema.org","@type":"WebPage","name":%q,"description":%q,"url":%q}</script>
  <style>
    :root { color-scheme: dark; --bg:#182825; --panel:#111f1c; --panel2:rgba(222,244,198,.035); --text:#def4c6; --muted:rgba(222,244,198,.56); --soft:rgba(222,244,198,.36); --ok:#73e2a7; --bad:#d62246; --warn:#e5b181; --line:rgba(222,244,198,.1); }
    * { box-sizing: border-box; }
    body { margin:0; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background:var(--bg); color:var(--text); }
    body:before { content:""; position:fixed; inset:0; pointer-events:none; background:linear-gradient(to right, rgba(222,244,198,.045) 1px, transparent 1px), linear-gradient(to bottom, rgba(222,244,198,.045) 1px, transparent 1px); background-size:72px 72px; mask-image:linear-gradient(to bottom, black, transparent 80%%); }
    a { color:inherit; }
    .shell { position:relative; width:min(1120px, calc(100%% - 32px)); margin:0 auto; }
    .nav { height:64px; display:flex; align-items:center; justify-content:space-between; border-bottom:1px solid var(--line); }
    .brand { display:flex; align-items:center; gap:.65rem; font-weight:900; text-decoration:none; }
    .mark { width:2rem; height:2rem; border-radius:8px; border:1px solid rgba(115,226,167,.35); background:linear-gradient(135deg, rgba(115,226,167,.24), rgba(222,244,198,.06)); display:grid; place-items:center; color:var(--ok); font-weight:950; }
    .navlink { color:var(--muted); font-size:.9rem; font-weight:800; text-decoration:none; }
    main { padding:64px 0 40px; }
    .hero { display:grid; grid-template-columns:minmax(0,1fr) 24rem; gap:2rem; align-items:start; }
    .badge { display:inline-flex; gap:.5rem; align-items:center; border:1px solid var(--line); border-radius:999px; padding:.45rem .75rem; color:var(--muted); font-size:.78rem; font-weight:900; text-transform:uppercase; }
    .dot { width:.55rem; height:.55rem; border-radius:999px; background:%s; }
    h1 { margin:1.25rem 0 .75rem; max-width:820px; font-size:clamp(2.4rem, 8vw, 5.8rem); line-height:.94; letter-spacing:0; }
    .lead { max-width:760px; color:var(--muted); font-size:1.08rem; line-height:1.75; }
    .actions { display:flex; flex-wrap:wrap; gap:.75rem; margin-top:1.7rem; }
    .button { display:inline-flex; align-items:center; justify-content:center; min-height:3rem; border-radius:8px; padding:.85rem 1.1rem; font-weight:950; text-decoration:none; }
    .button.primary { background:var(--ok); color:#182825; }
    .button.secondary { border:1px solid var(--line); background:var(--panel2); color:var(--text); }
    .status-card { border:1px solid rgba(115,226,167,.22); border-radius:8px; background:linear-gradient(180deg, rgba(17,31,28,.96), rgba(17,31,28,.76)); padding:1.25rem; box-shadow:0 24px 80px rgba(0,0,0,.22); }
    .status-card.bad { border-color:rgba(214,34,70,.28); }
    .big-status { margin-top:.65rem; color:%s; font-size:2.35rem; line-height:1; font-weight:950; }
    .endpoint { margin-top:1rem; overflow:hidden; color:var(--soft); font-size:.88rem; text-overflow:ellipsis; white-space:nowrap; }
    .grid { display:grid; grid-template-columns:repeat(3, 1fr); gap:1rem; margin:2rem 0; }
    .card { border:1px solid var(--line); border-radius:8px; background:rgba(17,31,28,.78); padding:1.25rem; }
    .showcase { display:grid; grid-template-columns:minmax(0,1.35fr) minmax(18rem,.65fr); gap:1rem; margin:1rem 0; }
    .chart-card { min-height:25rem; }
    .chart-head { display:flex; align-items:flex-start; justify-content:space-between; gap:1rem; margin-bottom:1.25rem; }
    .chart-title { margin:.35rem 0 0; font-size:1.8rem; line-height:1.05; font-weight:950; }
    .chart-meta { color:var(--soft); font-size:.88rem; text-align:right; }
    .latency-chart { display:block; width:100%%; height:auto; overflow:visible; }
    .axis { stroke:rgba(222,244,198,.12); stroke-width:1; }
    .area { fill:rgba(115,226,167,.12); }
    .line { fill:none; stroke:var(--ok); stroke-width:4; stroke-linecap:round; stroke-linejoin:round; }
    .point { fill:var(--ok); stroke:#182825; stroke-width:3; }
    .bad-point { fill:var(--bad); }
    .empty-chart { display:grid; min-height:15rem; place-items:center; border:1px dashed var(--line); border-radius:8px; color:var(--muted); }
    .availability { display:grid; grid-template-columns:repeat(30, minmax(0, 1fr)); gap:.25rem; margin-top:1.1rem; }
    .tick { height:1.45rem; border-radius:4px; background:rgba(222,244,198,.1); }
    .tick.ok { background:var(--ok); }
    .tick.bad { background:var(--bad); }
    .tick.unknown { background:rgba(222,244,198,.1); }
    .recent { display:grid; gap:.65rem; margin-top:1rem; }
    .check-row { display:grid; grid-template-columns:auto minmax(0,1fr) auto; gap:.75rem; align-items:center; border-top:1px solid var(--line); padding-top:.65rem; }
    .check-dot { width:.65rem; height:.65rem; border-radius:999px; background:var(--ok); }
    .check-dot.bad { background:var(--bad); }
    .check-status { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; font-weight:850; }
    .check-time { color:var(--soft); font-size:.82rem; }
    .status-meta { display:grid; grid-template-columns:repeat(2, minmax(0,1fr)); gap:.75rem; margin-top:1rem; }
    .meta-box { border:1px solid var(--line); border-radius:8px; background:rgba(222,244,198,.03); padding:.75rem; }
    .meta-value { margin-top:.35rem; font-weight:950; }
    .label { color:var(--muted); font-size:.75rem; font-weight:900; text-transform:uppercase; }
    .value { margin-top:.55rem; font-size:2rem; font-weight:950; }
    .muted { color:var(--muted); }
    .incidents { list-style:none; padding:0; margin:0; display:grid; gap:.75rem; }
    .incidents li { display:flex; justify-content:space-between; gap:1rem; border-top:1px solid var(--line); padding-top:.75rem; color:var(--muted); }
    .seo { margin-top:1rem; color:var(--soft); line-height:1.7; }
    .product-hook { display:grid; grid-template-columns:minmax(0,1fr) auto; gap:1rem; align-items:center; margin-top:1rem; border-color:rgba(115,226,167,.2); background:linear-gradient(135deg, rgba(115,226,167,.13), rgba(227,192,211,.06)); }
    .product-hook h2 { margin:0; font-size:1.9rem; line-height:1.1; }
    .product-hook p { margin:.5rem 0 0; color:var(--muted); line-height:1.65; }
    footer { margin-top:2rem; color:rgba(222,244,198,.38); font-size:.9rem; }
    @media (max-width: 900px) { .hero,.showcase,.product-hook { grid-template-columns:1fr; } .grid { grid-template-columns:1fr; } .incidents li { flex-direction:column; } .chart-meta { text-align:left; } main { padding-top:42px; } }
    @media (max-width: 540px) { .availability { grid-template-columns:repeat(15, minmax(0, 1fr)); } .chart-head { display:block; } }
  </style>
</head>
<body>
  <div class="shell">
    <header class="nav">
      <a class="brand" href="%s"><span class="mark">i</span><span>isitdead</span></a>
      <a class="navlink" href="%s">Create monitor</a>
    </header>
  </div>
  <main class="shell">
    <section class="hero">
      <div>
        <span class="badge"><span class="dot"></span>%s</span>
        <h1>%s status</h1>
        <p class="lead">Is %s down right now? This public status page tracks %s uptime, outages, response time, and recent incidents using isitdead monitoring.</p>
        <div class="actions">
          <a class="button primary" href="%s" rel="nofollow external">Open %s</a>
          <a class="button secondary" href="%s">Monitor your own service</a>
        </div>
      </div>
      <aside class="status-card%s">
        <div class="label">Current status</div>
        <div class="big-status">%s</div>
        <div class="endpoint">%s</div>
        <div class="status-meta">
          <div class="meta-box"><div class="label">Updated</div><div class="meta-value">%s</div></div>
          <div class="meta-box"><div class="label">Incidents</div><div class="meta-value">%d</div></div>
        </div>
        <div class="recent">%s</div>
      </aside>
    </section>
    <div class="grid">
      <section class="card"><div class="label">Current status</div><div class="value">%s</div></section>
      <section class="card"><div class="label">Uptime</div><div class="value">%s</div></section>
      <section class="card"><div class="label">Average latency</div><div class="value">%dms</div></section>
    </div>
    <section class="showcase">
      <div class="card chart-card">
        <div class="chart-head">
          <div>
            <div class="label">Response time</div>
            <h2 class="chart-title">Latency trend</h2>
          </div>
          <div class="chart-meta">Max %dms<br>%d checks tracked</div>
        </div>
        %s
      </div>
      <div class="card">
        <div class="label">Availability</div>
        <div class="value">%s</div>
        <p class="muted">Recent successful checks across this public monitor.</p>
        %s
      </div>
    </section>
    <section class="card">
      <div class="label">Monitored endpoint</div>
      <p><a href="%s" rel="nofollow external">%s</a></p>
      <p class="seo">People use this page to check %s down, %s status, %s outage reports, and %s uptime monitor data without opening the service first.</p>
    </section>
    <section class="card" style="margin-top:1rem">
      <div class="label">Recent incidents</div>
      %s
    </section>
    <section class="card product-hook">
      <div>
        <h2>Want a status page like this for your service?</h2>
        <p>isitdead turns uptime checks, response time, incidents, and public status pages into a dashboard that is easy to share.</p>
      </div>
      <a class="button primary" href="%s">Start monitoring</a>
    </section>
    <footer>Public uptime data by <a href="%s">isitdead.cc</a>. Updated automatically.</footer>
  </main>
</body>
</html>`,
		html.EscapeString(title),
		html.EscapeString(description),
		html.EscapeString(canonical),
		html.EscapeString(title),
		html.EscapeString(description),
		html.EscapeString(canonical),
		title,
		description,
		canonical,
		statusColor(healthy),
		statusColor(healthy),
		html.EscapeString(homeURL),
		html.EscapeString(homeURL+"/register"),
		html.EscapeString(statusText),
		html.EscapeString(server.Name),
		html.EscapeString(server.Name),
		html.EscapeString(server.Name),
		html.EscapeString(server.URL),
		html.EscapeString(server.Name),
		html.EscapeString(homeURL+"/register"),
		statusCardClass(healthy),
		html.EscapeString(statusText),
		html.EscapeString(server.URL),
		html.EscapeString(metrics.lastUpdated),
		metrics.incidentCount,
		metrics.recentChecksHTML,
		html.EscapeString(statusText),
		metrics.uptime,
		metrics.averageLatency,
		metrics.maxLatency,
		metrics.checkCount,
		metrics.latencyChartHTML,
		metrics.uptime,
		metrics.availabilityHTML,
		html.EscapeString(server.URL),
		html.EscapeString(server.URL),
		html.EscapeString(server.Name),
		html.EscapeString(server.Name),
		html.EscapeString(server.Name),
		html.EscapeString(server.Name),
		metrics.incidentsHTML,
		html.EscapeString(homeURL+"/register"),
		html.EscapeString(homeURL),
	)
}

func buildPublicPageMetrics(server model.Server, results []model.CheckResult, incidents []model.CheckResult) publicPageMetrics {
	checkCount := len(results)
	healthyCount := 0
	var latencyTotal int64
	var maxLatency int64

	for _, result := range results {
		if isPublicHealthy(result.Status) {
			healthyCount++
		}
		latencyTotal += result.Latency
		if result.Latency > maxLatency {
			maxLatency = result.Latency
		}
	}

	averageLatency := server.Latency
	if checkCount > 0 {
		averageLatency = latencyTotal / int64(checkCount)
	}

	uptime := "No data"
	if checkCount > 0 {
		uptime = fmt.Sprintf("%.2f%%", float64(healthyCount)/float64(checkCount)*100)
	}

	lastUpdated := "Waiting for first check"
	if server.LastCheck != nil {
		lastUpdated = formatPublicTime(*server.LastCheck)
	} else if checkCount > 0 {
		lastUpdated = formatPublicTime(results[checkCount-1].CreatedAt)
	}

	return publicPageMetrics{
		uptime:           uptime,
		averageLatency:   averageLatency,
		maxLatency:       maxLatency,
		incidentCount:    len(incidents),
		checkCount:       checkCount,
		lastUpdated:      lastUpdated,
		latencyChartHTML: renderLatencyChart(results),
		availabilityHTML: renderAvailabilityGrid(results),
		recentChecksHTML: renderRecentChecks(results, lastUpdated),
		incidentsHTML:    renderPublicIncidents(incidents),
	}
}

func renderLatencyChart(results []model.CheckResult) string {
	points := lastResults(results, 36)
	if len(points) == 0 {
		return `<div class="empty-chart">Waiting for monitor data</div>`
	}

	var maxLatency int64 = 1
	for _, result := range points {
		if result.Latency > maxLatency {
			maxLatency = result.Latency
		}
	}

	const width = 640.0
	const height = 220.0
	const paddingX = 22.0
	const paddingTop = 18.0
	const paddingBottom = 32.0
	plotWidth := width - paddingX*2
	plotHeight := height - paddingTop - paddingBottom

	coords := make([]string, 0, len(points))
	area := make([]string, 0, len(points)+2)
	circles := strings.Builder{}
	for i, result := range points {
		x := paddingX
		if len(points) > 1 {
			x += float64(i) / float64(len(points)-1) * plotWidth
		}
		y := paddingTop + (1-float64(result.Latency)/float64(maxLatency))*plotHeight
		coord := fmt.Sprintf("%.1f,%.1f", x, y)
		coords = append(coords, coord)
		area = append(area, coord)

		if i == 0 || i == len(points)-1 || !isPublicHealthy(result.Status) {
			className := "point"
			if !isPublicHealthy(result.Status) {
				className = "point bad-point"
			}
			fmt.Fprintf(&circles, `<circle class="%s" cx="%.1f" cy="%.1f" r="5"><title>%s, %dms</title></circle>`,
				className,
				x,
				y,
				html.EscapeString(formatPublicTime(result.CreatedAt)),
				result.Latency,
			)
		}
	}

	baseline := height - paddingBottom
	area = append([]string{fmt.Sprintf("%.1f,%.1f", paddingX, baseline)}, area...)
	area = append(area, fmt.Sprintf("%.1f,%.1f", paddingX+plotWidth, baseline))

	return fmt.Sprintf(`<svg class="latency-chart" viewBox="0 0 %.0f %.0f" role="img" aria-label="Latency trend chart">
  <line class="axis" x1="22" y1="188" x2="618" y2="188"></line>
  <line class="axis" x1="22" y1="103" x2="618" y2="103"></line>
  <polygon class="area" points="%s"></polygon>
  <polyline class="line" points="%s"></polyline>
  %s
</svg>`, width, height, strings.Join(area, " "), strings.Join(coords, " "), circles.String())
}

func renderAvailabilityGrid(results []model.CheckResult) string {
	points := lastResults(results, 60)
	if len(points) == 0 {
		return `<div class="availability"><span class="tick unknown" title="No data"></span></div>`
	}

	var b strings.Builder
	b.WriteString(`<div class="availability" aria-label="Recent availability checks">`)
	for _, result := range points {
		className := "ok"
		if !isPublicHealthy(result.Status) {
			className = "bad"
		}
		fmt.Fprintf(&b, `<span class="tick %s" title="%s, %s, %dms"></span>`,
			className,
			html.EscapeString(formatPublicTime(result.CreatedAt)),
			html.EscapeString(result.Status),
			result.Latency,
		)
	}
	b.WriteString(`</div>`)
	return b.String()
}

func renderRecentChecks(results []model.CheckResult, fallback string) string {
	points := lastResults(results, 5)
	if len(points) == 0 {
		return fmt.Sprintf(`<div class="check-row"><span class="check-dot"></span><span class="check-status">Waiting for first check</span><span class="check-time">%s</span></div>`, html.EscapeString(fallback))
	}

	var b strings.Builder
	for i := len(points) - 1; i >= 0; i-- {
		result := points[i]
		dotClass := "check-dot"
		if !isPublicHealthy(result.Status) {
			dotClass += " bad"
		}
		fmt.Fprintf(&b, `<div class="check-row"><span class="%s"></span><span class="check-status">%s</span><span class="check-time">%dms</span></div>`,
			dotClass,
			html.EscapeString(result.Status),
			result.Latency,
		)
	}
	return b.String()
}

func renderPublicIncidents(incidents []model.CheckResult) string {
	if len(incidents) == 0 {
		return `<p class="muted">No recent incidents reported.</p>`
	}

	var b strings.Builder
	b.WriteString(`<ul class="incidents">`)
	for _, incident := range incidents {
		fmt.Fprintf(&b, `<li><strong>%s</strong><span>%s, %dms</span></li>`,
			html.EscapeString(formatPublicTime(incident.CreatedAt)),
			html.EscapeString(incident.Status),
			incident.Latency,
		)
	}
	b.WriteString(`</ul>`)
	return b.String()
}

func lastResults(results []model.CheckResult, limit int) []model.CheckResult {
	if len(results) <= limit {
		return results
	}
	return results[len(results)-limit:]
}

func formatPublicTime(t time.Time) string {
	return t.UTC().Format("Jan 2, 2006 15:04 UTC")
}

func (s *Server) publicStatusURL(slug string) string {
	if s.Config.Env == "dev" {
		return fmt.Sprintf("http://localhost:%s/status/%s", s.Config.Port, slug)
	}
	return fmt.Sprintf("https://%s/status/%s", s.Config.Domain, slug)
}

func (s *Server) publicHomeURL() string {
	if s.Config.Env == "dev" {
		return fmt.Sprintf("http://localhost:%s", s.Config.Port)
	}
	return fmt.Sprintf("https://%s", s.Config.Domain)
}

func isPublicHealthy(status string) bool {
	return strings.HasPrefix(status, "2") || status == "Connected"
}

func statusColor(healthy bool) string {
	if healthy {
		return "var(--ok)"
	}
	return "var(--bad)"
}

func statusCardClass(healthy bool) string {
	if healthy {
		return ""
	}
	return " bad"
}
