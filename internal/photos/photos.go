package photos

type Photo struct {
	Slug        string    `json:"slug"`
	Original    string    `json:"original"`
	Description string    `json:"description,omitempty"`
	Location    string    `json:"location,omitempty"`
	Camera      string    `json:"camera,omitempty"`
	Aperture    string    `json:"aperture,omitempty"`
	ISO         string    `json:"iso,omitempty"`
	Shutter     string    `json:"shutter,omitempty"`
	GPS         []float64 `json:"gps,omitempty"`
}
