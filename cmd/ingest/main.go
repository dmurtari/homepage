package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/rwcarlsen/goexif/exif"
	"murtari.me/web/internal/database"
	"murtari.me/web/internal/photos"
)

const (
	thumbSize   = 400
	fullSize    = 2000
	jpegQuality = 90
)

func main() {
	srcDir := flag.String("src", "", "Source images")
	outDir := flag.String("out", "", "Output images")

	flag.Parse()

	if err := run(*srcDir, *outDir); err != nil {
		log.Fatal(err)
	}
}

func run(srcDir, outDir string) error {
	fmt.Println("Reading from:", srcDir)

	files, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("Failed to open src: %w", err)
	}

	var existingPhotos []photos.Photo

	dbPath := filepath.Join(outDir, "photos.db")
	db, err := database.Open(dbPath, "photos", slog.New(slog.NewTextHandler(os.Stdout, nil)))
	if err != nil {
		return fmt.Errorf("Failed to open database: %w", err)
	}

	photoRepo := photos.NewPhotosRepository(db)
	existingPhotos, err = readDatabase(photoRepo)
	if err != nil {
		return fmt.Errorf("Failed to read from database: %w", err)
	}

	existingSlugs := make(map[string]bool)
	for _, photo := range existingPhotos {
		existingSlugs[photo.Slug] = true
	}

	for _, file := range files {
		slug := generateSlug(file.Name())
		hasExisting := false

		if existingSlugs[slug] {
			fmt.Println("Has existing file:", file.Name())
			hasExisting = true
		}

		err = os.MkdirAll(outDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("Failed to open output: %w", err)
		}

		err = os.MkdirAll(filepath.Join(outDir, slug), os.ModePerm)
		if err != nil {
			return fmt.Errorf("Failed to generate slug directory: %w", err)
		}

		fullPath := filepath.Join(srcDir, file.Name())
		thumb, tw, th, err := generateImage(fullPath, thumbSize)
		if err != nil {
			return fmt.Errorf("Failed to create thumbnail: %w", err)
		}
		err = os.WriteFile(filepath.Join(outDir, slug, "thumb.jpg"), thumb, 0644)

		full, fw, fh, err := generateImage(fullPath, fullSize)
		if err != nil {
			return fmt.Errorf("Failed to create thumbnail: %w", err)
		}
		err = os.WriteFile(filepath.Join(outDir, slug, "full.jpg"), full, 0644)

		photo := photos.Photo{
			Slug:        slug,
			Original:    file.Name(),
			ThumbWidth:  tw,
			ThumbHeight: th,
			FullWidth:   fw,
			FullHeight:  fh,
		}

		err = populateExif(&photo, fullPath)
		if err != nil {
			return fmt.Errorf("Failed to populate exif: %w", err)
		}

		if hasExisting {
			for i, p := range existingPhotos {
				if p.Slug == slug {
					existingPhotos[i] = photo
					break
				}
			}
		} else {
			existingPhotos = append(existingPhotos, photo)
		}
	}

	err = writeDatabase(photoRepo, existingPhotos)
	if err != nil {
		return fmt.Errorf("Failed to write database: %w", err)
	}

	return nil
}

func generateSlug(name string) string {
	name = strings.TrimSuffix(name, filepath.Ext(name))
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	return name
}

func generateImage(srcPath string, size int) ([]byte, int, int, error) {
	img, err := vips.NewThumbnailFromFile(srcPath, size, size, vips.InterestingNone)
	if err != nil {
		return nil, 0, 0, err
	}

	buf, _, err := img.ExportJpeg(&vips.JpegExportParams{Quality: jpegQuality, Interlace: true})
	if err != nil {
		return nil, 0, 0, err
	}

	return buf, img.Width(), img.Height(), nil
}

func readDatabase(photoRepo *photos.PhotoRepository) ([]photos.Photo, error) {
	photos, err := photoRepo.ListAll(context.TODO())
	if err != nil {
		return nil, err
	}

	return photos, nil
}

func writeDatabase(photoRepo *photos.PhotoRepository, all []photos.Photo) error {
	for _, photo := range all {
		fmt.Println("Upserting photo", photo.Slug)
		err := photoRepo.Upsert(context.TODO(), photo)
		if err != nil {
			return err
		}
	}

	return nil
}

func populateExif(photo *photos.Photo, srcPath string) error {
	fmt.Println("Parsing exif for", srcPath)
	f, err := os.Open(srcPath)
	if err != nil {
		fmt.Println("Failed to open file for exif", err)
		return nil
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		fmt.Println("No exif to parse", err)
		return nil
	}

	if date, err := x.DateTime(); err == nil {
		layout := "2006-01-02T15:04:05"
		photo.Date = date.Format(layout)
		fmt.Println("Parsed date:", photo.Date)
	}

	if tag, err := x.Get(exif.Model); err == nil {
		if s, err := tag.StringVal(); err == nil {
			photo.Camera = s
		}
	}

	if tag, err := x.Get(exif.FNumber); err == nil {
		if r, err := tag.Rat(0); err == nil {
			v, _ := r.Float64()
			s := fmt.Sprintf("%.1f", v)
			photo.Aperture = "f/" + strings.TrimSuffix(s, ".0")
		}
	}

	if tag, err := x.Get(exif.ISOSpeedRatings); err == nil {
		if v, err := tag.Int(0); err == nil {
			photo.ISO = fmt.Sprintf("%d", v)
		}
	}

	if tag, err := x.Get(exif.ExposureTime); err == nil {
		if num, den, err := tag.Rat2(0); err == nil {
			if den == 1 {
				photo.Shutter = fmt.Sprintf("%d", num)
			} else {
				photo.Shutter = fmt.Sprintf("%d/%d", num, den)
			}
		}
	}

	if lat, lon, err := x.LatLong(); err == nil {
		photo.GPS = []float64{lat, lon}
	}

	return nil
}
