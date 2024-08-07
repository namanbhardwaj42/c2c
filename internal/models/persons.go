// persons.go

package models

type Person struct {
	ID         string `json:"id"`
	PlatformId string `json:"platform_id,omitempty"`

	Name   string `json:"name"`
	Middle string `json:"middle"`
	Last   string `json:"last"`

	Locale string `json:"locale"`

	Role int `json:"role"`

	Pictures []*Picture `json:"pictures"`
}
