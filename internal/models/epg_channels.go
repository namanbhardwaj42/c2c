//epg_channels.go

package models

type ChannelRight struct {
	Recordable bool `json:"recordable" example:"true"`
	StartOver  bool `json:"start_over" example:"true"`
	Shareable  bool `json:"shareable" example:"false"`
	Catchup    bool `json:"catchup" example:"false"`
}

type ChannelAnalytics struct {
	Parameters string `json:"parameters,omitempty" example:"areuh"`
}

type EpgChannelV1 struct {
	Metatype string `json:"metatype"`

	ID              int    `json:"id"`
	PlatformID      string `json:"platform_id" binding:"required" example:"123"`
	Provider        string `json:"dataset,omitempty" binding:"required" example:"epg_provider"`
	Sandbox         string `json:"sandbox,omitempty" example:""`
	SourceId        string `json:"source_id" example:"123_a"`
	PlaybackUrl     string `json:"playback_url" example:"stream.url.com"`
	PlaybackId      string `json:"playback_id" example:"stream.1.23_a"`
	StartOverUrl    string `json:"startover_url" example:"stream.url.com"`
	Number          int    `json:"number" binding:"required" example:"17"`
	Type            int    `json:"type" example:"0"`
	Iptv            bool   `json:"iptv"`
	Ott             bool   `json:"ott"`
	AdSupport       bool   `json:"ad_support"`
	DrmProvider     string `json:"drm_provider"`
	RatingId        string `json:"rating_id" example:"1"`
	SubscriptionIds string `json:"subscription_id" example:"sub_1,sub_2"`
	Locale          string `json:"locale,omitempty" binding:"required" example:"en_us"`

	StartValidity int `json:"availability_start"`
	EndValidity   int `json:"availability_end"`

	ChannelRight     ChannelRight      `json:"rights"`
	ChannelAnalytics *ChannelAnalytics `json:"analytics,omitempty"`

	Genre   string  `json:"category"`
	Title   *string `json:"name" binding:"required" example:"Channel_17"`
	Summary *string `json:"description" binding:"required" example:"This is channel 17"`

	Pictures ListPicture `json:"pictures"`
}

type EpgChannelV3 struct {
	Metatype string `json:"metatype"`

	MaculosaId      string   `json:"id"`
	PlatformId      string   `json:"platform_id" binding:"required" example:"123"`
	Device          int      `json:"device,omitempty" binding:"required" example:"2"`
	Provider        string   `json:"dataset,omitempty" binding:"required" example:"epg_provider"`
	Sandbox         string   `json:"sandbox,omitempty" example:""`
	SourceId        string   `json:"source_id" example:"123_a"`
	PlaybackUrl     string   `json:"playback_url" example:"stream.url.com"`
	PlaybackId      string   `json:"playback_id" example:"stream.1.23_a"`
	StartOverUrl    string   `json:"startover_url" example:"stream.url.com"`
	Number          int      `json:"number" binding:"required" example:"17"`
	Type            int      `json:"type" example:"0"`
	Iptv            bool     `json:"iptv"`
	Ott             bool     `json:"ott"`
	AdSupport       bool     `json:"ad_support"`
	DrmProvider     string   `json:"drm_provider"`
	RatingId        string   `json:"rating_id" example:"1"`
	SubscriptionIds []string `json:"subscription_id" example:"sub_1,sub_2"`
	Locale          string   `json:"locale,omitempty" binding:"required" example:"en_us"`

	StartValidity int `json:"availability_start"`
	EndValidity   int `json:"availability_end"`

	ChannelRight     ChannelRight      `json:"rights"`
	ChannelAnalytics *ChannelAnalytics `json:"analytics,omitempty"`

	Title   *string `json:"title" binding:"required" example:"Channel_17"`
	Summary *string `json:"description" binding:"required" example:"This is channel 17"`

	Pictures []*Picture `json:"pictures"`
}
