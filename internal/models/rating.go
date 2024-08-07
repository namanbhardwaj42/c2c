// rating.go

package models

import (
	"strings"
)

type RatingMap struct {
	RatingSystems []*RatingSystem `json:"rating_systems"`
}

func (r *RatingMap) GetRatingSystem(system_id string) *RatingSystem {
	for _, rs := range r.RatingSystems {
		if rs.ID == system_id {
			return rs
		}
	}
	return r.RatingSystems[0]
}

type RatingSystem struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Area        string `json:"area"`
	ContentType string `json:"contenttype"`
	Ratings     []struct {
		Title       string `json:"title"`
		Label       string `json:"label"`
		Description string `json:"description"`
		AgeHint     int    `json:"agehint"`
		Icon        string `json:"icon"`
	} `json:"ratings"`
}

func (r *RatingSystem) GetRatingLabel(title string) string {

	new_rating := title
	for _, r := range r.Ratings {
		if strings.ToLower(title) == r.Title {
			new_rating = r.Label
		}
	}

	new_rating = strings.ToUpper(new_rating)
	new_rating = strings.Replace(new_rating, "-", " ", 1)
	new_rating = strings.Replace(new_rating, "_", " ", 1)
	return new_rating
}
