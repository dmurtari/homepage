package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/rwcarlsen/goexif/exif"
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
	manifestDir := flag.String("manifest", "", "Manifest file")
	overwrite := flag.Bool("overwrite", false, "Overwrite existing images")

	flag.Parse()

	if err := run(*srcDir, *outDir, *manifestDir, *overwrite); err != nil {
		log.Fatal(err)
	}
}

func run(srcDir, outDir, manifestDir string, overwrite bool) error {
	fmt.Println("Reading from:", srcDir)

	files, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("Failed to open src: %w", err)
	}

	manifestPath := filepath.Join(manifestDir, "manifest.json")

	existingManifest, err := parseManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("Error reading manifest: %w", err)
	}

	existingSlugs := make(map[string]bool)
	for _, photo := range existingManifest {
		existingSlugs[photo.Slug] = true
	}

	for _, file := range files {
		slug := generateSlug(file.Name())
		overwrote := false

		// If file is already in the manifest, skip
		if existingSlugs[slug] && !overwrite {
			fmt.Println("Skipping existing file:", file.Name())
			overwrote = true
			continue
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
		thumb, err := generateImage(fullPath, thumbSize)
		if err != nil {
			return fmt.Errorf("Failed to create thumbnail: %w", err)
		}
		err = os.WriteFile(filepath.Join(outDir, slug, "thumb.jpg"), thumb, 0644)

		full, err := generateImage(fullPath, fullSize)
		if err != nil {
			return fmt.Errorf("Failed to create thumbnail: %w", err)
		}
		err = os.WriteFile(filepath.Join(outDir, slug, "full.jpg"), full, 0644)

		photo := photos.Photo{Slug: slug, Original: file.Name()}

		err = populateExif(&photo, fullPath)
		if err != nil {
			return fmt.Errorf("Failed to populate exif: %w", err)
		}

		if overwrote {
			for i, p := range existingManifest {
				if p.Slug == slug {
					existingManifest[i] = photo
				}
			}
		} else {
			existingManifest = append(existingManifest, photo)
		}
	}

	err = writeManifest(manifestPath, existingManifest)
	if err != nil {
		return fmt.Errorf("Failed to write manifest: %w", err)
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

func generateImage(srcPath string, size int) ([]byte, error) {
	img, err := vips.NewThumbnailFromFile(srcPath, size, size, vips.InterestingNone)
	if err != nil {
		return nil, err
	}

	buf, _, err := img.ExportJpeg(&vips.JpegExportParams{Quality: jpegQuality})
	if err != nil {
		return nil, err
	}

	return buf, nil
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
		if v, err := tag.Float(0); err == nil {
			photo.Aperture = fmt.Sprintf("f/%.1f", v)
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
