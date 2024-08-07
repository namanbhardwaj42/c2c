// pictures.go

package models

type ListPicture struct {
	Logos      []string `json:"logos"`
	LogosSmall []string `json:"logos_small"`
	LogosBig   []string `json:"logos_big"`
	Thumbnails []string `json:"thumbnails"`
	Backdrops  []string `json:"backdrops"`
}

type Picture struct {
	Type int    `json:"type" example:"0"`
	Url  string `json:"url" example:"picture.url.com"`
}
