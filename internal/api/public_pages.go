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
	homeURL := s.publicHomeURL()
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
    .hero { display:grid; grid-template-columns:minmax(0,1fr) 22rem; gap:2rem; align-items:start; }
    .badge { display:inline-flex; gap:.5rem; align-items:center; border:1px solid var(--line); border-radius:999px; padding:.45rem .75rem; color:var(--muted); font-size:.78rem; font-weight:900; text-transform:uppercase; }
    .dot { width:.55rem; height:.55rem; border-radius:999px; background:%s; }
    h1 { margin:1.25rem 0 .75rem; max-width:820px; font-size:clamp(2.4rem, 8vw, 5.8rem); line-height:.94; letter-spacing:0; }
    .lead { max-width:760px; color:var(--muted); font-size:1.08rem; line-height:1.75; }
    .actions { display:flex; flex-wrap:wrap; gap:.75rem; margin-top:1.7rem; }
    .button { display:inline-flex; align-items:center; justify-content:center; min-height:3rem; border-radius:8px; padding:.85rem 1.1rem; font-weight:950; text-decoration:none; }
    .button.primary { background:var(--ok); color:#182825; }
    .button.secondary { border:1px solid var(--line); background:var(--panel2); color:var(--text); }
    .status-card { border:1px solid rgba(115,226,167,.22); border-radius:8px; background:rgba(17,31,28,.9); padding:1.25rem; }
    .status-card.bad { border-color:rgba(214,34,70,.28); }
    .big-status { margin-top:.65rem; color:%s; font-size:2.35rem; line-height:1; font-weight:950; }
    .endpoint { margin-top:1rem; overflow:hidden; color:var(--soft); font-size:.88rem; text-overflow:ellipsis; white-space:nowrap; }
    .grid { display:grid; grid-template-columns:repeat(3, 1fr); gap:1rem; margin:2rem 0; }
    .card { border:1px solid var(--line); border-radius:8px; background:rgba(17,31,28,.78); padding:1.25rem; }
    .label { color:var(--muted); font-size:.75rem; font-weight:900; text-transform:uppercase; }
    .value { margin-top:.55rem; font-size:2rem; font-weight:950; }
    .muted { color:var(--muted); }
    .incidents { list-style:none; padding:0; margin:0; display:grid; gap:.75rem; }
    .incidents li { display:flex; justify-content:space-between; gap:1rem; border-top:1px solid var(--line); padding-top:.75rem; color:var(--muted); }
    .seo { margin-top:1rem; color:var(--soft); line-height:1.7; }
    footer { margin-top:2rem; color:rgba(222,244,198,.38); font-size:.9rem; }
    @media (max-width: 820px) { .hero { grid-template-columns:1fr; } .grid { grid-template-columns:1fr; } .incidents li { flex-direction:column; } main { padding-top:42px; } }
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
      </aside>
    </section>
    <div class="grid">
      <section class="card"><div class="label">Current status</div><div class="value">%s</div></section>
      <section class="card"><div class="label">Response time</div><div class="value">%dms</div></section>
      <section class="card"><div class="label">Checks recorded</div><div class="value">%d</div></section>
    </div>
    <section class="card">
      <div class="label">Monitored endpoint</div>
      <p><a href="%s" rel="nofollow external">%s</a></p>
      <p class="seo">People use this page to check %s down, %s status, %s outage reports, and %s uptime monitor data without opening the service first.</p>
    </section>
    <section class="card" style="margin-top:1rem">
      <div class="label">Recent incidents</div>
      %s
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
		html.EscapeString(homeURL),
	)
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

func statusCardClass(healthy bool) string {
	if healthy {
		return ""
	}
	return " bad"
}
