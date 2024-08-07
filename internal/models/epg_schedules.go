//epg_schedules.go

package models

import (
	"c2c/internal/config"
	"c2c/internal/tools"
	"fmt"
	"strings"
	"time"
)

type ScheduleBlackoutRight struct {
	Blackout           bool   `json:"enable" example:"true"`
	BlackoutNetwork    string `json:"network" example:"roger"`
	BlackoutMarketcode string `json:"marketcode" example:"aze"`
	BlackoutMessage    string `json:"message" example:"test"`
}

type ScheduleRecordRight struct {
	Capable            bool `json:"capable" example:"true"`
	Type               int  `json:"type" example:"3"`
	Expiry             int  `json:"expiry" example:"2"`
	ForbiddenTrickmode int  `json:"forbidden_trickmode" example:"1"`
}

type ScheduleTimeshiftRight struct {
	StartOver          bool `json:"start_over" example:"false"`
	WatchAgain         bool `json:"watch_again" example:"true"`
	Expiry             int  `json:"expiry" example:"1"`
	ForbiddenTrickmode int  `json:"forbidden_trickmode" example:"2"`
}

type ScheduleRight struct {
	ScheduleRecordRight    ScheduleRecordRight    `json:"record"`
	ScheduleTimeshiftRight ScheduleTimeshiftRight `json:"timeshift"`
	ScheduleBlackoutRight  ScheduleBlackoutRight  `json:"blackout"`
}

type EpgScheduleV1 struct {
	Metatype string `json:"metatype" gorm:"-"`

	ID         int    `json:"id"`
	PlatformID string `json:"platform_id,omitempty"`

	Provider string `json:"dataset,omitempty"`

	Sandbox  string `json:"sandbox,omitempty"`
	SourceID string `json:"source_id,omitempty"`

	Start         string `json:"broadcast_datetime,omitempty"`
	End           string `json:"broadcast_end_datetime,omitempty"`
	OriginalStart string `json:"original_broadcast_datetime,omitempty"`
	Runtime       int32  `json:"runtime,omitempty"`

	ChannelId         int    `json:"channel_id"`
	ChannelPlatformId string `json:"channel_platform_id"`

	ScheduleRight ScheduleRight `json:"rights"`

	Format int `json:"format,omitempty"`

	SubscriptionId string `json:"subscription_id,omitempty"`

	RatingID string `json:"rating_id,omitempty"`

	SeasonID      string `json:"season_id,omitempty"`
	SerieID       string `json:"serie_id,omitempty"`
	SeasonNumber  int    `json:"season_number,omitempty"`
	EpisodeNumber int    `json:"episode_number,omitempty"`
	SerieTitle    string `json:"serie_title,omitempty"`

	Title        string `json:"title,omitempty"`
	Subtitle     string `json:"subtitle,omitempty"`
	Summary      string `json:"description,omitempty"`
	ShortSummary string `json:"short_description,omitempty"`

	Ppv        bool   `json:"ppv"`
	Definition string `json:"definition"`

	Pictures ListPicture `json:"pictures"`
	Locale   string      `json:"locale"`

	Genres    string `json:"category,omitempty"`
	SubGenres string `json:"subcategory,omitempty"`
}

