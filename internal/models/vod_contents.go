//vod_contents.go

package models

import (
	"c2c/internal/config"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type VodContentV1 struct {
	Metatype string `json:"metatype" gorm:"-"`

	ID         int    `json:"id"`
	PlatformID string `json:"platform_id,omitempty"`

	Provider string `json:"dataset,omitempty"`
	Device   int    `json:"device,omitempty"`

	Sandbox          string `json:"sandbox,omitempty"`
	Type             int    `json:"type,omitempty"`
	Runtime          int32  `json:"runtime"`
	Population       string `json:"population,omitempty"`
	PlatformProvider string `json:"content_provider,omitempty"`
	SourceID         string `json:"source_id,omitempty"`
	AltSourceID      string `json:"alt_source_id,omitempty"`

	ProductionYear string `json:"production_year,omitempty"`
	ReleaseDate    int    `json:"year,omitempty"`
	Availability   struct {
		Start string `json:"start,omitempty"`
		End   string `json:"end,omitempty"`
	} `json:"availability,omitempty"`
	SubscriptionId       []string `json:"subscription_id"`
	SubscriptionDetailId []string `json:"subscription_detail_id"`
	PurchaseType         int      `json:"purchase_type,omitempty"`

	ViewCount     int    `json:"view_count,omitempty"`
	RatingID      string `json:"rating_id"`
	Eligibility   string `json:"eligibility"`
	Number        int    `json:"number,omitempty"`
	Start         int    `json:"start,omitempty"`
	Tracking      bool   `json:"tracking,omitempty"`
	FastForward   bool   `json:"fast_forward,omitempty"`
	Format        int    `json:"format,omitempty"`
	AudioLanguage string `json:"audio_language,omitempty"`
	Caption       string `json:"caption,omitempty"`
	PlayableDate  int    `json:"playable_date,omitempty"`
	AdSupport     bool   `json:"ad_support,omitempty"`

	SerieMaculosaId      string `json:"serie_id,omitempty"`
	SeasonMaculosaId     string `json:"season_id,omitempty"`
	CollectionMaculosaId string `json:"collection_id,omitempty"`

	Title        string   `json:"title"`
	Subtitle     string   `json:"subtitle,omitempty"`
	Summary      string   `json:"summary"`
	ShortSummary string   `json:"short_summary"`
	Country      string   `json:"country"`
	Cast         []string `json:"cast"`
	Producers    []string `json:"producers"`
	Directors    []string `json:"directors"`
	Writers      []string `json:"writers"`
	Locale       string   `json:"locale"`

	NodeParentId []string `json:"source_node_id,omitempty"`

	// https://github.com/go-gorm/gorm/issues/3613
	Pictures ListPicture `json:"pictures"`

	ChildrenCount int `json:"children_count"`

	Genres string `json:"genre,omitempty"`

	Streams []*Stream `json:"streams,omitempty"`
}

func (v *VodContentV1) ConvertToGasItem(car *CloudApiRequest, assConfig *config.Assistant, ratingSystem *RatingSystem, st *SearchType) GasItem {
	var gi GasItem

	gi.Badgings = []Badging{}
	gi.CallToActions = []CallToAction{}

	gi.Title = v.Title
	gi.Description = v.Summary
	gi.ID = v.PlatformID
	gi.Rating = ratingSystem.GetRatingLabel(v.RatingID)
	if v.Type == 0 {
		gi.Type = "MOVIE"
	} else if v.Type == 1 {
		gi.Type = "TV_SHOW"
	}
	gi.Genre = clean_genres(strings.Split(v.Genres, ","), assConfig.SearchConfig.MaxDisplayCategory)
	gi.Duration = v.Runtime
	gi.ReleaseStartYear = strconv.Itoa(v.ReleaseDate)
	// gi.TMSID = v.AltSourceID
	// gi.ReleaseEndYear = strconv.Itoa(v.ReleaseDate)
	gi.Network = v.PlatformProvider

	if !assConfig.SearchConfig.HidePosterImage && len(v.Pictures.Thumbnails) > 0 {
		var gp GasPicture
		gp.ImageUrl = v.Pictures.Thumbnails[0]
		gp.Width = 318
		gp.Height = 477
		gi.PosterImage = gp
	}

	// Create badges
	if assConfig.IsBadgeEnabled("NEW_CONTENT") {
		now := time.Now().Add(time.Duration(-assConfig.SearchConfig.Recency) * time.Second).UTC()
		layout := "2006-01-02T15:04:05-07:00"
		if start_availability, err := time.Parse(layout, v.Availability.Start); err == nil && start_availability.After(now) {
			var b Badging
			b.StaticBadge.BadgeID = "NEW_CONTENT"
			b.Type = "NEW_OR_FEATURED"
			gi.Badgings = append(gi.Badgings, b)
		}
	}
	// 1 == TVOD, 3 == SVOD
	// TODO USE ENUM!
	if assConfig.IsBadgeEnabled("NOT_ENTITLED") &&
		(v.Type == 1 && !car.isEntitled(v.PlatformID) ||
			v.Type == 3 && !car.isSubscribed(v.SubscriptionId) ||
			(v.Type != 1 && v.Type != 3)) {
		var b Badging
		b.StaticBadge.BadgeID = "NOT_ENTITLED"
		b.Type = "PRICING"
		gi.Badgings = append(gi.Badgings, b)
	}

	if assConfig.IsBadgeEnabled("ENTITY_LOGO") &&
		st.SubType == "ENTITY_SEARCH" {
		var b Badging
		b.StaticBadge.BadgeID = "ENTITY_LOGO"
		b.Type = "SYSTEM"
		gi.Badgings = append(gi.Badgings, b)
	}

	// Create CTAs
	if ctaO := assConfig.GetCTA("PLAY"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) {
		var cta CallToAction
		var ai AndroidIntent
		// ai.ContentUri = fmt.Sprintf("smartui://play/vod/%s", v.PlatformID)
		ai.ContentUri = fmt.Sprintf("smartui://play/vod/%s", v.PlatformID)
		ai.Action = "android.intent.action.VIEW"
		cta.AndroidIntent = ai
		cta.FromResumePoint = false
		cta.Title = ctaO.GetLabel(car.LanguageCode)
		cta.Type = "PLAY"
		cta.ContentSource = "VOD"
		gi.CallToActions = append(gi.CallToActions, cta)
	}

	if ctaO := assConfig.GetCTA("RESTART"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) {
		var cta CallToAction
		var ai AndroidIntent
		ai.ContentUri = fmt.Sprintf("smartui://restart/vod/%s", v.PlatformID)
		ai.Action = "android.intent.action.VIEW"
		cta.AndroidIntent = ai
		cta.FromResumePoint = false
		cta.Title = ctaO.GetLabel(car.LanguageCode)
		cta.Type = "RESTART"
		cta.ContentSource = "VOD"
		gi.CallToActions = append(gi.CallToActions, cta)
	}

	if ctaO := assConfig.GetCTA("PLAY_TRAILER"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) {
		hasTrailer := false

		for _, stream := range v.Streams {
			if stream.Type == 1 {
				hasTrailer = true
				break
			}
		}

		if hasTrailer {
			var cta CallToAction
			var ai AndroidIntent
			ai.ContentUri = fmt.Sprintf("smartui://trailer/vod/%s", v.PlatformID)
			ai.Action = "android.intent.action.VIEW"
			cta.AndroidIntent = ai
			cta.FromResumePoint = false
			cta.Title = ctaO.GetLabel(car.LanguageCode)
			cta.Type = "PLAY_TRAILER"
			cta.ContentSource = "VOD"
			gi.CallToActions = append(gi.CallToActions, cta)
		}
	}

	if ctaO := assConfig.GetCTA("OPEN_APP"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) {
		var cta CallToAction
		var ai AndroidIntent
		ai.ContentUri = fmt.Sprintf("smartui://vod_content/%s", v.PlatformID)
		ai.Action = "android.intent.action.VIEW"
		cta.AndroidIntent = ai
		cta.FromResumePoint = false
		cta.Title = ctaO.GetLabel(car.LanguageCode)
		cta.Type = "OPEN_APP"
		cta.ContentSource = "VOD"
		gi.CallToActions = append(gi.CallToActions, cta)
	}

	if ctaO := assConfig.GetCTA("CUSTOM2"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) {
		var cta CallToAction
		var ai AndroidIntent
		ai.ContentUri = fmt.Sprintf("smartui://favorites/add/vod/%s", v.PlatformID)
		ai.Action = "android.intent.action.VIEW"
		cta.AndroidIntent = ai
		cta.FromResumePoint = false
		cta.Title = ctaO.GetLabel(car.LanguageCode)
		cta.Type = "CUSTOM2"
		cta.ContentSource = "VOD"
		gi.CallToActions = append(gi.CallToActions, cta)
	}

	return gi
}

type VodContentV3 struct {
	Metatype string `json:"metatype" gorm:"-"`

	ID         string `json:"id"`
	PlatformId string `json:"platform_id,omitempty"`

	Provider string `json:"dataset,omitempty"`
	Device   int    `json:"device,omitempty"`

	Sandbox          string `json:"sandbox,omitempty"`
	Type             int    `json:"type,omitempty"`
	Runtime          int32  `json:"runtime"`
	Population       string `json:"population,omitempty"`
	PlatformProvider string `json:"platform_provider,omitempty"`
	SourceID         string `json:"source_id,omitempty"`

	ProductionYear       string   `json:"production_year,omitempty"`
	ReleaseDate          int      `json:"release_date,omitempty"`
	StartValidity        int64    `json:"start_validity,omitempty"`
	EndValidity          int64    `json:"end_validity,omitempty"`
	SubscriptionId       []string `json:"subscription_id"`
	SubscriptionDetailId []string `json:"subscription_detail_id"`
	PurchaseType         int      `json:"purchase_type,omitempty"`

	ViewCount     int    `json:"view_count,omitempty"`
	RatingID      string `json:"rating_id"`
	Eligibility   int    `json:"eligibility"`
	Number        int    `gorm:"column:vnumber" json:"number,omitempty"`
	Start         int    `gorm:"column:starttime" json:"start,omitempty"`
	Tracking      bool   `json:"tracking,omitempty"`
	FastForward   bool   `json:"fast_forward,omitempty"`
	Format        int    `json:"format,omitempty"`
	AudioLanguage string `json:"audio_language,omitempty"`
	Caption       string `json:"caption,omitempty"`
	PlayableDate  int    `json:"playable_date,omitempty"`
	AdSupport     bool   `json:"ad_support,omitempty"`

	SerieID      string `json:"serie_id,omitempty"`
	SeasonID     string `json:"season_id,omitempty"`
	CollectionID string `json:"collection_id,omitempty"`

	Title        string    `json:"title"`
	Subtitle     string    `json:"subtitle,omitempty"`
	Summary      string    `json:"description"`
	ShortSummary string    `json:"short_description"`
	Country      string    `json:"country"`
	Cast         []*Person `json:"cast"`
	Producers    []*Person `json:"producers"`
	Directors    []*Person `json:"directors"`
	Writers      []*Person `json:"writers"`
	Locale       string    `json:"locale"`

	NodeParentId []string `json:"source_node_id,omitempty"`

	// https://github.com/go-gorm/gorm/issues/3613
	Pictures []*Picture `json:"pictures"`

	ChildrenCount int `json:"children_count"`

	Genres []string `json:"genres,omitempty"`

	Streams []*Stream `json:"streams,omitempty"`
}

func (v *VodContentV3) ConvertToGasItem(car *CloudApiRequest, assConfig *config.Assistant, ratingSystem *RatingSystem, st *SearchType) GasItem {
	var gi GasItem

	gi.Badgings = []Badging{}
	gi.CallToActions = []CallToAction{}

	gi.Title = v.Title
	gi.Description = v.Summary
	gi.ID = v.PlatformId
	gi.Rating = ratingSystem.GetRatingLabel(v.RatingID)
	if v.Type == 0 {
		gi.Type = "MOVIE"
	} else if v.Type == 1 {
		gi.Type = "TV_SHOW"
	}
	gi.Genre = clean_genres(v.Genres, assConfig.SearchConfig.MaxDisplayCategory)
	gi.Duration = v.Runtime
	gi.ReleaseStartYear = v.ProductionYear
	// gi.TMSID = v.AltSourceID
	// gi.ReleaseEndYear = strconv.Itoa(v.ReleaseDate)
	if assConfig.SearchConfig.DisplayNetwork {
		gi.Network = v.PlatformProvider
	}

	if !assConfig.SearchConfig.HidePosterImage {
		for _, pict := range v.Pictures {
			if pict.Type == 4 {
				var gp GasPicture
				gp.ImageUrl = pict.Url
				gp.Width = 318
				gp.Height = 477
				gi.PosterImage = gp
				break
			}
		}
	}

	// Create badges
	if assConfig.IsBadgeEnabled("NEW_CONTENT") {
		now := time.Now().Add(time.Duration(-assConfig.SearchConfig.Recency) * time.Second).UTC()
		if start_availability := time.Unix(v.StartValidity, 0); start_availability.After(now) {
			var b Badging
			b.StaticBadge.BadgeID = "NEW_CONTENT"
			b.Type = "USER_ACTION"
			gi.Badgings = append(gi.Badgings, b)
		}
	}
	// 1 == TVOD, 3 == SVOD
	// TODO USE ENUM!
	if assConfig.IsBadgeEnabled("NOT_ENTITLED") &&
		(v.Type == 1 && !car.isEntitled(v.PlatformId) ||
			v.Type == 3 && !car.isSubscribed(v.SubscriptionId) ||
			(v.Type != 1 && v.Type != 3)) {
		var b Badging
		b.StaticBadge.BadgeID = "NOT_ENTITLED"
		b.Type = "PRICING"
		gi.Badgings = append(gi.Badgings, b)
	}

	if assConfig.IsBadgeEnabled("ENTITY_LOGO") &&
		st.SubType == "ENTITY_SEARCH" {
		var b Badging
		b.StaticBadge.BadgeID = "ENTITY_LOGO"
		b.Type = "SYSTEM"
		gi.Badgings = append(gi.Badgings, b)
	}

	// Create CTAs
	if ctaO := assConfig.GetCTA("PLAY"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) {
		var cta CallToAction
		var ai AndroidIntent
		// ai.ContentUri = fmt.Sprintf("smartui://play/vod/%s", v.PlatformId)
		ai.ContentUri = fmt.Sprintf("smartui://play/vod/%s", v.PlatformId)
		ai.Action = "android.intent.action.VIEW"
		cta.AndroidIntent = ai
		cta.FromResumePoint = false
		cta.Title = ctaO.GetLabel(car.LanguageCode)
		cta.Type = "PLAY"
		cta.ContentSource = "VOD"
		gi.CallToActions = append(gi.CallToActions, cta)
	}

	if ctaO := assConfig.GetCTA("RESTART"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) {
		var cta CallToAction
		var ai AndroidIntent
		ai.ContentUri = fmt.Sprintf("smartui://restart/vod/%s", v.PlatformId)
		ai.Action = "android.intent.action.VIEW"
		cta.AndroidIntent = ai
		cta.FromResumePoint = false
		cta.Title = ctaO.GetLabel(car.LanguageCode)
		cta.Type = "RESTART"
		cta.ContentSource = "VOD"
		gi.CallToActions = append(gi.CallToActions, cta)
	}

	if ctaO := assConfig.GetCTA("PLAY_TRAILER"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) {
		hasTrailer := false

		for _, stream := range v.Streams {
			if stream.Type == 1 {
				hasTrailer = true
				break
			}
		}

		if hasTrailer {
			var cta CallToAction
			var ai AndroidIntent
			ai.ContentUri = fmt.Sprintf("smartui://trailer/vod/%s", v.PlatformId)
			ai.Action = "android.intent.action.VIEW"
			cta.AndroidIntent = ai
			cta.FromResumePoint = false
			cta.Title = ctaO.GetLabel(car.LanguageCode)
			cta.Type = "PLAY_TRAILER"
			cta.ContentSource = "VOD"
			gi.CallToActions = append(gi.CallToActions, cta)
		}
	}

	if ctaO := assConfig.GetCTA("OPEN_APP"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) {
		var cta CallToAction
		var ai AndroidIntent
		ai.Action = "android.intent.action.VIEW"
		ai.ContentUri = fmt.Sprintf("smartui://vod_content/%s", v.PlatformId)
		cta.AndroidIntent = ai
		cta.FromResumePoint = false
		cta.Title = ctaO.GetLabel(car.LanguageCode)
		cta.Type = "OPEN_APP"
		cta.ContentSource = "VOD"
		gi.CallToActions = append(gi.CallToActions, cta)
	}

	if ctaO := assConfig.GetCTA("CUSTOM2"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) {
		var cta CallToAction
		var ai AndroidIntent
		ai.ContentUri = fmt.Sprintf("smartui://favorites/add/vod/%s", v.PlatformId)
		ai.Action = "android.intent.action.VIEW"
		cta.AndroidIntent = ai
		cta.FromResumePoint = false
		cta.Title = ctaO.GetLabel(car.LanguageCode)
		cta.Type = "CUSTOM2"
		cta.ContentSource = "VOD"
		gi.CallToActions = append(gi.CallToActions, cta)
	}

	return gi
}
