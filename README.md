# isitdead development notes

Цей файл не є публічною презентацією проєкту. Це коротка шпаргалка для розробки, деплою і дебагу, щоб не тримати деталі в голові.

## Архітектура

- Backend: Go, entrypoint `cmd/server`.
- Frontend: SvelteKit у `web`.
- Database: SQLite, шлях задається через `DB_PATH`.
- У production frontend збирається в `web/dist` і віддається Go server-ом.
- У dev можна запускати Go API і Vite окремо; Vite proxy-ить `/api` на backend.

## Карта проєкту

```text
cmd/server/main.go
  -> internal/app/app.go
       -> internal/config/config.go
       -> internal/database/database.go
       -> internal/checker/scheduler.go
       -> internal/mail/mail.go
       -> internal/api/server.go
            -> internal/api/routes.go
            -> internal/api/auth_handlers.go
            -> internal/api/google_oauth_handlers.go
            -> internal/api/handlers.go
  -> embed.go
       -> web/dist
```

Що за що відповідає:

- `cmd/server/main.go` - мінімальний entrypoint: налаштовує logger, створює `app.App`, запускає `Run`.
- `internal/app/app.go` - composition root. Тут збираються config, database, scheduler, mailer і API server. Тут же вирішується `dev` HTTP чи `prod` HTTPS через `autocert`.
- `internal/config/config.go` - всі env defaults. Якщо щось "чомусь не підхопилось", перше місце для перевірки.
- `internal/api/server.go` - Fiber app, request logging, API routes, static serving з `web/dist`, SPA fallback.
- `internal/api/routes.go` - список HTTP endpoints. Публічні auth routes реєструються до `authMiddleware`, monitor routes після нього.
- `internal/api/auth.go` - JWT middleware і helpers для auth.
- `internal/api/auth_handlers.go` - register/login/email confirmation.
- `internal/api/google_oauth_handlers.go` - Google OAuth start/callback і linking існуючого email account.
- `internal/api/handlers.go` - CRUD моніторів і results API.
- `internal/database/*.go` - GORM/SQLite storage. `database.go` запускає один write worker через `writerChan`, конкретні файли тримають methods для users/servers/results.
- `internal/checker/checker.go` - фактичні HTTP/TCP checks.
- `internal/checker/scheduler.go` - фонові goroutines для кожного monitor-а, перший check одразу, далі ticker по `CheckInterval`.
- `internal/mail/mail.go` - verification emails and notification emails через Resend API.
- `internal/model/*.go` - GORM models і request/domain structs.
- `embed.go` - embed-ить `web/dist` у Go binary. Якщо `web/dist` нема або він старий, production/static UI теж буде старий.
- `web/src/routes/+page.svelte` - головна сторінка/preview.
- `web/src/routes/login/+page.svelte` і `register/+page.svelte` - auth UI.
- `web/src/routes/dashboard/+page.svelte` - список моніторів, add/edit/delete, URL normalization, interval controls.
- `web/src/routes/dashboard/[id]/+page.svelte` - detail сторінка монітора.
- `web/src/routes/admin/public-pages/+page.svelte` - admin UI для керування public status pages.
- `web/src/lib/*` - shared UI/helpers: logo, chart, utils.
- `web/static/favicon.svg` - favicon, який Vite копіює у build.
- `web/static/robots.txt` - robots rules для public status pages і sitemap.

Основний runtime flow:

```text
process start
  -> config.Load()
  -> database.Init(DB_PATH) + AutoMigrate + write worker
  -> checker.NewScheduler(db)
  -> mail.New(config)
  -> api.New(db, scheduler, mailer, embedded web/dist)
  -> scheduler.Start() loads monitors from DB and starts checks
  -> HTTP/HTTPS server starts
```

Flow додавання монітора:

```text
dashboard form
  -> POST /api/servers with JWT
  -> internal/api/handlers.go validates/fills defaults
  -> internal/database/server.go writes monitor
  -> scheduler.RunServerMonitor(server)
  -> first check runs immediately
  -> check result saved and server status updated
```

Flow registration:

```text
register page
  -> POST /api/register
  -> user + email verification token in SQLite
  -> mail.SendVerificationEmail()
  -> user clicks /api/auth/confirm?token=...
  -> email becomes verified
```

