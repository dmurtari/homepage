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
	COALESCE(date, '') AS date,
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
 	gps_lon
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
	gps_lon
) VALUES (
	$1, $2, NULLIF($3, ''), $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
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
	gps_lon = EXCLUDED.gps_lon`
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
			&p.Camera, &p.Aperture, &p.ISO, &p.Shutter, &lat, &lon,
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
		p.Camera, p.Aperture, p.ISO, p.Shutter, lat, lon,
	)

	return err
}
