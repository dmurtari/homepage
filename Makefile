SOURCE_DIR := ~/Workspace/source
OUT_DIR := ./tmp/out
VPS_HOST := domenic@hetzner
VPS_PATH := /var/www/photos
R2_REMOTE := r2:murtari-photos

.PHONY: ingest-sync-all ingest sync-images sync-manifest

ingest-sync-all: ingest sync-images sync-manifest

ingest:
	go run ./cmd/ingest -src $(SOURCE_DIR) -out $(OUT_DIR)

sync-images:
	rclone sync $(OUT_DIR) $(R2_REMOTE) --exclude "manifest.json" \
		--header-upload "Cache-Control: public, max-age=31536000, immutable"

sync-manifest:
	rsync -avz --chmod=Du=rwx,Dgo=rx,Fu=rw,Fgo=r $(OUT_DIR)/manifest.json $(VPS_HOST):$(VPS_PATH)/manifest.json