Flow Google login:

```text
login page
  -> GET /api/auth/google
  -> Google callback
  -> create verified user OR link existing user by email
  -> set auth cookie/JWT
  -> redirect to dashboard

Flow public status pages:

```text
admin page
  -> PUT /api/admin/servers/:id/public
  -> store public flag + public_slug in SQLite
  -> GET /status/:slug renders SEO HTML on backend
  -> GET /sitemap.xml includes public pages
  -> robots.txt allows /status/ and points to sitemap.xml
```
```

## Швидкий локальний запуск

Backend:

```sh
ENV=dev \
PORT=8080 \
DB_PATH=./isitdead.db \
JWT_SECRET=local-dev-secret \
RESEND_API_KEY= \
RESEND_FROM=no-reply@localhost \
go run ./cmd/server
```

Frontend dev server:

```sh
cd web
VITE_API_PROXY_TARGET=http://localhost:8080 npm run dev
```

Перед запуском Go binary після змін у frontend треба зібрати `web/dist`:

```sh
cd web
npm install
npm run build
cd ..
```

## Часті команди

```sh
make dev-back      # go run ./cmd/server
make dev-front     # Vite dev server
make build-front   # npm ci + npm run build у web
make build         # Go binary у ./bin/isitdead
make build-all     # frontend + backend
make test          # go test ./... -v
make fmt           # gofmt
make tidy          # go mod tidy
```

Якщо Go test впирається в cache permission, використовуй локальний cache:

```sh
GOCACHE=/tmp/go-build-cache go test ./...
```

Frontend checks:

```sh
cd web
npm run check
```

## Конфіг

Go код читає конфіг тільки зі змінних середовища. `.env` сам по собі не завантажується. Для локальної розробки використовуй shell export, `direnv`, dotenv runner або systemd `EnvironmentFile`.

| Змінна | Default | Нотатка |
| --- | --- | --- |
| `ENV` | `dev` | `dev` слухає HTTP на `PORT`; `prod` слухає HTTPS на `:443` і HTTP challenge/redirect на `:80`. |
| `PORT` | `8080` | Dev port backend-а. Також впливає на dev confirmation/OAuth URLs. |
| `DOMAIN` | `localhost` | Production домен без protocol, наприклад `isitdead.cc`. |
| `DB_PATH` | `/tmp/isitdead.db` | У production ставити в persistent директорію. |
| `JWT_SECRET` | `dev-secret-change-me` | У production обов'язково довгий random secret. |
| `RESEND_API_KEY` | empty | API key для відправки листів через Resend. |
| `RESEND_FROM` | `no-reply@localhost` | Адреса відправника з верифікованого в Resend домену; можна передати display name, наприклад `Is It Dead <no-reply@isitdead.cc>`. |
| `CLIENT_ID` | empty | Google OAuth Client ID. |
| `CLIENT_SECRET` | empty | Google OAuth Client Secret. |
| `STRIPE_SECRET_KEY` | empty | Stripe secret key для Checkout і Billing Portal. |
| `STRIPE_WEBHOOK_SECRET` | empty | Signing secret для `/api/stripe/webhook`. |
| `STRIPE_PRO_PRICE_ID` | empty | Recurring Stripe Price ID для Pro плану. |
| `STRIPE_BUSINESS_PRICE_ID` | empty | Recurring Stripe Price ID для Business плану. |

Production приклад:

```sh
ENV=prod
DOMAIN=isitdead.cc
DB_PATH=/var/lib/isitdead/isitdead.db
JWT_SECRET=replace-with-long-random-secret

RESEND_API_KEY=re_replace-with-api-key
RESEND_FROM="Is It Dead <no-reply@isitdead.cc>"

CLIENT_ID=replace-with-google-client-id
CLIENT_SECRET=replace-with-google-client-secret

STRIPE_SECRET_KEY=sk_live_replace-with-key
STRIPE_WEBHOOK_SECRET=whsec_replace-with-secret
STRIPE_PRO_PRICE_ID=price_replace-with-pro-price
STRIPE_BUSINESS_PRICE_ID=price_replace-with-business-price
```

## Пошта

Verification email відправляється при email/password registration через Resend API. Якщо Resend поверне помилку або API key не налаштований, registration поверне помилку.

