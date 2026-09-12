#!/usr/bin/env bash
set -euo pipefail

cd /opt/homepage/homepage
sudo -u deploy git pull

sudo -u deploy go build -o app.new ./cmd/web
sudo -u deploy mv app app.old
sudo -u deploy mv app.new app

sudo systemctl daemon-reload
sudo systemctl restart homepage
sudo systemctl reload caddy
