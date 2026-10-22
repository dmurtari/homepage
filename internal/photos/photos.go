package photos

var MediaBaseURL = ""

type Photo struct {
	Slug        string    `json:"slug"`
	Original    string    `json:"original"`
	Date        string    `json:"date,omitempty"`
	ThumbWidth  int       `json:"thumb_width,omitempty"`
	ThumbHeight int       `json:"thumb_height,omitempty"`
	FullWidth   int       `json:"full_width,omitempty"`
	FullHeight  int       `json:"full_height,omitempty"`
	Description string    `json:"description,omitempty"`
	Location    string    `json:"location,omitempty"`
	Camera      string    `json:"camera,omitempty"`
	Aperture    string    `json:"aperture,omitempty"`
	ISO         string    `json:"iso,omitempty"`
	Shutter     string    `json:"shutter,omitempty"`
	GPS         []float64 `json:"gps,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	AlbumID     string    `json:"album_id,omitempty"`
}

func (p Photo) ThumbPath() string     { return MediaBaseURL + "/" + p.Slug + "/thumb.jpg" }
func (p Photo) FullPath() string      { return MediaBaseURL + "/" + p.Slug + "/full.jpg" }
func (p Photo) WebpThumbPath() string { return MediaBaseURL + "/" + p.Slug + "/thumb.webp" }
func (p Photo) WebpFullPath() string  { return MediaBaseURL + "/" + p.Slug + "/full.webp" }

type Album struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
