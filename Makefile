include .env

.PHONY: ingest-sync-all ingest sync-images sync-manifest sync-manifest-db

ingest-sync-all: ingest sync-images sync-manifest

ingest:
	go run ./cmd/ingest -src $(SOURCE_DIR) -out $(OUT_DIR) -db

sync-images:
	rclone sync $(OUT_DIR) $(R2_REMOTE) --exclude "manifest.json" --exclude "photos.db" \
		--header-upload "Cache-Control: public, max-age=31536000, immutable"

sync-manifest:
	rsync -avz --chmod=Du=rwx,Dgo=rx,Fu=rw,Fgo=r $(OUT_DIR)/manifest.json $(VPS_HOST):$(VPS_PATH)/manifest.json

sync-manifest-db:
	rsync -avz --chmod=Du=rwx,Dgo=rx,Fu=rw,Fgo=r $(OUT_DIR)/photos.db $(VPS_HOST):/var/lib/data/photos.db
