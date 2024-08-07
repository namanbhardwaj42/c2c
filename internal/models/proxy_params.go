// proxyRequest.go

package models

type ProxySearchQueryContext struct {
	Text               string  `json:"text,omitempty"`
	Context            string  `json:"context,omitempty"`
	Metatype           string  `json:"metatypes,omitempty"`
	VodYear            int     `json:"vod_year,omitempty"`
	VodEligibility     *string `json:"vod_eligibility,omitempty"`
	Type               *string `json:"in_type,omitempty"`
	ChannelMarket      *string `json:"in_market,omitempty"`
	ChannelSubMarket   *string `json:"in_submarket,omitempty"`
	Eligibility        *int    `json:"eligibility,omitempty"`
	PlatformID         string  `json:"platform_id,omitempty"`
	Number             int     `json:"number,omitempty"`
	AtDate             int64   `json:"at_date,omitempty"`
	RatingId           string  `json:"notin_rating,omitempty"`
	InPlatformID       string  `json:"in_platform_id,omitempty"`
	ChannelsPlatformID string  `json:"in_channel_platform_id,omitempty"`
	EpgSerieOnly       *bool   `json:"epg_serie_only,omitempty"`
}

type ProxySearchQuery struct {
	Context []ProxySearchQueryContext `json:"query"`
}

type ProxySearchQueries struct {
	Queries []ProxySearchQuery `json:"queries"`
}
