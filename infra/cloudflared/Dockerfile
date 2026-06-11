# Quick Cloudflare Tunnel + auto-updater for the Telegram webhook.
# Single container: runs `cloudflared tunnel --url http://backend:8080` and,
# whenever the assigned trycloudflare.com hostname changes, re-registers it as
# the Telegram bot webhook URL. Restart-safe: on every container start a fresh
# tunnel URL is registered automatically.
FROM alpine:3.20

RUN apk add --no-cache curl bash ca-certificates tini \
 && curl -fsSL -o /usr/local/bin/cloudflared \
      https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 \
 && chmod +x /usr/local/bin/cloudflared

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/sbin/tini", "--", "/entrypoint.sh"]
