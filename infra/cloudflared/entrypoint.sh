#!/bin/bash
set -euo pipefail

: "${TARGET:=http://backend:8080}"
: "${TELEGRAM_BOT_TOKEN:?TELEGRAM_BOT_TOKEN must be set}"
WEBHOOK_PATH="/api/v1/integrations/telegram/webhook"
LOG=/tmp/cloudflared.log
: > "$LOG"

echo "[tunnel] starting cloudflared --url $TARGET"
cloudflared tunnel --no-autoupdate --url "$TARGET" 2>&1 | tee -a "$LOG" &
CF_PID=$!

(
  LAST=""
  # cloudflared prints a banner with the trycloudflare URL once the tunnel is up
  while kill -0 "$CF_PID" 2>/dev/null; do
    URL=$(grep -oE 'https://[a-z0-9-]+\.trycloudflare\.com' "$LOG" | tail -1 || true)
    if [[ -n "$URL" && "$URL" != "$LAST" ]]; then
      FULL="${URL}${WEBHOOK_PATH}"
      echo "[updater] tunnel URL: $URL — registering webhook"
      RESP=$(curl -sS --max-time 15 -F "url=$FULL" \
        "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/setWebhook" || true)
      echo "[updater] setWebhook response: $RESP"
      LAST="$URL"
    fi
    sleep 5
  done
) &
UPDATER_PID=$!

trap 'kill $CF_PID $UPDATER_PID 2>/dev/null || true' TERM INT
wait "$CF_PID"
