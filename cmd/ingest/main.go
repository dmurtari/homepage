package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
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
	webpQuality = 90
)

type encodedImage struct {
	JPEG, WebP    []byte
	Width, Height int
}

type PhotoIngester struct {
	srcDir    string
	outDir    string
	photoRepo *photos.PhotoRepository
	photos    []photos.Photo
}

func NewPhotoIngester(srcDir, outDir string, photoRepo *photos.PhotoRepository) *PhotoIngester {
	return &PhotoIngester{
		srcDir:    srcDir,
		outDir:    outDir,
		photoRepo: photoRepo,
		photos:    []photos.Photo{},
	}
}

func main() {
	srcDir := flag.String("src", "", "Source images")
	outDir := flag.String("out", "", "Output images")

	flag.Parse()

	dbPath := filepath.Join(*outDir, "photos.db")
	db, err := database.Open(dbPath, "photos", slog.New(slog.NewTextHandler(os.Stdout, nil)))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	err = os.MkdirAll(*outDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	photoRepo := photos.NewPhotosRepository(db)
	pi := NewPhotoIngester(*srcDir, *outDir, photoRepo)

	err = pi.Ingest()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	err = pi.WriteDatabase()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func (p *PhotoIngester) Ingest() error {
	err := filepath.WalkDir(p.srcDir, p.visit)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (p *PhotoIngester) visit(path string, di fs.DirEntry, err error) error {
	if di.IsDir() {
		return nil
	}

	slug := generateSlug(di.Name())
	// album := filepath.Base(filepath.Dir(path))

	err = os.MkdirAll(filepath.Join(p.outDir, slug), os.ModePerm)
	if err != nil {
		return fmt.Errorf("Failed to generate slug directory: %w", err)
	}

	thumb, err := generateImage(path, thumbSize)
	if err != nil {
		return fmt.Errorf("Failed to create thumbnail: %w", err)
	}
	if err := writeImages(filepath.Join(p.outDir, slug), "thumb", thumb); err != nil {
		return err
	}

	full, err := generateImage(path, fullSize)
	if err != nil {
		return fmt.Errorf("Failed to create full image: %w", err)
	}
	if err := writeImages(filepath.Join(p.outDir, slug), "full", full); err != nil {
		return err
	}

	photo := photos.Photo{
		Slug:        slug,
		Original:    di.Name(),
		ThumbWidth:  thumb.Width,
		ThumbHeight: thumb.Height,
		FullWidth:   full.Width,
		FullHeight:  full.Height,
	}

	err = populateExif(&photo, path)
	if err != nil {
		return fmt.Errorf("Failed to populate exif: %w", err)
	}

	p.photos = append(p.photos, photo)
	return nil
}

func (p *PhotoIngester) WriteDatabase() error {
	err := writeDatabase(p.photoRepo, p.photos)
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

func generateImage(srcPath string, size int) (*encodedImage, error) {
	vips.LoggingSettings(nil, vips.LogLevelWarning)

	jpgImg, err := vips.NewThumbnailFromFile(srcPath, size, size, vips.InterestingNone)
	if err != nil {
		return nil, fmt.Errorf("thumbnail (jpeg): %w", err)
	}
	defer jpgImg.Close()

	jpg, _, err := jpgImg.ExportJpeg(&vips.JpegExportParams{Quality: jpegQuality, Interlace: true})
	if err != nil {
		return nil, fmt.Errorf("jpeg export: %w", err)
	}

	webpImg, err := vips.NewThumbnailFromFile(srcPath, size, size, vips.InterestingNone)
	if err != nil {
		return nil, fmt.Errorf("thumbnail (webp): %w", err)
	}
	defer webpImg.Close()

	webp, _, err := webpImg.ExportWebp(&vips.WebpExportParams{
		Quality:         webpQuality,
		ReductionEffort: 4,
		StripMetadata:   true,
	})

	if err != nil {
		return nil, fmt.Errorf("webp export: %w", err)
	}

	return &encodedImage{
		JPEG:   jpg,
		WebP:   webp,
		Width:  jpgImg.Width(),
		Height: jpgImg.Height(),
	}, nil
}

func writeImages(dir, name string, img *encodedImage) error {
	if err := os.WriteFile(filepath.Join(dir, name+".jpg"), img.JPEG, 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name+".webp"), img.WebP, 0644)
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
