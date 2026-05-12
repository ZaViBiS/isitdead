package api

import (
	"fmt"
	"html"
	"strings"

	"github.com/ZaViBiS/isitdead/internal/model"
	"github.com/gofiber/fiber/v3"
)

func (s *Server) handlePublicStatusPage(c fiber.Ctx) error {
	server, err := s.DB.GetPublicServerBySlug(c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).Type("html").SendString(publicNotFoundHTML())
	}

	results, _ := s.DB.GetHistory(server.ID)
	incidents, _ := s.DB.GetIncidents(server.ID, 10)

	c.Set("Cache-Control", "public, max-age=120")
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(s.publicStatusHTML(*server, results, incidents))
}

func (s *Server) publicStatusHTML(server model.Server, results []model.CheckResult, incidents []model.CheckResult) string {
	title := fmt.Sprintf("%s status: is %s down?", server.Name, server.Name)
	description := fmt.Sprintf("Live uptime monitor for %s. Check if %s is down, view response time, availability, outages, and recent incidents.", server.Name, server.Name)
	canonical := s.publicStatusURL(server.PublicSlug)
	healthy := isPublicHealthy(server.Status)
	statusText := "Operational"
	if !healthy {
		statusText = "Incident detected"
	}

	var incidentsHTML strings.Builder
	if len(incidents) == 0 {
		incidentsHTML.WriteString(`<p class="muted">No recent incidents reported.</p>`)
	} else {
		incidentsHTML.WriteString(`<ul class="incidents">`)
		for _, incident := range incidents {
			fmt.Fprintf(&incidentsHTML, `<li><strong>%s</strong><span>%s, %dms</span></li>`,
				html.EscapeString(incident.CreatedAt.Format("Jan 2, 2006 15:04 UTC")),
				html.EscapeString(incident.Status),
				incident.Latency,
			)
		}
		incidentsHTML.WriteString(`</ul>`)
	}

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
    :root { color-scheme: dark; --bg:#0b1412; --panel:#111f1c; --text:#def4c6; --muted:rgba(222,244,198,.55); --ok:#73e2a7; --bad:#d62246; --line:rgba(222,244,198,.1); }
    * { box-sizing: border-box; }
    body { margin:0; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: radial-gradient(circle at 20%% 0%%, rgba(115,226,167,.12), transparent 28rem), var(--bg); color:var(--text); }
    main { width:min(960px, calc(100%% - 32px)); margin:0 auto; padding:56px 0; }
    .badge { display:inline-flex; gap:.5rem; align-items:center; border:1px solid var(--line); border-radius:999px; padding:.45rem .75rem; color:var(--muted); font-size:.78rem; font-weight:800; text-transform:uppercase; }
    .dot { width:.55rem; height:.55rem; border-radius:999px; background:%s; }
    h1 { margin:1.25rem 0 .75rem; font-size:clamp(2.35rem, 6vw, 4.6rem); line-height:.95; letter-spacing:0; }
    .lead { max-width:740px; color:var(--muted); font-size:1.08rem; line-height:1.7; }
    .grid { display:grid; grid-template-columns:repeat(3, 1fr); gap:1rem; margin:2rem 0; }
    .card { border:1px solid var(--line); border-radius:1.5rem; background:rgba(17,31,28,.78); padding:1.25rem; }
    .label { color:var(--muted); font-size:.75rem; font-weight:800; text-transform:uppercase; }
    .value { margin-top:.55rem; font-size:2rem; font-weight:950; }
    .muted { color:var(--muted); }
    a { color:var(--ok); }
    .incidents { list-style:none; padding:0; margin:0; display:grid; gap:.75rem; }
    .incidents li { display:flex; justify-content:space-between; gap:1rem; border-top:1px solid var(--line); padding-top:.75rem; color:var(--muted); }
    footer { margin-top:2rem; color:rgba(222,244,198,.38); font-size:.9rem; }
    @media (max-width: 720px) { .grid { grid-template-columns:1fr; } .incidents li { flex-direction:column; } }
  </style>
</head>
<body>
  <main>
    <span class="badge"><span class="dot"></span>%s</span>
    <h1>%s status</h1>
    <p class="lead">Is %s down? This public status page tracks %s uptime, outages, response time, and recent incidents using isitdead monitoring.</p>
    <div class="grid">
      <section class="card"><div class="label">Current status</div><div class="value">%s</div></section>
      <section class="card"><div class="label">Response time</div><div class="value">%dms</div></section>
      <section class="card"><div class="label">Checks recorded</div><div class="value">%d</div></section>
    </div>
    <section class="card">
      <div class="label">Monitored endpoint</div>
      <p><a href="%s" rel="nofollow external">%s</a></p>
      <p class="muted">Keywords: %s down, %s status, %s outage, %s uptime monitor.</p>
    </section>
    <section class="card" style="margin-top:1rem">
      <div class="label">Recent incidents</div>
      %s
    </section>
    <footer>Public uptime data by isitdead.cc. Updated automatically.</footer>
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
		html.EscapeString(statusText),
		html.EscapeString(server.Name),
		html.EscapeString(server.Name),
		html.EscapeString(server.Name),
		html.EscapeString(statusText),
		server.Latency,
		len(results),
		html.EscapeString(server.URL),
		html.EscapeString(server.URL),
		html.EscapeString(server.Name),
		html.EscapeString(server.Name),
		html.EscapeString(server.Name),
		html.EscapeString(server.Name),
		incidentsHTML.String(),
	)
}

func (s *Server) publicStatusURL(slug string) string {
	if s.Config.Env == "dev" {
		return fmt.Sprintf("http://localhost:%s/status/%s", s.Config.Port, slug)
	}
	return fmt.Sprintf("https://%s/status/%s", s.Config.Domain, slug)
}

func publicNotFoundHTML() string {
	return `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="robots" content="noindex"><title>Status page not found</title></head><body><h1>Status page not found</h1></body></html>`
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
