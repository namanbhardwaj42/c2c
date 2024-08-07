// assistant_config.go

package config

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

func contains(value string, slist []string) bool {
	for _, v := range slist {
		if v == value {
			return true
		}
	}
	return false
}

// Config defines expected yaml structure for conf file.
type Badge struct {
	Enable  bool   `yaml:"enable" json:"enable"`
	BadgeId string `yaml:"badge_id" json:"badge_id"`
}

type CTA struct {
	Enable        bool     `yaml:"enable" json:"enable"`
	Type          string   `yaml:"type" json:"type"`
	SearchQueries []string `yaml:"searchQueries" json:"searchQueries"`
	Labels        struct {
		EnUS string `yaml:"en_US" json:"en_US"`
		FrFR string `yaml:"fr_FR" json:"fr_FR"`
	} `yaml:"labels" json:"labels"`
}

func (c *CTA) GetLabel(locale string) string {
	if locale == "fr-FR" {
		return c.Labels.FrFR
	}

	return c.Labels.EnUS
}

func (c *CTA) IsEnableAndValid(query string) bool {
	return c.Enable &&
		(len(c.SearchQueries) == 0 ||
			contains("ALL", c.SearchQueries) ||
			contains(query, c.SearchQueries))
}

type ContentQuery struct {
	Movies  *[]string `yaml:"movies" json:"movies"`
	TVShows *[]string `yaml:"tvshows" json:"tvshows"`
}

type SearchConfig struct {
	OrderBy              string       `yaml:"order_by" json:"order_by"`
	GroupOrder           string       `yaml:"group_order" json:"group_order"`
	ContentOrdering      string       `yaml:"content_ordering" json:"content_ordering"`
	MaxDisplayCategory   int          `yaml:"max_display_category" json:"max_display_category"`
	DisplayNetwork       bool         `yaml:"display_network" json:"display_network"`
	Recency              int64        `yaml:"recency" json:"recency"`
	Timeout              int          `yaml:"timeout" json:"timeout"`
	Limit                int          `yaml:"limit" json:"limit"`
	AdultRating          *string      `yaml:"adult_rating" json:"adult_rating"`
	HidePosterImage      bool         `yaml:"hide_poster_image" json:"hide_poster_image"`
	Datasets             *[]string    `yaml:"datasets" json:"datasets"`
	EnableSearchByID     bool         `yaml:"enable_search_by_content_id" json:"enable_search_by_content_id"`
	EpgSerieGathering    bool         `yaml:"epg_serie_gathering" json:"epg_serie_gathering"`
	EpgOrderingStartDate bool         `yaml:"epg_ordering_start_asc" json:"epg_ordering_start_asc"`
	EpgLimit             int          `yaml:"epg_limit" json:"epg_limit"`
	VodLimit             int          `yaml:"vod_limit" json:"vod_limit"`
	ContentQuery         ContentQuery `yaml:"content_query" json:"content_query"`
}

func (s *SearchConfig) Init() {
	s.OrderBy = "title"
	s.GroupOrder = "asc"
	s.ContentOrdering = "vodfirst"
	s.MaxDisplayCategory = 3
	s.Recency = 5184000
	s.Timeout = 20000
	s.Limit = 50
	s.EnableSearchByID = true
	s.HidePosterImage = false
	s.EpgSerieGathering = false
	s.EpgOrderingStartDate = false
	s.EpgLimit = 25
	s.VodLimit = 25
	var defaultAdultRaging = "XXX"
	s.AdultRating = &defaultAdultRaging
	var defaultDatasets = []string{"ALL"}
	s.Datasets = &defaultDatasets
	var defaultQueryMovies = []string{"vod"}
	s.ContentQuery.Movies = &defaultQueryMovies
	var defaultQueryTVShows = []string{"vod"}
	s.ContentQuery.TVShows = &defaultQueryTVShows
}

