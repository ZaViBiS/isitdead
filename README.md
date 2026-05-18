X isitdead development notes

Xей файл не є публічною презентацією проєкту. Це коротка шпаргалка для розробки, деплою і дебагу, щоб не тримати деталі в голові.

X# Архітектура

X Backend: Go, entrypoint `cmd/server`.
X Frontend: SvelteKit у `web`.
X Database: SQLite, шлях задається через `DB_PATH`.
X У production frontend збирається в `web/dist` і віддається Go server-ом.
X У dev можна запускати Go API і Vite окремо; Vite proxy-ить `/api` на backend.

X# Карта проєкту

X``text
Xmd/server/main.go
X -> internal/app/app.go
X      -> internal/config/config.go
X      -> internal/database/database.go
X      -> internal/checker/scheduler.go
X      -> internal/mail/mail.go
X      -> internal/api/server.go
X           -> internal/api/routes.go
X           -> internal/api/auth_handlers.go
X           -> internal/api/google_oauth_handlers.go
X           -> internal/api/handlers.go
X -> embed.go
X      -> web/dist
X``

Xо за що відповідає:

X `cmd/server/main.go` - мінімальний entrypoint: налаштовує logger, створює `app.App`, запускає `Run`.
X `internal/app/app.go` - composition root. Тут збираються config, database, scheduler, mailer і API server. Тут же вирішується `dev` HTTP чи `prod` HTTPS через `autocert`.
X `internal/config/config.go` - всі env defaults. Якщо щось "чомусь не підхопилось", перше місце для перевірки.
X `internal/api/server.go` - Fiber app, request logging, API routes, static serving з `web/dist`, SPA fallback.
X `internal/api/routes.go` - список HTTP endpoints. Публічні auth routes реєструються до `authMiddleware`, monitor routes після нього.
X `internal/api/auth.go` - JWT middleware і helpers для auth.
X `internal/api/auth_handlers.go` - register/login/email confirmation.
X `internal/api/google_oauth_handlers.go` - Google OAuth start/callback і linking існуючого email account.
X `internal/api/handlers.go` - CRUD моніторів і results API.
X `internal/database/*.go` - GORM/SQLite storage. `database.go` запускає один write worker через `writerChan`, конкретні файли тримають methods для users/servers/results.
X `internal/checker/checker.go` - фактичні HTTP/TCP checks.
X `internal/checker/scheduler.go` - фонові goroutines для кожного monitor-а, перший check одразу, далі ticker по `CheckInterval`.
X `internal/mail/mail.go` - verification emails and notification emails через Resend API.
X `internal/model/*.go` - GORM models і request/domain structs.
X `embed.go` - embed-ить `web/dist` у Go binary. Якщо `web/dist` нема або він старий, production/static UI теж буде старий.
X `web/src/routes/+page.svelte` - головна сторінка/preview.
X `web/src/routes/login/+page.svelte` і `register/+page.svelte` - auth UI.
X `web/src/routes/dashboard/+page.svelte` - список моніторів, add/edit/delete, URL normalization, interval controls.
X `web/src/routes/dashboard/[id]/+page.svelte` - detail сторінка монітора.
X `web/src/routes/admin/public-pages/+page.svelte` - admin UI для керування public status pages.
X `web/src/lib/*` - shared UI/helpers: logo, chart, utils.
X `web/static/favicon.svg` - favicon, який Vite копіює у build.
X `web/static/robots.txt` - robots rules для public status pages і sitemap.

Xсновний runtime flow:

X``text
Xrocess start
X -> config.Load()
X -> database.Init(DB_PATH) + AutoMigrate + write worker
X -> checker.NewScheduler(db)
X -> mail.New(config)
X -> api.New(db, scheduler, mailer, embedded web/dist)
X -> scheduler.Start() loads monitors from DB and starts checks
X -> HTTP/HTTPS server starts
X``

Xlow додавання монітора:

