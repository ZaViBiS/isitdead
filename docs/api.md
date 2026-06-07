# Frontend API Calls

This file lists API requests that already exist in the frontend code.

It does not describe future backend routes.

## Shared Frontend Types

Source: `web/src/lib/utils.ts`

```ts
interface User {
  id: number;
  username: string;
  email: string;
  plan: string;
  stripe_subscription_status?: string;
  plan_current_period_end?: string;
}

interface BillingPlan {
  id: string;
  name: string;
  description: string;
  price: string;
  monitor_limit: number;
  min_interval: number;
  history_days: number;
  public_pages: boolean;
  ssl_monitoring: boolean;
  telegram_alerts: boolean;
  stripe_available: boolean;
}

interface Server {
  id: number;
  name: string;
  url: string;
  check_type: string;
  public: boolean;
  public_slug: string;
  check_interval: number;
  timeout: number;
  slow_threshold: number;
  ssl_enabled: boolean;
  current_status?: string;
  current_latency?: number;
  check_count_30d?: number;
  uptime_30d?: number;
  avg_latency_30d?: number;
}

interface CheckResult {
  id: number;
  region: string;
  status: string;
  latency: number;
  created_at: string;
}

interface NotificationPreference {
  id?: number;
  user_id?: number;
  server_id?: number;
  channel: string;
  event: string;
  enabled: boolean;
  destination?: string;
}
```

## Auth

### `POST /api/register`

Used in: `web/src/routes/register/+page.svelte`

Request:

```json
{
  "username": "johndoe",
  "email": "user@example.com",
  "password": "password"
}
```

Headers:

```text
Content-Type: application/json
```

Frontend expects:

```json
{
  "message": "Registration successful! Please check your email to confirm your account."
}
```

Error response used by frontend:

```json
{
  "error": "Registration failed. Please try again."
}
```

### `POST /api/login`

Used in: `web/src/routes/login/+page.svelte`

Request:

```json
{
  "email": "user@example.com",
  "password": "password"
}
```

Headers:

```text
Content-Type: application/json
```

Frontend expects:

```json
{
  "token": "jwt-or-session-token",
  "user": {
    "id": 1,
    "username": "johndoe",
    "email": "user@example.com",
    "plan": "free"
  }
}
```

Error response used by frontend:

```json
{
  "error": "Invalid email or password",
  "verification_error": true
}
```

### `POST /api/auth/session`

Used in: `web/src/routes/login/+page.svelte`

Used after Google OAuth redirects back with `?oauth=success`.

Request body: none

Frontend expects:

```json
{
  "token": "jwt-or-session-token",
  "user": {
    "id": 1,
    "username": "johndoe",
    "email": "user@example.com",
    "plan": "free"
  }
}
```

Error response used by frontend:

```json
{
  "error": "Google sign-in failed"
}
```

### `POST /api/auth/resend-confirmation`

Used in:

- `web/src/routes/register/+page.svelte`
- `web/src/routes/login/+page.svelte`

Request:

```json
{
  "email": "user@example.com"
}
```

Headers:

```text
Content-Type: application/json
```

Frontend expects:

```json
{
  "message": "Confirmation email sent. Please check your inbox."
}
```

Error response used by frontend:

```json
{
  "error": "Could not resend confirmation email"
}
```

### `GET /api/me`

Used in:

- `web/src/routes/+layout.svelte`
- `web/src/routes/pricing/+page.svelte`
- `web/src/routes/dashboard/+page.svelte`

Headers:

```text
Authorization: Bearer {token}
```

Frontend expects: `User`

```json
{
  "id": 1,
  "username": "johndoe",
  "email": "user@example.com",
  "plan": "free",
  "stripe_subscription_status": "active",
  "plan_current_period_end": "2026-06-05T12:00:00Z"
}
```

## Billing

### `GET /api/billing/plans`

Used in:

- `web/src/routes/pricing/+page.svelte`
- `web/src/routes/dashboard/+page.svelte`

Frontend expects: `BillingPlan[]`

```json
[
  {
    "id": "free",
    "name": "Free",
    "description": "Free plan",
    "price": "$0",
    "monitor_limit": 3,
    "min_interval": 300,
    "history_days": 7,
    "public_pages": true,
    "ssl_monitoring": false,
    "telegram_alerts": false,
    "stripe_available": true
  }
]
```

### `POST /api/billing/checkout`

Used in: `web/src/routes/pricing/+page.svelte`

Request:

```json
{
  "plan": "pro"
}
```

Headers:

```text
Content-Type: application/json
Authorization: Bearer {token}
```

Frontend expects:

```json
{
  "url": "https://checkout.stripe.com/..."
}
```

Error response used by frontend:

```json
{
  "error": "Could not start checkout"
}
```

### `POST /api/billing/portal`

Used in: `web/src/routes/pricing/+page.svelte`

Request body: none

Headers:

```text
Authorization: Bearer {token}
```

Frontend expects:

```json
{
  "url": "https://billing.stripe.com/..."
}
```

Error response used by frontend:

```json
{
  "error": "Could not open billing portal"
}
```

## Dashboard Servers

### `GET /api/dashboard/servers`

Used in:

- `web/src/routes/dashboard/+page.svelte`
- `web/src/routes/dashboard/[id]/+page.svelte`

Headers:

```text
Authorization: Bearer {token}
```

Frontend expects: `Server[]`

```json
[
  {
    "id": 1,
    "name": "Main Website",
    "url": "https://example.com",
    "check_type": "http",
    "public": false,
    "public_slug": "main-website",
    "check_interval": 300,
    "timeout": 10,
    "slow_threshold": 300,
    "ssl_enabled": false,
    "current_status": "200 OK",
    "current_latency": 123,
    "check_count_30d": 100,
    "uptime_30d": 99.9,
    "avg_latency_30d": 120
  }
]
```

