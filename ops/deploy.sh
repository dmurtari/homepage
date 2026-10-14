#!/usr/bin/env bash
set -euo pipefail

cd /opt/homepage/homepage
run() { sudo -u deploy "$@"; }

run git pull

run go build -o app.new ./cmd/web
run go build -o activities-webhook.new ./cmd/webhooks

run mv app app.old
run mv app.new app

run mv activities-webhook activities-webhook.old
run mv activities-webhook.new activities-webhook

sudo systemctl daemon-reload
sudo systemctl restart homepage activities-webhook
sudo systemctl reload caddy