Важливо:

- Домен відправника має бути верифікований у Resend.
- `RESEND_FROM` має використовувати адресу з верифікованого домену.
- Для прямого HTTP API застосунок використовує `Authorization: Bearer ...` і `User-Agent`, як вимагає Resend.
- Поточний subject verification email: `Confirm your email for isitdead.cc`.

Confirmation URL:

```text
dev:  http://localhost:PORT/api/auth/confirm?token=...
prod: https://DOMAIN/api/auth/confirm?token=...
```

## Auth і Google OAuth

Google callback URLs:

```text
dev:  http://localhost:8080/api/auth/google/callback
prod: https://isitdead.cc/api/auth/google/callback
```

Якщо dev backend не на `8080`, callback у Google Cloud Console має відповідати фактичному `PORT`.

Поведінка, яку важливо не зламати:

- Email/password registration створює user і verification token.
- Google login створює verified user, якщо такого email ще нема.
- Якщо email уже є після незавершеної email registration, Google login має прив'язати `google_id` і підтвердити email.

## Монітори

- HTTP monitor без protocol у UI нормалізується в `https://example.com`.
- Явний `http://example.com` або `https://example.com` треба залишати як є.
- Default polling interval: `5m` / `300s`.
- Якщо backend отримує некоректний interval менше `10s`, fallback має бути `300s`.

## Public pages

Public status pages не вмикаються з user dashboard. Їх вмикає тільки admin через `/admin/public-pages`.

Для цього потрібне:

- `ADMIN_EMAILS` у config;
- admin email має бути серед дозволених;
- public page має свій `public_slug`, наприклад `wikipedia-down`.

SEO/індексація:

- URL виглядає як `/status/wikipedia-down`;
- HTML рендериться backend-ом, не через client-side fetch;
- у сторінки є `title`, `meta description`, canonical, OpenGraph і JSON-LD;
- `sitemap.xml` генерується автоматично;
- `robots.txt` дозволяє `/status/` і вказує sitemap.

Корисні env для production:

```sh
ADMIN_EMAILS=your@email.com
DOMAIN=isitdead.cc
```

## Production

У `ENV=prod` застосунок сам підіймає HTTPS через Let's Encrypt `autocert`.

Перед запуском перевірити:

- `DOMAIN` вказує на VPS через A/AAAA record.
- Порти `80` і `443` відкриті.
- Процес має право слухати privileged ports або має відповідні capabilities.
- Cert cache пишеться в `./certs` відносно working directory.
- `DB_PATH` не в `/tmp`.
- `JWT_SECRET`, Resend і Google OAuth змінні задані реально, не default.

Systemd unit лежить у `isitdead.service`. Перед використанням звірити:

- `User`
- `Group`
- `WorkingDirectory`
- `ExecStart`
- `EnvironmentFile`

Типові команди:

```sh
sudo cp isitdead.service /etc/systemd/system/isitdead.service
sudo systemctl daemon-reload
sudo systemctl enable --now isitdead
sudo systemctl status isitdead
journalctl -u isitdead -f
```

## Перед комітом

Мінімально:

```sh
GOCACHE=/tmp/go-build-cache go test ./...
cd web && npm run check
```

Якщо змінював UI, перевірити dashboard на desktop і mobile. Особливо mobile layout, бо це місце вже ламалося.

Якщо змінював email/auth:

- registration через email/password;
- verification link;
- Google login для нового email;
- Google login для email, який уже є, але ще не verified;
- помилки Resend у logs.

## Нюанси, які легко забути

- Не додавати `https://`, якщо користувач явно ввів `http://`.
- `web/dist` має існувати для production/static serving.
- Frontend dev proxy працює тільки у Vite dev server, production йде через один Go server.
- `.env` не читається автоматично Go кодом.
- Gmail дуже строго перевіряє email headers, особливо `From`, `To`, `Subject`, `Date`, MIME headers.

## Discord integration

- Generate link token: `POST /api/discord/link-token` (auth required).
- Link Discord webhook: `GET /api/discord/token/:token?webhook_url=<discord_webhook_url>`.
- Check link status: `GET /api/discord/status` (auth required).
- Supported notification channel: `discord`.
