# isitdead

Self-hosted uptime monitor на Go + SvelteKit. Сервіс перевіряє HTTP endpoints і TCP порти, зберігає історію в SQLite та віддає dashboard зі статичних файлів, вбудованих у Go binary.

## Можливості

- HTTP GET checks і TCP Ping checks.
- Історія перевірок, uptime, latency та список інцидентів.
- Реєстрація з email verification.
- Login/password auth і Google OAuth.
- SQLite без окремого database server.
- Один production binary після збірки frontend.

## Вимоги

- Go `1.26+`, відповідно до `go.mod`.
- Node.js і npm, сумісні з поточним SvelteKit/Vite стеком у `web/package.json`.
- SMTP акаунт для реєстрації користувачів через email verification.
- Домен з A/AAAA записом на сервер, якщо запускаєш `ENV=prod`.

## Локальний запуск

Збери frontend. Це потрібно, бо Go binary embed-ить `web/dist`.

```sh
cd web
npm install
npm run build
cd ..
```

Запусти backend:

```sh
ENV=dev \
PORT=8080 \
DB_PATH=./isitdead.db \
JWT_SECRET=local-dev-secret \
SMTP_HOST=localhost \
SMTP_PORT=1025 \
SMTP_USER= \
SMTP_PASS= \
SMTP_FROM=no-reply@localhost \
go run ./cmd/server
```

Для frontend dev server:

```sh
cd web
VITE_API_PROXY_TARGET=http://localhost:8080 npm run dev
```

`VITE_API_PROXY_TARGET` використовується тільки у Vite dev server для proxy `/api` на Go backend. У production frontend і API працюють з одного Go server, тому окремий proxy не потрібен.

## Збірка

Повна збірка frontend + backend:

```sh
make build-all
```

Binary буде тут:

```sh
./bin/isitdead
```

Запуск з уже зібраним binary:

```sh
ENV=dev PORT=8080 DB_PATH=./isitdead.db JWT_SECRET=local-dev-secret ./bin/isitdead
```

## Конфігурація

Застосунок читає конфігурацію зі змінних середовища. Файл `.env` не завантажується автоматично самим Go кодом. Для локального запуску експортуй змінні shell-ом, використовуй `direnv`/dotenv runner або запускай через systemd `EnvironmentFile`.

| Змінна | Default | Опис |
| --- | --- | --- |
| `ENV` | `dev` | `dev` запускає HTTP server на `PORT`. `prod` запускає HTTPS на `:443` через Let's Encrypt і HTTP challenge на `:80`. |
| `PORT` | `8080` | Порт для `ENV=dev`. Також використовується у dev email/OAuth callback URL. |
| `DOMAIN` | `localhost` | Production домен без `http://` або `https://`, наприклад `status.example.com`. |
| `DB_PATH` | `/tmp/isitdead.db` | Шлях до SQLite database file. У production краще використовувати persistent директорію. |
| `JWT_SECRET` | `dev-secret-change-me` | Secret для підпису JWT. У production обов'язково задай довге випадкове значення. |
| `SMTP_HOST` | `localhost` | SMTP host для листів підтвердження email. |
| `SMTP_PORT` | `587` | SMTP port. |
| `SMTP_USER` | empty | SMTP username. |
| `SMTP_PASS` | empty | SMTP password. |
| `SMTP_FROM` | `no-reply@localhost` | From address у листах підтвердження. |
| `CLIENT_ID` | empty | Google OAuth Client ID. Потрібен тільки для Google login. |
| `CLIENT_SECRET` | empty | Google OAuth Client Secret. Потрібен тільки для Google login. |

Приклад `.env` для production:

```sh
ENV=prod
DOMAIN=status.example.com
DB_PATH=/var/lib/isitdead/isitdead.db
JWT_SECRET=replace-with-a-long-random-secret

SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=postmaster@example.com
SMTP_PASS=replace-with-smtp-password
SMTP_FROM=no-reply@example.com

CLIENT_ID=replace-with-google-client-id
CLIENT_SECRET=replace-with-google-client-secret
```

## Email verification

Реєстрація користувача викликає `SMTP_*` конфігурацію і надсилає confirmation link.

Для `ENV=dev` confirmation URL буде:

```text
http://localhost:PORT/api/auth/confirm?token=...
```

Для `ENV=prod` confirmation URL буде:

```text
https://DOMAIN/api/auth/confirm?token=...
```

Якщо SMTP не налаштований або недоступний, registration поверне помилку і користувач не зможе підтвердити email.

## Google OAuth

У Google Cloud Console створи OAuth Client типу `Web application` і додай callback URL.

Для local dev з default port:

```text
http://localhost:8080/api/auth/google/callback
```

Для production:

```text
https://status.example.com/api/auth/google/callback
```

Після цього задай `CLIENT_ID` і `CLIENT_SECRET`. Якщо запускаєш dev backend не на `8080`, callback URL має використовувати твій `PORT`.

## Production notes

У `ENV=prod` застосунок слухає `:443` і використовує Let's Encrypt через `autocert`. Для цього:

- `DOMAIN` має вказувати на сервер.
- Порти `80` і `443` мають бути відкриті.
- Процес має право слухати privileged ports або запускатися за reverse proxy/з потрібними capabilities.
- Сертифікати кешуються в директорії `./certs` відносно working directory.

Systemd unit є в `isitdead.service`. Перед використанням зміни `User`, `Group`, `WorkingDirectory`, `ExecStart` і `EnvironmentFile` під свій сервер.

```sh
sudo cp isitdead.service /etc/systemd/system/isitdead.service
sudo systemctl daemon-reload
sudo systemctl enable --now isitdead
sudo systemctl status isitdead
```

## Команди розробки

```sh
make build-front   # зібрати SvelteKit frontend
make build         # зібрати Go binary
make build-all     # frontend + backend
make test          # Go tests
make dev-back      # go run ./cmd/server
make dev-front     # Vite dev server
```
