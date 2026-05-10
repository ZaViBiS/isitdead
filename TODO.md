# TODO проєкту: SaaS для моніторингу серверів

## Етап 1 — MVP

- [x] [mvp] Реалізувати GORM модель `Server`: Name, URL, Status, Latency, CheckInterval, LastCheck (редагувати `internal/database/model/server.go`)
- [x] [mvp] Реалізувати GORM модель `CheckResult`: ServerID, Status, Latency, CreatedAt (створити `internal/database/model/check_result.go`)
- [x] [mvp] Реалізувати механізм запису в БД через канал та воркер (реалізовано в `internal/database/database.go`)
- [ ] [mvp] Реалізувати логіку автоочищення результатів перевірки, старіших за 30 днів (редагувати `internal/database/database.go`)
- [ ] [mvp] Реалізувати CRUD API ендпоінти: POST/GET/DELETE для серверів (редагувати `internal/api/api.go`)
- [x] [mvp] Реалізувати HTTP/TCP чекер: ping URL, замірювання затримки, визначення статусу (створити `internal/checker/checker.go`)
- [x] [mvp] Реалізувати планувальник (scheduler) для запуску перевірок згідно з `CheckInterval` (створити `internal/checker/scheduler.go`)
- [x] [mvp] Створити дашборд на Svelte з таблицею відображення статусу (зелений/червоний) (редагувати `web/src/routes/+page.svelte`)
- [x] [mvp] Застосувати колірну схему (#182825, #73E2A7, #DEF4C6, #D62246, #E3C0D3) (редагувати `web/src/routes/layout.css`)
- [x] [mvp] Реалізувати збереження історії перевірок (редагувати `internal/database/database.go`)
- [ ] [mvp] Реалізувати автоматичне резервне копіювання SQLite бази даних (через Go routine) (редагувати `internal/database/database.go`)
- [ ] [mvp] Налаштувати розгортання на VPS: створити unit-файл для systemd та конфігурацію nginx reverse proxy (створити `deploy/`)

## Етап 2 — MVP+

- [ ] [mvp+] Реалізувати Telegram-бота для сповіщень (сервер впав/відновився) (створити `internal/checker/notifications/telegram.go`)
- [ ] [mvp+] Покращити чекер: перевірка HTTP статус-кодів та виявлення 4xx/5xx відповідей (редагувати `internal/checker/checker.go`)
- [ ] [mvp+] Реалізувати OAuth-авторизацію через `goth` (Google, GitHub, GitLab) (редагувати `internal/api/auth.go`)
- [ ] [mvp+] Реалізувати реєстрацію через email/пароль з підтвердженням пошти (редагувати `internal/api/auth.go`)
- [ ] [mvp+] Створити сторінки "Умови використання" та "Політика конфіденційності" (створити `web/src/routes/legal/+page.svelte`)
- [ ] [mvp+] Реалізувати ізоляцію даних користувачів (кожен бачить лише свої сервери) (редагувати `internal/api/api.go`)
- [ ] [mvp+] Реалізувати логіку підписок: обмеження кількості серверів згідно з тарифним планом (редагувати `internal/database/account.go`)
- [ ] [mvp+] Інтегрувати вебхуки Discord та Slack для сповіщень (створити `internal/checker/notifications/webhook.go`)
- [ ] [mvp+] Додати підтримку Cron для зовнішнього запуску перевірок (редагувати `internal/api/api.go`)

## Етап 3 — Реліз

- [ ] [release] Інтегрувати Stripe для платежів (створити `internal/billing/stripe.go`)
- [ ] [release] Реалізувати моніторинг DNS-записів та оповіщення (створити `internal/checker/dns.go`)
- [ ] [release] Реалізувати моніторинг доступності окремих API-ендпоінтів (редагувати `internal/checker/checker.go`)
- [ ] [release] Реалізувати валідатор API-відповідей (перевірка вмісту JSON/body) (редагувати `internal/checker/checker.go`)
- [ ] [release] Реалізувати моніторинг "повільних" відповідей з оповіщенням при перевищенні порогу (редагувати `internal/checker/checker.go`)
- [ ] [release] SEO-оптимізація: генерація sitemap.xml та додавання мета-тегів (редагувати `web/src/app.html`)
- [ ] [release] Створити публічну сторінку статусу (створити `web/src/routes/status/[id]/+page.svelte`)
- [ ] [release] Просування: додавання на Product Hunt та каталоги SaaS-сервісів (alternativeto.net, saashub.com, toolify.ai)
