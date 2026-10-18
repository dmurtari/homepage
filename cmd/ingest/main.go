package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
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
	useDB := flag.Bool("db", false, "Write to database")
	overwrite := flag.Bool("overwrite", false, "Overwrite existing images")

	flag.Parse()

	if err := run(*srcDir, *outDir, *overwrite, *useDB); err != nil {
		log.Fatal(err)
	}
}

func run(srcDir, outDir string, overwrite bool, useDB bool) error {
	fmt.Println("Reading from:", srcDir)

	files, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("Failed to open src: %w", err)
	}

	var existingPhotos []photos.Photo

	if useDB {
		dbPath := filepath.Join(outDir, "photos.db")
		existingPhotos, err = readDatabase(dbPath)
		if err != nil {
			return fmt.Errorf("Failed to read from database: %w", err)
		}
	} else {
		manifestPath := filepath.Join(outDir, "manifest.json")
		existingPhotos, err = parseManifest(manifestPath)
		if err != nil {
			return fmt.Errorf("Error reading manifest: %w", err)
		}
	}

	existingSlugs := make(map[string]bool)
	for _, photo := range existingPhotos {
		existingSlugs[photo.Slug] = true
	}

	for _, file := range files {
		slug := generateSlug(file.Name())
		overwrote := false

		if existingSlugs[slug] {
			fmt.Println("Skipping existing file:", file.Name())
			overwrote = true

			if !overwrite {
				continue
			}
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

		if overwrote {
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

	if useDB {
		err = writeDatabase(filepath.Join(outDir, "photos.db"), existingPhotos)
		if err != nil {
			return fmt.Errorf("Failed to write database: %w", err)
		}
	} else {
		manifestPath := filepath.Join(outDir, "manifest.json")
		err = writeManifest(manifestPath, existingPhotos)
		if err != nil {
			return fmt.Errorf("Failed to write manifest: %w", err)
		}
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

func parseManifest(manifestPath string) ([]photos.Photo, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var photos []photos.Photo
	err = json.Unmarshal(data, &photos)
	if err != nil {
		return nil, err
	}

	return photos, nil
}

func writeManifest(path string, all []photos.Photo) error {
	fmt.Println("Writing manifest to", path)
	data, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func readDatabase(path string) ([]photos.Photo, error) {
	fmt.Println("Reading database from", path)
	db, err := database.Open(path, "photos", nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	photoRepo := photos.NewPhotosRepository(db)

	photos, err := photoRepo.ListAll(context.TODO())
	if err != nil {
		return nil, err
	}

	return photos, nil
}

func writeDatabase(path string, all []photos.Photo) error {
	fmt.Println("Writing database to", path)
	db, err := database.Open(path, "photos", nil)
	if err != nil {
		return err
	}
	defer db.Close()

	photoRepo := photos.NewPhotosRepository(db)

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
