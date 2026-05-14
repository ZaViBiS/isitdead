package api

import (
	"fmt"
	"html"
	"strings"
)

func (s *Server) siteNotFoundHTML(path string) string {
	homeURL := "https://isitdead.cc"
	return fmt.Sprintf(
		`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="robots" content="noindex">
  <title>Page not found - isitdead</title>
  <style>
    :root { color-scheme: dark; --bg:#182825; --text:#def4c6; --muted:rgba(222,244,198,.56); --ok:#73e2a7; --bad:#d62246; --line:rgba(222,244,198,.1); }
    * { box-sizing:border-box; }
    body { margin:0; min-height:100vh; font-family:ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background:var(--bg); color:var(--text); }
    body:before { content:""; position:fixed; inset:0; pointer-events:none; background:linear-gradient(to right, rgba(222,244,198,.045) 1px, transparent 1px), linear-gradient(to bottom, rgba(222,244,198,.045) 1px, transparent 1px); background-size:72px 72px; mask-image:linear-gradient(to bottom, black, transparent 80%%); }
    .shell { position:relative; width:min(1120px, calc(100%% - 32px)); margin:0 auto; }
    .nav { height:64px; display:flex; align-items:center; justify-content:space-between; border-bottom:1px solid var(--line); }
    .brand { display:flex; align-items:center; gap:.65rem; color:var(--text); font-weight:900; text-decoration:none; }
    .mark { width:2rem; height:2rem; border-radius:8px; border:1px solid rgba(115,226,167,.35); background:linear-gradient(135deg, rgba(115,226,167,.24), rgba(222,244,198,.06)); display:grid; place-items:center; color:var(--ok); font-weight:950; }
    main { position:relative; display:grid; min-height:calc(100vh - 64px); grid-template-columns:minmax(0,1fr) 22rem; gap:2rem; align-items:center; padding:56px 0; }
    .badge { display:inline-flex; align-items:center; gap:.5rem; border:1px solid rgba(214,34,70,.24); border-radius:999px; background:rgba(214,34,70,.1); padding:.45rem .75rem; color:var(--bad); font-size:.78rem; font-weight:900; text-transform:uppercase; }
    h1 { margin:1.2rem 0 .8rem; max-width:760px; font-size:clamp(2.6rem, 8vw, 5.8rem); line-height:.94; letter-spacing:0; }
    p { max-width:660px; color:var(--muted); font-size:1.08rem; line-height:1.75; }
    .actions { display:flex; flex-wrap:wrap; gap:.75rem; margin-top:1.8rem; }
    .button { display:inline-flex; min-height:3rem; align-items:center; justify-content:center; border-radius:8px; padding:.85rem 1.1rem; font-weight:950; text-decoration:none; }
    .button.primary { background:var(--ok); color:#182825; }
    .button.secondary { border:1px solid var(--line); background:rgba(222,244,198,.035); color:var(--text); }
    .panel { border:1px solid var(--line); border-radius:8px; background:rgba(17,31,28,.9); padding:1.25rem; }
    .label { color:var(--muted); font-size:.75rem; font-weight:900; text-transform:uppercase; }
    .path { margin-top:.7rem; overflow-wrap:anywhere; font-size:1.4rem; font-weight:950; }
    @media (max-width:820px) { main { grid-template-columns:1fr; align-items:start; } }
  </style>
</head>
<body>
  <div class="shell">
    <header class="nav">
      <a class="brand" href="%s"><span class="mark">i</span><span>isitdead</span></a>
    </header>
    <main>
      <section>
        <span class="badge">404 not found</span>
        <h1>This page is not being monitored.</h1>
        <p>The route does not exist, moved, or was never published. Return to isitdead or open your monitoring dashboard.</p>
        <div class="actions">
          <a class="button primary" href="%s">Go home</a>
          <a class="button secondary" href="%s/dashboard">Open dashboard</a>
        </div>
      </section>
      <aside class="panel">
        <div class="label">Requested path</div>
        <div class="path">%s</div>
      </aside>
    </main>
  </div>
</body>
</html>`,
		html.EscapeString(homeURL),
		html.EscapeString(homeURL),
		html.EscapeString(homeURL),
		html.EscapeString(path),
	)
}

func isKnownSPARoute(path string) bool {
	cleanPath := strings.TrimSuffix(path, "/")
	if cleanPath == "" {
		cleanPath = "/"
	}

	switch cleanPath {
	case "/", "/login", "/register", "/dashboard", "/admin/public-pages":
		return true
	}

	return strings.HasPrefix(cleanPath, "/dashboard/") || strings.HasPrefix(cleanPath, "/status/")
}
