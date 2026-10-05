#!/usr/bin/env bash
set -euo pipefail

cd /opt/homepage/homepage
run() { sudo -u deploy "$@"; }

run git pull

run go build -o app.new ./cmd/web
run go build -o workout-webhook.new ./cmd/webhooks

run mv app app.old
run mv app.new app

run mv workout-webhook workout-webhook.old
run mv workout-webhook.new workout-webhook

sudo systemctl daemon-reload
sudo systemctl restart homepage workout-webhook
sudo systemctl reload caddy