func (s *SearchConfig) Unmarshal(data []byte) error {
	var res SearchConfig

	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}

	s.OrderBy = res.OrderBy
	s.GroupOrder = res.GroupOrder
	s.ContentOrdering = res.ContentOrdering
	s.MaxDisplayCategory = res.MaxDisplayCategory
	s.DisplayNetwork = res.DisplayNetwork
	s.EpgSerieGathering = res.EpgSerieGathering
	s.EpgOrderingStartDate = res.EpgOrderingStartDate
	s.EpgLimit = res.EpgLimit
	s.VodLimit = res.VodLimit
	s.Recency = res.Recency
	s.Timeout = res.Timeout
	s.Limit = res.Limit

	if res.AdultRating == nil {
		var defaultAdultRaging = "XXX"
		s.AdultRating = &defaultAdultRaging
	} else {
		s.AdultRating = res.AdultRating
	}

	if res.Datasets == nil || len(*res.Datasets) == 0 {
		var defaultDatasets = []string{"ALL"}
		s.Datasets = &defaultDatasets
	} else {
		s.Datasets = res.Datasets
	}

	if res.ContentQuery.Movies == nil || len(*res.ContentQuery.Movies) == 0 {
		var defaultQueryMovies = []string{"vod"}
		s.ContentQuery.Movies = &defaultQueryMovies
	} else {
		s.ContentQuery.Movies = res.Datasets
	}

	if res.ContentQuery.TVShows == nil || len(*res.ContentQuery.TVShows) == 0 {
		var defaultQueryTVShows = []string{"vod"}
		s.ContentQuery.TVShows = &defaultQueryTVShows
	} else {
		s.ContentQuery.TVShows = res.ContentQuery.TVShows
	}

	return nil
}
func (s *SearchConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var res map[string]interface{}
	if err := unmarshal(&res); err != nil {
		return err
	}

	return s.unmarshalYAML(res)
}

func (s *SearchConfig) unmarshalYAML(resI interface{}) error {
	// var res SearchConfig

	res := resI.(map[string]interface{})

	orderBy, ok := res["order_by"]
	if ok {
		s.OrderBy = orderBy.(string)
	} else {
		s.OrderBy = "title"
	}
	groupOrder, ok := res["group_order"]
	if ok {
		s.GroupOrder = groupOrder.(string)
	} else {
		s.GroupOrder = "asc"
	}
	contentOrdering, ok := res["content_ordering"]
	if ok {
		s.ContentOrdering = contentOrdering.(string)
	} else {
		s.ContentOrdering = "vodfirst"
	}
	maxDisplayCategory, ok := res["max_display_category"]
	if ok {
		s.MaxDisplayCategory = maxDisplayCategory.(int)
	} else {
		s.MaxDisplayCategory = 5
	}
	displayNetwork, ok := res["display_network"]
	if ok {
		s.DisplayNetwork = displayNetwork.(bool)
	} else {
		s.DisplayNetwork = true
	}
	recency, ok := res["recency"]
	if ok {
		switch t := recency.(type) {
		case float64:
			s.Recency = int64(recency.(float64))
		case int:
			s.Recency = int64(recency.(int))
		default:
			fmt.Printf("Can't handle %s type", t)
		}
	} else {
		s.Recency = int64(5184000)
	}
	timeout, ok := res["timeout"]
	if ok {
		s.Timeout = timeout.(int)
	} else {
		s.Timeout = 20000
	}
	limit, ok := res["limit"]
	if ok {
		s.Limit = limit.(int)
	} else {
		s.Limit = 50
	}
	enableSearchByID, ok := res["enable_search_by_content_id"]
	if ok {
		s.EnableSearchByID = enableSearchByID.(bool)
	} else {
		s.EnableSearchByID = true
	}

	epgSerieGathering, ok := res["epg_serie_gathering"]
	if ok {
		s.EpgSerieGathering = epgSerieGathering.(bool)
	} else {
		s.EpgSerieGathering = false
	}

	eppOrderingStartDate, ok := res["epg_ordering_start_asc"]
	if ok {
		s.EpgOrderingStartDate = eppOrderingStartDate.(bool)
	} else {
		s.EpgOrderingStartDate = false
	}

	epg_limit, ok := res["epg_limit"]
	if ok {
		s.EpgLimit = epg_limit.(int)
	} else {
		s.EpgLimit = s.Limit / 2
	}

	vod_limit, ok := res["vod_limit"]
	if ok {
		s.VodLimit = vod_limit.(int)
	} else {
		s.VodLimit = s.Limit / 2
	}

	hidePosterImage, ok := res["hide_poster_image"]
	if ok {
		s.HidePosterImage = hidePosterImage.(bool)
	} else {
		s.HidePosterImage = false
	}

	adultRating, ok := res["adult_rating"]
	if !ok || adultRating == nil {
		var defaultAdultRaging = "XXX"
		s.AdultRating = &defaultAdultRaging
	} else {
		adultRatingS := adultRating.(string)
		s.AdultRating = &adultRatingS
	}

	datasets, ok := res["datasets"]
	if !ok || datasets == nil || len(datasets.([]interface{})) == 0 {
		var defaultDatasets = []string{"ALL"}
		s.Datasets = &defaultDatasets
	} else {
		datasetsS := datasets.([]interface{})
		ldatasets := []string{}
		for _, d := range datasetsS {
			ldatasets = append(ldatasets, d.(string))
		}
		s.Datasets = &ldatasets
	}

	contentQueryRes, ok := res["content_query"].(map[interface{}]interface{})
	if ok {
		for key, value := range contentQueryRes {
			if key.(string) == "movies" {
				if value == nil || len(value.([]interface{})) == 0 {
					var defaultQueryMovies = []string{"vod"}
					s.ContentQuery.Movies = &defaultQueryMovies
				} else {
					moviesS := value.([]interface{})
					lmovies := []string{}
					for _, d := range moviesS {
						lmovies = append(lmovies, d.(string))
					}
					s.ContentQuery.Movies = &lmovies
				}
			} else if key.(string) == "tvshows" {
				if value == nil || len(value.([]interface{})) == 0 {
					var defaultQueryTVShows = []string{"vod"}
					s.ContentQuery.TVShows = &defaultQueryTVShows
				} else {
					tvshowsS := value.([]interface{})
					ltvshows := []string{}
					for _, d := range tvshowsS {
						ltvshows = append(ltvshows, d.(string))
					}
					s.ContentQuery.TVShows = &ltvshows
				}
			}
		}
	} else {
		var defaultQueryMovies = []string{"vod"}
		s.ContentQuery.Movies = &defaultQueryMovies
		var defaultQueryTVShows = []string{"vod"}
		s.ContentQuery.TVShows = &defaultQueryTVShows
	}

	return nil
}

