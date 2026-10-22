package photos

import (
	"context"
	"database/sql"
)

var (
	selectAllQuery = `
SELECT
	slug,
	original,
	date,
	thumb_width,
	thumb_height,
	full_width,
	full_height,
	description,
	location,
	camera,
	aperture,
	iso,
 	shutter,
	gps_lat,
 	gps_lon,
	COALESCE(album_id, '') AS album_id
FROM photos
ORDER BY slug`

	upsertQuery = `
INSERT INTO photos (
	slug,
	original,
	date,
	thumb_width,
	thumb_height,
	full_width,
	full_height,
	description,
	location,
	camera,
	aperture,
	iso,
	shutter,
	gps_lat,
	gps_lon,
	album_id
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NULLIF($16, '')
)
ON CONFLICT (slug) DO UPDATE SET
	original = EXCLUDED.original,
	date = EXCLUDED.date,
	thumb_width = EXCLUDED.thumb_width,
	thumb_height = EXCLUDED.thumb_height,
	full_width = EXCLUDED.full_width,
	full_height = EXCLUDED.full_height,
	description = COALESCE(NULLIF(EXCLUDED.description, ''), photos.description),
	location = COALESCE(NULLIF(EXCLUDED.location, ''), photos.location),
	camera = EXCLUDED.camera,
	aperture = EXCLUDED.aperture,
	iso = EXCLUDED.iso,
	shutter = EXCLUDED.shutter,
	gps_lat = EXCLUDED.gps_lat,
	gps_lon = EXCLUDED.gps_lon,
	album_id = COALESCE(EXCLUDED.album_id, photos.album_id)`

	upsertAlbumQuery = `
INSERT INTO albums (id, name, description) VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE SET
	name = EXCLUDED.name,
	description = COALESCE(NULLIF(EXCLUDED.description, ''), albums.description)`
)

type PhotoRepository struct {
	db *sql.DB
}

func NewPhotosRepository(db *sql.DB) *PhotoRepository {
	return &PhotoRepository{db: db}
}

func (r *PhotoRepository) ListAll(ctx context.Context) ([]Photo, error) {
	rows, err := r.db.QueryContext(ctx, selectAllQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	photos := []Photo{}
	for rows.Next() {
		var p Photo
		var lat, lon sql.NullFloat64

		if err := rows.Scan(
			&p.Slug, &p.Original, &p.Date, &p.ThumbWidth, &p.ThumbHeight,
			&p.FullWidth, &p.FullHeight, &p.Description, &p.Location,
			&p.Camera, &p.Aperture, &p.ISO, &p.Shutter, &lat, &lon, &p.AlbumID,
		); err != nil {
			return nil, err
		}

		if lat.Valid && lon.Valid {
			p.GPS = []float64{lat.Float64, lon.Float64}
		}

		photos = append(photos, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return photos, nil
}

func (r *PhotoRepository) Upsert(ctx context.Context, p Photo) error {
	lat, lon := sql.NullFloat64{}, sql.NullFloat64{}
	if len(p.GPS) == 2 {
		lat = sql.NullFloat64{Float64: p.GPS[0], Valid: true}
		lon = sql.NullFloat64{Float64: p.GPS[1], Valid: true}
	}

	_, err := r.db.ExecContext(ctx, upsertQuery,
		p.Slug, p.Original, p.Date, p.ThumbWidth, p.ThumbHeight,
		p.FullWidth, p.FullHeight, p.Description, p.Location,
		p.Camera, p.Aperture, p.ISO, p.Shutter, lat, lon, p.AlbumID,
	)

	return err
}

func (r *PhotoRepository) UpsertAlbum(ctx context.Context, a Album) error {
	_, err := r.db.ExecContext(ctx, upsertAlbumQuery, a.ID, a.Name, a.Description)
	return err
}