func (s *EpgScheduleV1) ConvertToGasItem(car *CloudApiRequest, assConfig *config.Assistant, ratingSystem *RatingSystem, st *SearchType) GasItem {
	var gi GasItem

	gi.Badgings = []Badging{}
	gi.CallToActions = []CallToAction{}

	gi.Title = s.Title
	gi.Description = s.Summary
	gi.ID = s.PlatformID
	gi.Rating = ratingSystem.GetRatingLabel(s.RatingID)
	gi.Type = "TV_SHOW"
	gi.Genre = clean_genres(strings.Split(s.Genres, ","), assConfig.SearchConfig.MaxDisplayCategory)
	gi.Duration = s.Runtime
	if !assConfig.SearchConfig.HidePosterImage && len(s.Pictures.Thumbnails) > 0 {
		var gp GasPicture
		gp.ImageUrl = s.Pictures.Thumbnails[0]
		gp.Width = 477
		gp.Height = 288
		gi.PosterImage = gp
	}

	now := time.Now().UTC()
	layout := "2006-01-02T15:04:05-07:00"
	start, err := time.Parse(layout, s.Start)
	if err != nil {
		tools.Logger.Error(err)
	}
	end, err := time.Parse(layout, s.End)
	if err != nil {
		tools.Logger.Error(err)
	}

	if assConfig.IsBadgeEnabled("TIME_LEFT") {
		if start.Before(now) && now.Before(end) {
			var b Badging
			b.StaticBadge.BadgeID = "TIME_LEFT"
			b.Type = "USER_ACTION"
			gi.Badgings = append(gi.Badgings, b)
		}
	}

	if assConfig.IsBadgeEnabled("RECORDING") && car.isRecordStatus(s.PlatformID, []string{"recording"}) {
		var b Badging
		b.StaticBadge.BadgeID = "RECORDING"
		b.Type = "USER_ACTION"
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
		ai.ContentUri = fmt.Sprintf("smartui://play_livetv/%s", s.ChannelPlatformId)
		ai.Action = "android.intent.action.VIEW"
		cta.AndroidIntent = ai
		cta.FromResumePoint = false
		cta.Title = ctaO.GetLabel(car.LanguageCode)
		cta.Type = "PLAY"
		cta.ContentSource = "LIVE"
		gi.CallToActions = append(gi.CallToActions, cta)
	}

	if ctaO := assConfig.GetCTA("RECORD"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) && now.Before(end) {
		var cta CallToAction
		var ai AndroidIntent
		ai.ContentUri = fmt.Sprintf("smartui://record_content/%s", s.PlatformID)
		ai.Action = "android.intent.action.VIEW"
		cta.AndroidIntent = ai
		cta.FromResumePoint = false
		cta.Title = ctaO.GetLabel(car.LanguageCode)
		cta.Type = "REC_CONTENT"
		cta.ContentSource = "LIVE"
		gi.CallToActions = append(gi.CallToActions, cta)
	}

	if ctaO := assConfig.GetCTA("CANCEL_RECORDING"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) && car.isRecordStatus(s.PlatformID, []string{"recording", "scheduled"}) {
		var cta CallToAction
		var ai AndroidIntent
		ai.ContentUri = fmt.Sprintf("smartui://record_content/cancel/%s", s.PlatformID)
		ai.Action = "android.intent.action.VIEW"
		cta.AndroidIntent = ai
		cta.FromResumePoint = false
		cta.Title = ctaO.GetLabel(car.LanguageCode)
		cta.Type = "REC_CANCEL"
		cta.ContentSource = "LIVE"
		gi.CallToActions = append(gi.CallToActions, cta)
	}

	if ctaO := assConfig.GetCTA("DELETE_RECORDING"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) && car.isRecordStatus(s.PlatformID, []string{"recorded"}) {
		var cta CallToAction
		var ai AndroidIntent
		ai.ContentUri = fmt.Sprintf("smartui://record_content/delete/%s", s.PlatformID)
		ai.Action = "android.intent.action.VIEW"
		cta.AndroidIntent = ai
		cta.FromResumePoint = false
		cta.Title = ctaO.GetLabel(car.LanguageCode)
		cta.Type = "REC_DELETE"
		cta.ContentSource = "LIVE"
		gi.CallToActions = append(gi.CallToActions, cta)
	}

	if ctaO := assConfig.GetCTA("RESTART"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) && now.After(end) && s.ScheduleRight.ScheduleTimeshiftRight.StartOver {
		var cta CallToAction
		var ai AndroidIntent
		ai.ContentUri = fmt.Sprintf("smartui://restart/epg/%s", s.PlatformID)
		ai.Action = "android.intent.action.VIEW"
		cta.AndroidIntent = ai
		cta.FromResumePoint = false
		cta.Title = ctaO.GetLabel(car.LanguageCode)
		cta.Type = "RESTART"
		cta.ContentSource = "LIVE"
		gi.CallToActions = append(gi.CallToActions, cta)
	}

	if ctaO := assConfig.GetCTA("OPEN_APP"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) {
		var cta CallToAction
		var ai AndroidIntent
		ai.ContentUri = fmt.Sprintf("smartui://epg/%s", s.PlatformID)
		ai.Action = "android.intent.action.VIEW"
		cta.AndroidIntent = ai
		cta.FromResumePoint = false
		cta.Title = ctaO.GetLabel(car.LanguageCode)
		cta.Type = "OPEN_APP"
		cta.ContentSource = "LIVE"
		gi.CallToActions = append(gi.CallToActions, cta)
	}

	if ctaO := assConfig.GetCTA("CUSTOM1"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) {
		var cta CallToAction
		var ai AndroidIntent
		ai.ContentUri = fmt.Sprintf("smartui://reminder/add/%s", s.PlatformID)
		ai.Action = "android.intent.action.VIEW"
		cta.AndroidIntent = ai
		cta.FromResumePoint = false
		cta.Title = ctaO.GetLabel(car.LanguageCode)
		cta.Type = "CUSTOM2"
		cta.ContentSource = "LIVE"
		gi.CallToActions = append(gi.CallToActions, cta)
	}

	return gi
}