X``text
Xashboard form
X -> POST /api/servers with JWT
X -> internal/api/handlers.go validates/fills defaults
X -> internal/database/server.go writes monitor
X -> scheduler.RunServerMonitor(server)
X -> first check runs immediately
X -> check result saved and server status updated
X``

Xlow registration:

X``text
Xegister page
X -> POST /api/register
X -> user + email verification token in SQLite
X -> mail.SendVerificationEmail()
X -> user clicks /api/auth/confirm?token=...
X -> email becomes verified
X``

Xlow Google login:

X``text
Xogin page
X -> GET /api/auth/google
X -> Google callback
X -> create verified user OR link existing user by email
X -> set auth cookie/JWT
X -> redirect to dashboard

Xlow public status pages:

X``text
Xdmin page
X -> PUT /api/admin/servers/:id/public
X -> store public flag + public_slug in SQLite
X -> GET /status/:slug renders SEO HTML on backend
X -> GET /sitemap.xml includes public pages
X -> robots.txt allows /status/ and points to sitemap.xml
X``
X``

X# Швидкий локальний запуск

Xackend:

X``sh
XNV=dev \
XORT=8080 \
XB_PATH=./isitdead.db \
XWT_SECRET=local-dev-secret \
XESEND_API_KEY= \
XESEND_FROM=no-reply@localhost \
Xo run ./cmd/server
X``

Xrontend dev server:

X``sh
Xd web
XITE_API_PROXY_TARGET=http://localhost:8080 npm run dev
X``

Xеред запуском Go binary після змін у frontend треба зібрати `web/dist`:

X``sh
Xd web
Xpm install
Xpm run build
Xd ..
X``

X# Часті команди

X``sh
Xake dev-back      # go run ./cmd/server
Xake dev-front     # Vite dev server
Xake build-front   # npm ci + npm run build у web
Xake build         # Go binary у ./bin/isitdead
Xake build-all     # frontend + backend
Xake test          # go test ./... -v
Xake fmt           # gofmt
Xake tidy          # go mod tidy
X``

Xкщо Go test впирається в cache permission, використовуй локальний cache:

X``sh
XOCACHE=/tmp/go-build-cache go test ./...
X``

Xrontend checks:

X``sh
Xd web
Xpm run check
X``

X# Конфіг

Xo код читає конфіг тільки зі змінних середовища. `.env` сам по собі не завантажується. Для локальної розробки використовуй shell export, `direnv`, dotenv runner або systemd `EnvironmentFile`.

X Змінна | Default | Нотатка |
X --- | --- | --- |
X `ENV` | `dev` | `dev` слухає HTTP на `PORT`; `prod` слухає HTTPS на `:443` і HTTP challenge/redirect на `:80`. |
X `PORT` | `8080` | Dev port backend-а. Також впливає на dev confirmation/OAuth URLs. |
X `DOMAIN` | `localhost` | Production домен без protocol, наприклад `isitdead.cc`. |
X `DB_PATH` | `/tmp/isitdead.db` | У production ставити в persistent директорію. |
X `JWT_SECRET` | `dev-secret-change-me` | У production обов'язково довгий random secret. |
X `RESEND_API_KEY` | empty | API key для відправки листів через Resend. |
X `RESEND_FROM` | `no-reply@localhost` | Адреса відправника з верифікованого в Resend домену; можна передати display name, наприклад `Is It Dead <no-reply@isitdead.cc>`. |
X `CLIENT_ID` | empty | Google OAuth Client ID. |
X `CLIENT_SECRET` | empty | Google OAuth Client Secret. |

Xroduction приклад:

X``sh
XNV=prod
XOMAIN=isitdead.cc
XB_PATH=/var/lib/isitdead/isitdead.db
XWT_SECRET=replace-with-long-random-secret

XESEND_API_KEY=re_replace-with-api-key
XESEND_FROM="Is It Dead <no-reply@isitdead.cc>"

XLIENT_ID=replace-with-google-client-id
XLIENT_SECRET=replace-with-google-client-secret
X``

X# Пошта

Xerification email відправляється при email/password registration через Resend API. Якщо Resend поверне помилку або API key не налаштований, registration поверне помилку.