Special frontend behavior:

- `401` removes local token and redirects to `/login`.

## Servers

### `POST /api/servers`

Used in: `web/src/routes/dashboard/+page.svelte`

Request:

```json
{
  "name": "Main Website",
  "url": "https://example.com",
  "check_type": "http",
  "check_interval": 300,
  "timeout": 10,
  "slow_threshold": 300,
  "ssl_enabled": false
}
```

Headers:

```text
Content-Type: application/json
Authorization: Bearer {token}
```

Frontend expects: `Server`

Error response used by frontend:

```json
{
  "error": "Failed to add server"
}
```

### `PUT /api/servers/{id}`

Used in: `web/src/routes/dashboard/+page.svelte`

Request:

```json
{
  "name": "Main Website",
  "url": "https://example.com",
  "check_type": "http",
  "check_interval": 300,
  "timeout": 10,
  "slow_threshold": 300,
  "ssl_enabled": false
}
```

Headers:

```text
Content-Type: application/json
Authorization: Bearer {token}
```

Frontend expects: `Server`

Error response used by frontend:

```json
{
  "error": "Failed to update server"
}
```

### `DELETE /api/servers/{id}`

Used in: `web/src/routes/dashboard/+page.svelte`

Request body: none

Headers:

```text
Authorization: Bearer {token}
```

Frontend behavior:

- If response is OK, removes the server from local dashboard state.

## Server Results

### `GET /api/servers/{id}/results?limit=1`

Used in: `web/src/routes/dashboard/+page.svelte`

Headers:

```text
Authorization: Bearer {token}
```

Frontend expects: `CheckResult[]`

Used to wait until the first check result exists after creating a server.

### `GET /api/servers/{id}/results?hours=72`

Used in: `web/src/routes/dashboard/[id]/+page.svelte`

Headers:

```text
Authorization: Bearer {token}
```

Frontend expects: `CheckResult[]`

### `GET /api/servers/{id}/results?hours=72&region=all`

Used in: `web/src/routes/dashboard/[id]/+page.svelte`

Headers:

```text
Authorization: Bearer {token}
```

Frontend expects: `CheckResult[]`

### `GET /api/servers/{id}/results?incidents=true&limit=50`

Used in: `web/src/routes/dashboard/[id]/+page.svelte`

Headers:

```text
Authorization: Bearer {token}
```

Frontend expects: `CheckResult[]`

### `GET /api/servers/{id}/results?incidents=true&limit=50&region=all`

Used in: `web/src/routes/dashboard/[id]/+page.svelte`

Headers:

```text
Authorization: Bearer {token}
```

Frontend expects: `CheckResult[]`

## Server Notifications

### `GET /api/servers/{id}/notifications`

Used in: `web/src/routes/dashboard/+page.svelte`

Headers:

```text
Authorization: Bearer {token}
```

Frontend expects: `NotificationPreference[]`

```json
[
  {
    "id": 1,
    "user_id": 1,
    "server_id": 1,
    "channel": "telegram",
    "event": "down",
    "enabled": true,
    "destination": "123456789"
  }
]
```

### `PUT /api/servers/{id}/notifications`

Used in: `web/src/routes/dashboard/+page.svelte`

Request:

```json
[
  {
    "channel": "email",
    "event": "down",
    "enabled": true
  },
  {
    "channel": "email",
    "event": "recovered",
    "enabled": true
  },
  {
    "channel": "telegram",
    "event": "down",
    "enabled": false
  },
  {
    "channel": "telegram",
    "event": "recovered",
    "enabled": false
  },
  {
    "channel": "discord",
    "event": "down",
    "enabled": false
  },
  {
    "channel": "discord",
    "event": "recovered",
    "enabled": false
  }
]
```

Headers:

```text
Content-Type: application/json
Authorization: Bearer {token}
```

Frontend behavior:

- Throws frontend error if response is not OK.

## Telegram

### `GET /api/telegram/status`

Used in: `web/src/routes/dashboard/+page.svelte`

Headers:

```text
Authorization: Bearer {token}
```

Frontend expects:

```json
{
  "linked": false,
  "link_available": true,
  "bot_name": "isitdead_bot",
  "linked_at": null
}
```

### `POST /api/telegram/link-token`

Used in: `web/src/routes/dashboard/+page.svelte`

Request body: none

Headers:

```text
Authorization: Bearer {token}
```

Frontend expects:

```json
{
  "token": "abc123",
  "url": "https://t.me/isitdead_bot?start=abc123",
  "bot_name": "isitdead_bot",
  "link_available": true
}
```

## Public Status Pages

### `GET /api/public/monitors/{slug}`

Used in: `web/src/routes/status/[slug]/+page.svelte`

Auth: none

Frontend expects public monitor data. It is used like a `Server`-style object with these fields:

```json
{
  "id": 1,
  "name": "Main Website",
  "url": "https://example.com",
  "check_type": "http",
  "check_interval": 300,
  "slow_threshold": 300,
  "current_status": "200 OK",
  "updated_at": "2026-06-05T12:00:00Z"
}
```

### `GET /api/public/monitors/{slug}/results?hours=720`

Used in: `web/src/routes/status/[slug]/+page.svelte`

Auth: none

Frontend expects: `CheckResult[]`

### `GET /api/public/monitors/{slug}/results?incidents=true&limit=50`

Used in: `web/src/routes/status/[slug]/+page.svelte`

Auth: none

Frontend expects: `CheckResult[]`
