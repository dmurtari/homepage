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
}

func (p Photo) ThumbPath() string { return MediaBaseURL + "/" + p.Slug + "/thumb.jpg" }
func (p Photo) FullPath() string  { return MediaBaseURL + "/" + p.Slug + "/full.jpg" }