Xажливо:

X Домен відправника має бути верифікований у Resend.
X `RESEND_FROM` має використовувати адресу з верифікованого домену.
X Для прямого HTTP API застосунок використовує `Authorization: Bearer ...` і `User-Agent`, як вимагає Resend.
X Поточний subject verification email: `Confirm your email for isitdead.cc`.

Xonfirmation URL:

X``text
Xev:  http://localhost:PORT/api/auth/confirm?token=...
Xrod: https://DOMAIN/api/auth/confirm?token=...
X``

X# Auth і Google OAuth

Xoogle callback URLs:

X``text
Xev:  http://localhost:8080/api/auth/google/callback
Xrod: https://isitdead.cc/api/auth/google/callback
X``

Xкщо dev backend не на `8080`, callback у Google Cloud Console має відповідати фактичному `PORT`.

Xоведінка, яку важливо не зламати:

X Email/password registration створює user і verification token.
X Google login створює verified user, якщо такого email ще нема.
X Якщо email уже є після незавершеної email registration, Google login має прив'язати `google_id` і підтвердити email.

X# Монітори

X HTTP monitor без protocol у UI нормалізується в `https://example.com`.
X Явний `http://example.com` або `https://example.com` треба залишати як є.
X Default polling interval: `5m` / `300s`.
X Якщо backend отримує некоректний interval менше `10s`, fallback має бути `300s`.

X# Public pages

Xublic status pages не вмикаються з user dashboard. Їх вмикає тільки admin через `/admin/public-pages`.

Xля цього потрібне:

X `ADMIN_EMAILS` у config;
X admin email має бути серед дозволених;
X public page має свій `public_slug`, наприклад `wikipedia-down`.

XEO/індексація:

X URL виглядає як `/status/wikipedia-down`;
X HTML рендериться backend-ом, не через client-side fetch;
X у сторінки є `title`, `meta description`, canonical, OpenGraph і JSON-LD;
X `sitemap.xml` генерується автоматично;
X `robots.txt` дозволяє `/status/` і вказує sitemap.

Xорисні env для production:

X``sh
XDMIN_EMAILS=your@email.com
XOMAIN=isitdead.cc
X``

X# Production

X `ENV=prod` застосунок сам підіймає HTTPS через Let's Encrypt `autocert`.

Xеред запуском перевірити:

X `DOMAIN` вказує на VPS через A/AAAA record.
X Порти `80` і `443` відкриті.
X Процес має право слухати privileged ports або має відповідні capabilities.
X Cert cache пишеться в `./certs` відносно working directory.
X `DB_PATH` не в `/tmp`.
X `JWT_SECRET`, Resend і Google OAuth змінні задані реально, не default.

Xystemd unit лежить у `isitdead.service`. Перед використанням звірити:

X `User`
X `Group`
X `WorkingDirectory`
X `ExecStart`
X `EnvironmentFile`

Xипові команди:

X``sh
Xudo cp isitdead.service /etc/systemd/system/isitdead.service
Xudo systemctl daemon-reload
Xudo systemctl enable --now isitdead
Xudo systemctl status isitdead
Xournalctl -u isitdead -f
X``

X# Перед комітом

Xінімально:

X``sh
XOCACHE=/tmp/go-build-cache go test ./...
Xd web && npm run check
X``

Xкщо змінював UI, перевірити dashboard на desktop і mobile. Особливо mobile layout, бо це місце вже ламалося.

Xкщо змінював email/auth:

X registration через email/password;
X verification link;
X Google login для нового email;
X Google login для email, який уже є, але ще не verified;
X помилки Resend у logs.

X# Нюанси, які легко забути

X Не додавати `https://`, якщо користувач явно ввів `http://`.
X `web/dist` має існувати для production/static serving.
X Frontend dev proxy працює тільки у Vite dev server, production йде через один Go server.
X `.env` не читається автоматично Go кодом.
X Gmail дуже строго перевіряє email headers, особливо `From`, `To`, `Subject`, `Date`, MIME headers.
