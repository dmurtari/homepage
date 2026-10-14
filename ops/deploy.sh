#!/usr/bin/env bash
set -euo pipefail

cd /opt/homepage/homepage
run() { sudo -u deploy "$@"; }

run git pull

run go build -o app.new ./cmd/web
run go build -o webhooks.new ./cmd/webhooks

run mv app app.old
run mv app.new app

run mv webhooks webhooks.old
run mv webhooks.new webhooks

sudo systemctl daemon-reload
sudo systemctl restart homepage webhooks
sudo systemctl reload caddy
