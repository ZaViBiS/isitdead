
## env

TOKEN=<telegram bot token>
BASE_URL=<https://isitdead.cc> # without / at the end!
TELEGRAM_API_SECRET=<shared secret with main app>
PORT=8081

Main app:

TELEGRAM_BOT_NAME=<bot username without @>
TELEGRAM_API_URL=<http://127.0.0.1:8081>
TELEGRAM_API_SECRET=<same shared secret>

## build

From the repository root:

make build-telegram

## systemd

Install `isitdead-telegram.service` next to the main `isitdead.service`.
Create `/home/ubuntu/isitdead/.env.telegram` with the Telegram env values above.
