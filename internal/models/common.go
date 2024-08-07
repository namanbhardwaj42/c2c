// common.go

package models

type Content struct {
	Metatype string `json:"metatype"`
}

type SearchType struct {
	Type           string
	SubType        string
	CanFallback    bool
	ShouldFallback bool
}