type EpgScheduleV3 struct {
	Metatype string `json:"metatype" gorm:"-"`

	ID         int    `json:"-"`
	MaculosaId string `json:"id"`
	PlatformId string `json:"platform_id,omitempty"`

	Provider string `json:"dataset,omitempty"`
	Device   int    `json:"device,omitempty"`

	Sandbox  string `json:"sandbox,omitempty"`
	SourceID string `json:"source_id,omitempty"`

	Start         int64 `gorm:"column:starttime" json:"start,omitempty"`
	End           int64 `gorm:"column:endtime" json:"end,omitempty"`
	Runtime       int32 `json:"runtime,omitempty"`
	OriginalStart int   `gorm:"column:original_starttime" json:"original_starttime,omitempty"`

	ChannelId         string `json:"channel_id"`
	ChannelPlatformId string `json:"channel_platform_id"`

	PrimeTimeLevel int `json:"prime_time_level,omitempty"`

	ScheduleRight ScheduleRight `json:"rights"`

	Format int `json:"format,omitempty"`

	SubscriptionId string `json:"subscription_id,omitempty"`

	RatingID string `json:"rating,omitempty"`

	SeasonID      string `json:"season_id,omitempty"`
	SerieID       string `json:"serie_id,omitempty"`
	SeasonNumber  int    `json:"season_number,omitempty"`
	EpisodeNumber int    `json:"episode_number,omitempty"`
	SerieTitle    string `json:"serie_title,omitempty"`

	Title        string `json:"title,omitempty"`
	Subtitle     string `json:"subtitle,omitempty"`
	Summary      string `json:"description,omitempty"`
	ShortSummary string `json:"short_description,omitempty"`

	ContentType int `json:"content_type, omitempty"`

	Cast      []Person `json:"casts"`
	Producers []Person `json:"producers"`
	Directors []Person `json:"directors"`

	Ppv        bool   `json:"ppv"`
	Definition string `json:"definition"`

	// https://github.com/go-gorm/gorm/issues/3613
	Pictures []*Picture `json:"pictures"`

	Locale string `json:"locale"`

	Genres []string `json:"genres,omitempty"`

	Streams []*Stream `gorm:"foreignKey:EpgScheduleId" json:"streams,omitempty"`
}

