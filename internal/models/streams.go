// streams.go

package models

type Stream struct {
	Type      int    `json:"type"`
	Uri       string `json:"uri"`
	PlaybacId string `json:"playback_id"`
}