type Badges struct {
	Static  []Badge `yaml:"static" json:"static"`
	Dynamic []Badge `yaml:"dynamic" json:"dynamic"`
}

type Assistant struct {
	SearchConfig SearchConfig `yaml:"search" json:"search"`

	Badges Badges `yaml:"badges" json:"badges"`

	CTAs []CTA `yaml:"ctas" json:"ctas"`
}

func (a *Assistant) Init() {
	var badge Badge
	badge.Enable = true
	badge.BadgeId = "NEW_CONTENT"
	a.Badges.Static = append(a.Badges.Static, badge)

	var cta CTA
	cta.Enable = true
	cta.Type = "PLAY"
	cta.SearchQueries = []string{"ALL"}
	cta.Labels.EnUS = "Play"
	cta.Labels.FrFR = "Démarrer"
	a.CTAs = append(a.CTAs, cta)

	var sc SearchConfig
	sc.Init()
	a.SearchConfig = sc
}

func (a *Assistant) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var res map[string]interface{}
	var err error
	if err = unmarshal(&res); err != nil {
		return err
	}

	badges, ok := res["badges"]
	a.Badges.Static = []Badge{}
	a.Badges.Dynamic = []Badge{}
	if ok && badges != nil {
		datas, err := yaml.Marshal(badges)
		if err == nil {
			_ = yaml.Unmarshal(datas, &a.Badges)
		}
	}
	if len(a.Badges.Static) == 0 && len(a.Badges.Dynamic) == 0 {
		var badge Badge
		badge.Enable = true
		badge.BadgeId = "NEW_CONTENT"
		a.Badges.Static = append(a.Badges.Static, badge)
	}

	ctas, ok := res["ctas"]
	a.CTAs = []CTA{}
	if ok && ctas != nil {
		for _, item := range ctas.([]interface{}) {
			datas, err := yaml.Marshal(item)
			if err == nil {
				var cta CTA
				err = yaml.Unmarshal(datas, &cta)
				if err == nil {
					a.CTAs = append(a.CTAs, cta)
				}
			}
		}
	}
	if len(a.CTAs) == 0 {
		var cta CTA
		cta.Enable = true
		cta.Type = "PLAY"
		cta.SearchQueries = []string{"ALL"}
		cta.Labels.EnUS = "Play"
		cta.Labels.FrFR = "Démarrer"
		a.CTAs = append(a.CTAs, cta)
	}

	searchConfigI, ok := res["search"]
	searchConfig := map[string]interface{}{}
	if ok && searchConfigI != nil {
		for k, v := range searchConfigI.(map[interface{}]interface{}) {
			searchConfig[k.(string)] = v
		}
	}
	a.SearchConfig.unmarshalYAML(searchConfig)

	return nil
}

func (a *Assistant) IsBadgeEnabled(badge_id string) bool {

	for _, b := range a.Badges.Static {
		if b.BadgeId == badge_id {
			return true
		}
	}

	return false
}

func (a *Assistant) GetCTA(cta_id string) *CTA {

	for _, c := range a.CTAs {
		if c.Type == cta_id {
			return &c
		}
	}

	return nil
}