func (s *EpgScheduleV3) ConvertToGasItem(car *CloudApiRequest, assConfig *config.Assistant, ratingSystem *RatingSystem, st *SearchType) GasItem {
	var gi GasItem

	gi.Badgings = []Badging{}
	gi.CallToActions = []CallToAction{}

	if assConfig.SearchConfig.EpgSerieGathering && s.ContentType == 1 {
		gi.Title = fmt.Sprintf("Binded to %s", s.SerieTitle)
		gi.Description = ""
	} else {
		gi.Title = s.Title
		gi.Description = s.Summary
	}

	gi.ID = s.PlatformId
	gi.Rating = ratingSystem.GetRatingLabel(s.RatingID)
	gi.Type = "TV_SHOW"
	gi.Genre = clean_genres(s.Genres, assConfig.SearchConfig.MaxDisplayCategory)
	gi.Duration = s.Runtime
	if !assConfig.SearchConfig.HidePosterImage {
		for _, pict := range s.Pictures {
			if pict.Type == 4 {
				var gp GasPicture
				gp.ImageUrl = pict.Url
				gp.Width = 477
				gp.Height = 288
				gi.PosterImage = gp
				break
			}
		}
	}

	now := time.Now().UTC()
	start := time.Unix(s.Start, 0)
	end := time.Unix(s.End, 0)

	if assConfig.IsBadgeEnabled("TIME_LEFT") {
		if start.Before(now) && now.Before(end) {
			var b Badging
			b.StaticBadge.BadgeID = "TIME_LEFT"
			b.Type = "USER_ACTION"
			gi.Badgings = append(gi.Badgings, b)
		}
	}

	if assConfig.IsBadgeEnabled("RECORDING") && car.isRecordStatus(s.PlatformId, []string{"recording"}) {
		var b Badging
		b.StaticBadge.BadgeID = "RECORDING"
		b.Type = "USER_ACTION"
		gi.Badgings = append(gi.Badgings, b)
	}

	if assConfig.IsBadgeEnabled("ENTITY_LOGO") &&
		st.SubType == "ENTITY_SEARCH" {
		var b Badging
		b.StaticBadge.BadgeID = "ENTITY_LOGO"
		b.Type = "SYSTEM"
		gi.Badgings = append(gi.Badgings, b)
	}

	if assConfig.SearchConfig.EpgSerieGathering && s.ContentType == 1 {
		if ctaO := assConfig.GetCTA("OPEN_APP"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) {
			var cta CallToAction
			var ai AndroidIntent
			ai.ContentUri = fmt.Sprintf("smartui://epg_serie/%s", s.SerieID)
			ai.Action = "android.intent.action.VIEW"
			cta.AndroidIntent = ai
			cta.FromResumePoint = false
			cta.Title = ctaO.GetLabel(car.LanguageCode)
			cta.Type = "OPEN_APP"
			cta.ContentSource = "LIVE"
			gi.CallToActions = append(gi.CallToActions, cta)
		}
	} else {
		// Create CTAs
		if ctaO := assConfig.GetCTA("PLAY"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) {
			var cta CallToAction
			var ai AndroidIntent
			ai.ContentUri = fmt.Sprintf("smartui://play_livetv/%s", s.ChannelPlatformId)
			ai.Action = "android.intent.action.VIEW"
			cta.AndroidIntent = ai
			cta.FromResumePoint = false
			cta.Title = ctaO.GetLabel(car.LanguageCode)
			cta.Type = "PLAY"
			cta.ContentSource = "LIVE"
			gi.CallToActions = append(gi.CallToActions, cta)
		}

		if ctaO := assConfig.GetCTA("RECORD"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) && now.Before(end) {
			var cta CallToAction
			var ai AndroidIntent
			ai.ContentUri = fmt.Sprintf("smartui://record_content/%s", s.PlatformId)
			ai.Action = "android.intent.action.VIEW"
			cta.AndroidIntent = ai
			cta.FromResumePoint = false
			cta.Title = ctaO.GetLabel(car.LanguageCode)
			cta.Type = "REC_CONTENT"
			cta.ContentSource = "LIVE"
			gi.CallToActions = append(gi.CallToActions, cta)
		}

		if ctaO := assConfig.GetCTA("CANCEL_RECORDING"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) && car.isRecordStatus(s.PlatformId, []string{"recording", "scheduled"}) {
			var cta CallToAction
			var ai AndroidIntent
			ai.ContentUri = fmt.Sprintf("smartui://record_content/cancel/%s", s.PlatformId)
			ai.Action = "android.intent.action.VIEW"
			cta.AndroidIntent = ai
			cta.FromResumePoint = false
			cta.Title = ctaO.GetLabel(car.LanguageCode)
			cta.Type = "REC_CANCEL"
			cta.ContentSource = "LIVE"
			gi.CallToActions = append(gi.CallToActions, cta)
		}

		if ctaO := assConfig.GetCTA("DELETE_RECORDING"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) && car.isRecordStatus(s.PlatformId, []string{"recorded"}) {
			var cta CallToAction
			var ai AndroidIntent
			ai.ContentUri = fmt.Sprintf("smartui://record_content/delete/%s", s.PlatformId)
			ai.Action = "android.intent.action.VIEW"
			cta.AndroidIntent = ai
			cta.FromResumePoint = false
			cta.Title = ctaO.GetLabel(car.LanguageCode)
			cta.Type = "REC_DELETE"
			cta.ContentSource = "LIVE"
			gi.CallToActions = append(gi.CallToActions, cta)
		}

		if ctaO := assConfig.GetCTA("RESTART"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) && now.After(end) && s.ScheduleRight.ScheduleTimeshiftRight.StartOver {
			var cta CallToAction
			var ai AndroidIntent
			ai.ContentUri = fmt.Sprintf("smartui://restart/epg/%s", s.PlatformId)
			ai.Action = "android.intent.action.VIEW"
			cta.AndroidIntent = ai
			cta.FromResumePoint = false
			cta.Title = ctaO.GetLabel(car.LanguageCode)
			cta.Type = "RESTART"
			cta.ContentSource = "LIVE"
			gi.CallToActions = append(gi.CallToActions, cta)
		}

		if ctaO := assConfig.GetCTA("OPEN_APP"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) {
			var cta CallToAction
			var ai AndroidIntent
			ai.ContentUri = fmt.Sprintf("smartui://epg/%s", s.PlatformId)
			ai.Action = "android.intent.action.VIEW"
			cta.AndroidIntent = ai
			cta.FromResumePoint = false
			cta.Title = ctaO.GetLabel(car.LanguageCode)
			cta.Type = "OPEN_APP"
			cta.ContentSource = "LIVE"
			gi.CallToActions = append(gi.CallToActions, cta)
		}

		if ctaO := assConfig.GetCTA("CUSTOM1"); ctaO != nil && ctaO.IsEnableAndValid(st.SubType) {
			var cta CallToAction
			var ai AndroidIntent
			ai.ContentUri = fmt.Sprintf("smartui://reminder/add/%s", s.PlatformId)
			ai.Action = "android.intent.action.VIEW"
			cta.AndroidIntent = ai
			cta.FromResumePoint = false
			cta.Title = ctaO.GetLabel(car.LanguageCode)
			cta.Type = "CUSTOM2"
			cta.ContentSource = "LIVE"
			gi.CallToActions = append(gi.CallToActions, cta)
		}
	}

	return gi
}
