// convertToProxy.go

package utils

import (
	"c2c/internal/config"
	models "c2c/internal/models"
	"c2c/internal/tools"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Metatype struct {
	TypeName string
	Now      bool
	Type     string
}

var ConfigToMetatype = map[string]Metatype{
	"vod_movie":     {TypeName: "Vod", Now: false, Type: "0"},
	"epg_movie":     {TypeName: "Schedule", Now: false, Type: "0"},
	"epg_now_movie": {TypeName: "Schedule", Now: true, Type: "0"},
	"vod_serie":     {TypeName: "Vod", Now: false, Type: "1"},
	"epg_serie":     {TypeName: "Schedule", Now: false, Type: "1"},
	"epg_now_serie": {TypeName: "Schedule", Now: true, Type: "1"},
}

func getMetatypes(ass *config.Assistant, entity_type string) (bool, []Metatype) {
	metatypes := []Metatype{}
	found := true
	if Contains([]string{"TV_SHOW", "TV shows", "tv shows", "tvshows", "tv show", "TV show"}, entity_type) {
		for _, c := range *ass.SearchConfig.ContentQuery.TVShows {
			if val, ok := ConfigToMetatype[fmt.Sprintf("%s_serie", c)]; ok {
				metatypes = append(metatypes, val)
			}
		}
	} else if Contains([]string{"MOVIE", "movies"}, entity_type) {
		for _, c := range *ass.SearchConfig.ContentQuery.TVShows {
			if val, ok := ConfigToMetatype[fmt.Sprintf("%s_movie", c)]; ok {
				metatypes = append(metatypes, val)
			}
		}
	} else {
		metatypes = []Metatype{ConfigToMetatype["vod_movie"], ConfigToMetatype["epg_movie"]}
		found = false
	}
	return found, metatypes
}

func ConvertToProxyParams(car *models.CloudApiRequest, ass *config.Assistant, st *models.SearchType) (*models.ProxySearchQueries, error) {

	if json_data, erra := json.Marshal(car); erra == nil {
		tools.Logger.Infof("%s", json_data)
	}

	var err error
	err = nil
	st.SubType = car.StructuredQuery.QueryIntent
	var proxySearchParams models.ProxySearchQueries
	switch {
	case car.StructuredQuery.QueryIntent == "ENTITY_SEARCH":
		st.Type = "OperationMedia"
		if err := convertEntitySearchToProxyParams(car, &proxySearchParams, ass, st); err != nil {
			tools.Logger.Error("Parsing error ", err)
		}
	case car.StructuredQuery.QueryIntent == "SEARCH":
		st.Type = "OperationMedia"
		if err := convertSearchToProxyParams(car, &proxySearchParams, ass, st); err != nil {
			tools.Logger.Error("Parsing error ", err)
		}
	case car.StructuredQuery.QueryIntent == "PLAY":
		st.Type = "OperationMedia"
		if len(car.StructuredQuery.Entities) > 0 {
			if err := convertEntitySearchToProxyParams(car, &proxySearchParams, ass, st); err != nil {
				tools.Logger.Error("Parsing error ", err)
			}
		} else if len(car.StructuredQuery.SearchQuery) > 0 {
			if car.HasRight("live") {
				st.Type = "DirectExecution"
				if err := convertSwitchChannelToProxyParams(car, &proxySearchParams, ass, st); err != nil {
					tools.Logger.Error("Parsing error ", err)
				}
			} else {
				err = new(models.C2CError)
				err.(*models.C2CError).ErrorCode = "FEATURE_NOT_SUPPORTED"
			}
		}
	case car.StructuredQuery.QueryIntent == "PLAY_TVM":
		st.Type = "OperationMedia"
		if err := convertEntitySearchToProxyParams(car, &proxySearchParams, ass, st); err != nil {
			tools.Logger.Error("Parsing error ", err)
		}
	case car.StructuredQuery.QueryIntent == "SWITCH_CHANNEL":
		if car.HasRight("live") {
			st.Type = "DirectExecution"
			if err := convertSwitchChannelToProxyParams(car, &proxySearchParams, ass, st); err != nil {
				tools.Logger.Error("Parsing error ", err)
			}
		} else {
			err = new(models.C2CError)
			err.(*models.C2CError).ErrorCode = "FEATURE_NOT_SUPPORTED"
		}
	case car.StructuredQuery.QueryIntent == "TV_GUIDE":
		if car.HasRight("live") {
			st.Type = "DirectExecution"
			if err := convertTVGuideToProxyParams(car, &proxySearchParams, st); err != nil {
				tools.Logger.Error("Parsing error ", err)
			}
		} else {
			err = new(models.C2CError)
			err.(*models.C2CError).ErrorCode = "FEATURE_NOT_SUPPORTED"
		}
	case car.StructuredQuery.QueryIntent == "START_RECORDING":
		if car.HasRight("dvr") {
			st.Type = "DirectExecution"
			if err := convertStartRecordingToProxyParams(car, &proxySearchParams, ass, st); err != nil {
				tools.Logger.Error("Parsing error ", err)
			}
		} else {
			err = new(models.C2CError)
			err.(*models.C2CError).ErrorCode = "FEATURE_NOT_SUPPORTED"
		}
	}

	if err == nil {
		addFilteringProxyParams(car, &proxySearchParams, ass)
	}

	return &proxySearchParams, err
}

func addFilteringProxyParams(car *models.CloudApiRequest, proxySearchParams *models.ProxySearchQueries, ass *config.Assistant) error {

	for _, queries := range proxySearchParams.Queries {
		for idx, context := range queries.Context {
			if context.Metatype == "Channel" {
				if car.CustomContext.ChannelsList != nil {
					queries.Context[idx].InPlatformID = strings.Join(*car.CustomContext.ChannelsList, ",")
				}
			} else if context.Metatype == "Schedule" {
				if car.CustomContext.ChannelsList != nil {
					queries.Context[idx].ChannelsPlatformID = strings.Join(*car.CustomContext.ChannelsList, ",")
				}
				if ass.SearchConfig.EpgSerieGathering {
					bv := true
					queries.Context[idx].EpgSerieOnly = &bv
				}
			}
		}
	}

	return nil
}

func convertEntitySearchToProxyParams(car *models.CloudApiRequest, proxySearchParams *models.ProxySearchQueries, ass *config.Assistant, st *models.SearchType) error {

	for _, ent := range car.StructuredQuery.Entities {
		_, metatypes := getMetatypes(ass, ent.EntityType)

		fallback := false
		for _, metatype := range metatypes {
			var proxyQuery models.ProxySearchQuery
			var proxyContext models.ProxySearchQueryContext
			proxyContext.Metatype = metatype.TypeName
			if ent.ContentID != nil && ass.SearchConfig.EnableSearchByID && !st.CanFallback {
				proxyContext.PlatformID = *ent.ContentID
				fallback = true
			} else {
				proxyContext.Context = "title"
				proxyContext.Text = ent.Title
			}
			mtype := metatype.Type
			proxyContext.Type = &mtype
			if metatype.TypeName == "Vod" && len(*ass.SearchConfig.AdultRating) > 0 {
				proxyContext.RatingId = *ass.SearchConfig.AdultRating
			}
			if car.CustomContext.Profile != nil &&
				car.CustomContext.Profile.VodEligibility != nil &&
				len(*(car.CustomContext.Profile.VodEligibility)) > 0 {
				proxyContext.VodEligibility = car.CustomContext.Profile.VodEligibility
				proxyContext.Eligibility = car.CustomContext.Profile.VodEligibilityAsInt
			}
			if metatype.Now {
				proxyContext.AtDate = time.Now().UTC().Unix()
			}
			proxyQuery.Context = append(proxyQuery.Context, proxyContext)
			proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
		}
		st.CanFallback = fallback
	}

	return nil
}

func convertSearchToProxyParams(car *models.CloudApiRequest, proxySearchParams *models.ProxySearchQueries, ass *config.Assistant, st *models.SearchType) error {

	if len(car.StructuredQuery.MediaType) == 0 && (car.StructuredQuery.SearchQuery == "movies" || car.StructuredQuery.SearchQuery == "tv shows" || car.StructuredQuery.SearchQuery == "tv show") {
		car.StructuredQuery.MediaType = car.StructuredQuery.SearchQuery
		car.StructuredQuery.SearchQuery = ""
	}
	if len(car.StructuredQuery.MediaType) > 0 {
		_, metatypes := getMetatypes(ass, car.StructuredQuery.MediaType)

		for _, metatype := range metatypes {
			var proxyQuery models.ProxySearchQuery
			if len(car.StructuredQuery.Genre) > 0 {
				for _, g := range car.StructuredQuery.Genre {
					var proxyContext models.ProxySearchQueryContext
					proxyContext.Metatype = metatype.TypeName
					proxyContext.Context = "category"
					proxyContext.Text = g
					mtype := metatype.Type
					proxyContext.Type = &mtype
					if metatype.TypeName == "Vod" && len(*ass.SearchConfig.AdultRating) > 0 {
						proxyContext.RatingId = *ass.SearchConfig.AdultRating
					}
					if metatype.Now {
						proxyContext.AtDate = time.Now().UTC().Unix()
					}
					if car.CustomContext.Profile != nil &&
						car.CustomContext.Profile.VodEligibility != nil &&
						len(*(car.CustomContext.Profile.VodEligibility)) > 0 {
						proxyContext.VodEligibility = car.CustomContext.Profile.VodEligibility
						proxyContext.Eligibility = car.CustomContext.Profile.VodEligibilityAsInt
					}
					proxyQuery.Context = append(proxyQuery.Context, proxyContext)
				}
			}
			if len(car.StructuredQuery.Actor) > 0 {
				for _, a := range car.StructuredQuery.Actor {
					var proxyContext models.ProxySearchQueryContext
					proxyContext.Metatype = metatype.TypeName
					proxyContext.Context = "cast"
					proxyContext.Text = a.Name
					mtype := metatype.Type
					proxyContext.Type = &mtype
					if metatype.TypeName == "Vod" && len(*ass.SearchConfig.AdultRating) > 0 {
						proxyContext.RatingId = *ass.SearchConfig.AdultRating
					}
					if metatype.Now {
						proxyContext.AtDate = time.Now().UTC().Unix()
					}
					if car.CustomContext.Profile != nil &&
						car.CustomContext.Profile.VodEligibility != nil &&
						len(*(car.CustomContext.Profile.VodEligibility)) > 0 {
						proxyContext.VodEligibility = car.CustomContext.Profile.VodEligibility
						proxyContext.Eligibility = car.CustomContext.Profile.VodEligibilityAsInt
					}
					proxyQuery.Context = append(proxyQuery.Context, proxyContext)
				}
			}
			if len(car.StructuredQuery.Director) > 0 {
				for _, d := range car.StructuredQuery.Director {
					var proxyContext models.ProxySearchQueryContext
					proxyContext.Metatype = metatype.TypeName
					proxyContext.Context = "director"
					proxyContext.Text = d.Name
					mtype := metatype.Type
					proxyContext.Type = &mtype
					if metatype.TypeName == "Vod" && len(*ass.SearchConfig.AdultRating) > 0 {
						proxyContext.RatingId = *ass.SearchConfig.AdultRating
					}
					if metatype.Now {
						proxyContext.AtDate = time.Now().UTC().Unix()
					}
					if car.CustomContext.Profile != nil &&
						car.CustomContext.Profile.VodEligibility != nil &&
						len(*(car.CustomContext.Profile.VodEligibility)) > 0 {
						proxyContext.VodEligibility = car.CustomContext.Profile.VodEligibility
						proxyContext.Eligibility = car.CustomContext.Profile.VodEligibilityAsInt
					}
					proxyQuery.Context = append(proxyQuery.Context, proxyContext)
				}
			}
			if len(proxyQuery.Context) > 0 {
				proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
			}
		}
	}

	if len(proxySearchParams.Queries) == 0 {
		// ^((.*)( movies| TV shows)|(.*))(( with (.*) in))(in (.*)|(.*))$
		// r_genre_type, _ := regexp.Compile("^(.*)(movies|TV shows)(.*)")
		// r_genre, _ := regexp.Compile("^(.*) (movies|TV shows)(.*)")
		r_search, _ := regexp.Compile("^(search for |find |play |)(.*)$")
		r_type, _ := regexp.Compile("^((.*)(movies|TV shows|tv shows|tvshows|TV show|tv show))|(.*).*$")
		r_cast_and_year, _ := regexp.Compile("^(.*) with (.*)$")
		r_cast_or_year, _ := regexp.Compile("^((.*) in |in )(.*)$")

		no_search := ""
		match := r_search.FindStringSubmatch(car.StructuredQuery.SearchQuery)
		if len(match) == 3 {
			no_search = match[2]
		}
		// metatype := "Vod"
		// match = r_type.FindStringSubmatch(no_search)
		// // get type
		// if len(match) == 5 {
		// 	switch {
		// 	case match[3] == "movies":
		// 		metatype = "Vod"
		// 	case match[3] == "TV shows" || match[3] == "tv shows" || match[3] == "tvshows":
		// 		metatype = "Schedule"
		// 	}
		// }
		metatypes := []Metatype{}
		match = r_type.FindStringSubmatch(no_search)
		// get type
		if len(match) == 5 {
			_, metatypes = getMetatypes(ass, match[3])
		}
		// maybe add metatype
		for _, metatype := range metatypes {
			var proxyQuery models.ProxySearchQuery

			// get category
			if len(match) == 5 && len(match[2]) > 0 {
				var proxyContext models.ProxySearchQueryContext
				proxyContext.Text = match[2][:len(match[2])-1]
				proxyContext.Context = "category"
				proxyContext.Metatype = metatype.TypeName
				mtype := metatype.Type
				proxyContext.Type = &mtype
				if metatype.TypeName == "Vod" && len(*ass.SearchConfig.AdultRating) > 0 {
					proxyContext.RatingId = *ass.SearchConfig.AdultRating
				}
				if metatype.Now {
					proxyContext.AtDate = time.Now().UTC().Unix()
				}
				if car.CustomContext.Profile != nil &&
					car.CustomContext.Profile.VodEligibility != nil &&
					len(*(car.CustomContext.Profile.VodEligibility)) > 0 {
					proxyContext.VodEligibility = car.CustomContext.Profile.VodEligibility
					proxyContext.Eligibility = car.CustomContext.Profile.VodEligibilityAsInt
				}
				proxyQuery.Context = append(proxyQuery.Context, proxyContext)
			}

			// get title
			if len(match) == 5 && len(match[4]) > 0 {
				var proxyContext models.ProxySearchQueryContext
				proxyContext.Text = match[4]
				proxyContext.Context = "title"
				proxyContext.Metatype = metatype.TypeName
				mtype := metatype.Type
				proxyContext.Type = &mtype
				if metatype.TypeName == "Vod" && len(*ass.SearchConfig.AdultRating) > 0 {
					proxyContext.RatingId = *ass.SearchConfig.AdultRating
				}
				if metatype.Now {
					proxyContext.AtDate = time.Now().UTC().Unix()
				}
				if car.CustomContext.Profile != nil &&
					car.CustomContext.Profile.VodEligibility != nil &&
					len(*(car.CustomContext.Profile.VodEligibility)) > 0 {
					proxyContext.VodEligibility = car.CustomContext.Profile.VodEligibility
					proxyContext.Eligibility = car.CustomContext.Profile.VodEligibilityAsInt
				}
				proxyQuery.Context = append(proxyQuery.Context, proxyContext)
			}

			if match_ry := r_cast_and_year.FindStringSubmatch(car.StructuredQuery.SearchQuery); len(match_ry) == 3 {
				if match := r_cast_or_year.FindStringSubmatch(match_ry[2]); len(match) == 4 {
					if len(match[2]) > 0 {
						var proxyContext models.ProxySearchQueryContext
						proxyContext.Text = match[2]
						proxyContext.Context = "cast"
						proxyContext.Metatype = metatype.TypeName
						mtype := metatype.Type
						proxyContext.Type = &mtype
						if metatype.TypeName == "Vod" && len(*ass.SearchConfig.AdultRating) > 0 {
							proxyContext.RatingId = *ass.SearchConfig.AdultRating
						}
						if metatype.Now {
							proxyContext.AtDate = time.Now().UTC().Unix()
						}
						if car.CustomContext.Profile != nil &&
							car.CustomContext.Profile.VodEligibility != nil &&
							len(*(car.CustomContext.Profile.VodEligibility)) > 0 {
							proxyContext.VodEligibility = car.CustomContext.Profile.VodEligibility
							proxyContext.Eligibility = car.CustomContext.Profile.VodEligibilityAsInt
						}
						proxyQuery.Context = append(proxyQuery.Context, proxyContext)
					}
					intVar, err := strconv.Atoi(match[3])
					if err == nil {
						var proxyContext models.ProxySearchQueryContext
						proxyContext.VodYear = intVar
						proxyContext.Metatype = metatype.TypeName
						mtype := metatype.Type
						proxyContext.Type = &mtype
						if metatype.TypeName == "Vod" && len(*ass.SearchConfig.AdultRating) > 0 {
							proxyContext.RatingId = *ass.SearchConfig.AdultRating
						}
						if metatype.Now {
							proxyContext.AtDate = time.Now().UTC().Unix()
						}
						if car.CustomContext.Profile != nil &&
							car.CustomContext.Profile.VodEligibility != nil &&
							len(*(car.CustomContext.Profile.VodEligibility)) > 0 {
							proxyContext.VodEligibility = car.CustomContext.Profile.VodEligibility
							proxyContext.Eligibility = car.CustomContext.Profile.VodEligibilityAsInt
						}
						proxyQuery.Context = append(proxyQuery.Context, proxyContext)
					} else {
						tools.Logger.Error("Can't parse to int ", match[2])
					}
				} else {
					var proxyContext models.ProxySearchQueryContext
					proxyContext.Text = match_ry[2]
					proxyContext.Context = "cast"
					proxyContext.Metatype = metatype.TypeName
					mtype := metatype.Type
					proxyContext.Type = &mtype
					if metatype.TypeName == "Vod" && len(*ass.SearchConfig.AdultRating) > 0 {
						proxyContext.RatingId = *ass.SearchConfig.AdultRating
					}
					if metatype.Now {
						proxyContext.AtDate = time.Now().UTC().Unix()
					}
					if car.CustomContext.Profile != nil &&
						car.CustomContext.Profile.VodEligibility != nil &&
						len(*(car.CustomContext.Profile.VodEligibility)) > 0 {
						proxyContext.VodEligibility = car.CustomContext.Profile.VodEligibility
						proxyContext.Eligibility = car.CustomContext.Profile.VodEligibilityAsInt
					}
					proxyQuery.Context = append(proxyQuery.Context, proxyContext)
				}
			}

			if len(proxyQuery.Context) == 1 && proxyQuery.Context[0].Context == "category" {
				proxyQuery.Context[0].Context = "category,title,cast,directors"
			}

			// If can't detect anything, search for the title
			if len(proxyQuery.Context) == 0 {
				var proxyContext models.ProxySearchQueryContext
				proxyContext.Text = car.StructuredQuery.SearchQuery
				proxyContext.Context = "title"
				proxyContext.Metatype = metatype.TypeName
				mtype := metatype.Type
				proxyContext.Type = &mtype
				if metatype.TypeName == "Vod" && len(*ass.SearchConfig.AdultRating) > 0 {
					proxyContext.RatingId = *ass.SearchConfig.AdultRating
				}
				if metatype.Now {
					proxyContext.AtDate = time.Now().UTC().Unix()
				}
				if car.CustomContext.Profile != nil &&
					car.CustomContext.Profile.VodEligibility != nil &&
					len(*(car.CustomContext.Profile.VodEligibility)) > 0 {
					proxyContext.VodEligibility = car.CustomContext.Profile.VodEligibility
					proxyContext.Eligibility = car.CustomContext.Profile.VodEligibilityAsInt
				}
				proxyQuery.Context = append(proxyQuery.Context, proxyContext)

			}
			proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
		}
	}

	return nil
}

func convertPlayToProxyParams(car *models.CloudApiRequest, proxySearchParams *models.ProxySearchQueries) error {
	return nil
}

func convertPlayTVMToProxyParams(car *models.CloudApiRequest, proxySearchParams *models.ProxySearchQueries) error {
	return nil
}

func convertSwitchChannelToProxyParams(car *models.CloudApiRequest, proxySearchParams *models.ProxySearchQueries, ass *config.Assistant, st *models.SearchType) error {

	if len(car.StructuredQuery.ChannelID) > 0 {
		added := false
		if car.CustomContext.Profile != nil &&
			car.CustomContext.Profile.Market != nil &&
			len(*(car.CustomContext.Profile.Market)) > 0 {
			var proxyQuery models.ProxySearchQuery
			var proxyContext models.ProxySearchQueryContext
			proxyContext.Metatype = "Channel"
			proxyContext.PlatformID = car.StructuredQuery.ChannelID
			market := fmt.Sprintf("%s,100", *car.CustomContext.Profile.Market)
			proxyContext.ChannelMarket = &market
			proxyContext.Eligibility = car.CustomContext.Profile.Eligibility
			proxyQuery.Context = append(proxyQuery.Context, proxyContext)
			proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
			added = true
		}
		if car.CustomContext.Profile != nil &&
			car.CustomContext.Profile.SubMarket != nil &&
			len(*(car.CustomContext.Profile.SubMarket)) > 0 {
			var proxyQuery models.ProxySearchQuery
			var proxyContext models.ProxySearchQueryContext
			proxyContext.Metatype = "Channel"
			proxyContext.PlatformID = car.StructuredQuery.ChannelID
			proxyContext.ChannelSubMarket = car.CustomContext.Profile.SubMarket
			proxyContext.Eligibility = car.CustomContext.Profile.Eligibility
			proxyQuery.Context = append(proxyQuery.Context, proxyContext)
			proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
			added = true
		}

		if !added {
			var proxyQuery models.ProxySearchQuery
			var proxyContext models.ProxySearchQueryContext
			proxyContext.Metatype = "Channel"
			proxyContext.PlatformID = car.StructuredQuery.ChannelID
			proxyContext.Eligibility = car.CustomContext.Profile.Eligibility
			proxyQuery.Context = append(proxyQuery.Context, proxyContext)
			proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
		}
	} else if len(car.StructuredQuery.ChannelName) > 0 {
		added := false
		if car.CustomContext.Profile != nil &&
			car.CustomContext.Profile.Market != nil &&
			len(*(car.CustomContext.Profile.Market)) > 0 {
			var proxyQuery models.ProxySearchQuery
			var proxyContext models.ProxySearchQueryContext
			proxyContext.Metatype = "Channel"
			proxyContext.Text = car.StructuredQuery.ChannelName
			proxyContext.Context = "title"
			market := fmt.Sprintf("%s,100", *car.CustomContext.Profile.Market)
			proxyContext.ChannelMarket = &market
			proxyContext.Eligibility = car.CustomContext.Profile.Eligibility
			proxyQuery.Context = append(proxyQuery.Context, proxyContext)
			proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
			added = true
		}
		if car.CustomContext.Profile != nil &&
			car.CustomContext.Profile.SubMarket != nil &&
			len(*(car.CustomContext.Profile.SubMarket)) > 0 {
			var proxyQuery models.ProxySearchQuery
			var proxyContext models.ProxySearchQueryContext
			proxyContext.Metatype = "Channel"
			proxyContext.Text = car.StructuredQuery.ChannelName
			proxyContext.Context = "title"
			proxyContext.ChannelSubMarket = car.CustomContext.Profile.SubMarket
			proxyContext.Eligibility = car.CustomContext.Profile.Eligibility
			proxyQuery.Context = append(proxyQuery.Context, proxyContext)
			proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
			added = true
		}

		if !added {
			var proxyQuery models.ProxySearchQuery
			var proxyContext models.ProxySearchQueryContext
			proxyContext.Metatype = "Channel"
			proxyContext.Text = car.StructuredQuery.ChannelName
			proxyContext.Context = "title"
			proxyContext.Eligibility = car.CustomContext.Profile.Eligibility
			proxyQuery.Context = append(proxyQuery.Context, proxyContext)
			proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
		}
	} else if len(car.StructuredQuery.ChannelNumber) > 0 {
		added := false
		if car.CustomContext.Profile != nil &&
			car.CustomContext.Profile.Market != nil &&
			len(*(car.CustomContext.Profile.Market)) > 0 {
			var proxyQuery models.ProxySearchQuery
			var proxyContext models.ProxySearchQueryContext
			proxyContext.Metatype = "Channel"
			intVar, err := strconv.Atoi(car.StructuredQuery.ChannelNumber)
			if err == nil {
				proxyContext.Number = intVar
			} else {
				proxyContext.Text = car.StructuredQuery.ChannelNumber
				proxyContext.Context = "title"
			}
			market := fmt.Sprintf("%s,100", *car.CustomContext.Profile.Market)
			proxyContext.ChannelMarket = &market
			proxyContext.Eligibility = car.CustomContext.Profile.Eligibility
			proxyQuery.Context = append(proxyQuery.Context, proxyContext)
			proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
			added = true
		}
		if car.CustomContext.Profile != nil &&
			car.CustomContext.Profile.SubMarket != nil &&
			len(*(car.CustomContext.Profile.SubMarket)) > 0 {
			var proxyQuery models.ProxySearchQuery
			var proxyContext models.ProxySearchQueryContext
			proxyContext.Metatype = "Channel"
			intVar, err := strconv.Atoi(car.StructuredQuery.ChannelNumber)
			if err == nil {
				proxyContext.Number = intVar
			} else {
				proxyContext.Text = car.StructuredQuery.ChannelNumber
				proxyContext.Context = "title"
			}
			proxyContext.ChannelSubMarket = car.CustomContext.Profile.SubMarket
			proxyContext.Eligibility = car.CustomContext.Profile.Eligibility
			proxyQuery.Context = append(proxyQuery.Context, proxyContext)
			proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
			added = true
		}

		if !added {
			var proxyQuery models.ProxySearchQuery
			var proxyContext models.ProxySearchQueryContext
			proxyContext.Metatype = "Channel"
			intVar, err := strconv.Atoi(car.StructuredQuery.ChannelNumber)
			if err == nil {
				proxyContext.Number = intVar
			} else {
				proxyContext.Text = car.StructuredQuery.ChannelNumber
				proxyContext.Context = "title"
			}
			proxyContext.Eligibility = car.CustomContext.Profile.Eligibility
			proxyQuery.Context = append(proxyQuery.Context, proxyContext)
			proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
		}
	} else if len(car.StructuredQuery.SearchQuery) > 0 {
		r_n, _ := regexp.Compile("^(((tune|switch) (to |))|(play |watch ))([0-9]+)$")
		r_t, _ := regexp.Compile("^(((tune|switch) (to |))|(play |watch ))(.*)$")

		if match := r_n.FindStringSubmatch(car.StructuredQuery.SearchQuery); len(match) == 7 {
			added := false
			if car.CustomContext.Profile != nil &&
				car.CustomContext.Profile.Market != nil &&
				len(*(car.CustomContext.Profile.Market)) > 0 {
				var proxyQuery models.ProxySearchQuery
				var proxyContext models.ProxySearchQueryContext
				proxyContext.Metatype = "Channel"
				proxyContext.Number, _ = strconv.Atoi(match[6])
				market := fmt.Sprintf("%s,100", *car.CustomContext.Profile.Market)
				proxyContext.ChannelMarket = &market
				proxyContext.Eligibility = car.CustomContext.Profile.Eligibility
				proxyQuery.Context = append(proxyQuery.Context, proxyContext)
				proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
				added = true
			}
			if car.CustomContext.Profile != nil &&
				car.CustomContext.Profile.SubMarket != nil &&
				len(*(car.CustomContext.Profile.SubMarket)) > 0 {
				var proxyQuery models.ProxySearchQuery
				var proxyContext models.ProxySearchQueryContext
				proxyContext.Metatype = "Channel"
				proxyContext.Number, _ = strconv.Atoi(match[6])
				proxyContext.ChannelSubMarket = car.CustomContext.Profile.SubMarket
				proxyContext.Eligibility = car.CustomContext.Profile.Eligibility
				proxyQuery.Context = append(proxyQuery.Context, proxyContext)
				proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
				added = true
			}

			if !added {
				var proxyQuery models.ProxySearchQuery
				var proxyContext models.ProxySearchQueryContext
				proxyContext.Metatype = "Channel"
				proxyContext.Number, _ = strconv.Atoi(match[6])
				proxyContext.Eligibility = car.CustomContext.Profile.Eligibility
				proxyQuery.Context = append(proxyQuery.Context, proxyContext)
				proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)

			}
		} else if match := r_t.FindStringSubmatch(car.StructuredQuery.SearchQuery); len(match) == 7 {
			r_p, _ := regexp.Compile("^play (.*)$")
			r_name, _ := regexp.Compile("^(.*) channel$")
			metatypes := []string{"Channel"}
			if match_p := r_p.FindStringSubmatch(car.StructuredQuery.SearchQuery); len(match_p) == 2 {
				metatypes = []string{"Channel", "Schedule", "Vod"}
			}
			channelname := match[6]
			if match_channel := r_name.FindStringSubmatch(channelname); len(match_channel) == 2 {
				channelname = match_channel[1]
			}
			for _, m := range metatypes {
				if m == "Channel" {
					added := false
					if car.CustomContext.Profile != nil &&
						car.CustomContext.Profile.Market != nil &&
						len(*(car.CustomContext.Profile.Market)) > 0 {
						var proxyQuery models.ProxySearchQuery
						var proxyContext models.ProxySearchQueryContext
						proxyContext.Metatype = m
						proxyContext.Text = channelname
						proxyContext.Context = "title"
						market := fmt.Sprintf("%s,100", *car.CustomContext.Profile.Market)
						proxyContext.ChannelMarket = &market
						proxyContext.Eligibility = car.CustomContext.Profile.Eligibility
						proxyQuery.Context = append(proxyQuery.Context, proxyContext)
						proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
						added = true
					}
					if car.CustomContext.Profile != nil &&
						car.CustomContext.Profile.SubMarket != nil &&
						len(*(car.CustomContext.Profile.SubMarket)) > 0 {
						var proxyQuery models.ProxySearchQuery
						var proxyContext models.ProxySearchQueryContext
						proxyContext.Metatype = m
						proxyContext.Text = channelname
						proxyContext.Context = "title"
						proxyContext.ChannelSubMarket = car.CustomContext.Profile.SubMarket
						proxyContext.Eligibility = car.CustomContext.Profile.Eligibility
						proxyQuery.Context = append(proxyQuery.Context, proxyContext)
						proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
						added = true
					}

					if !added {
						var proxyQuery models.ProxySearchQuery
						var proxyContext models.ProxySearchQueryContext
						proxyContext.Metatype = m
						proxyContext.Text = channelname
						proxyContext.Context = "title"
						proxyContext.Eligibility = car.CustomContext.Profile.Eligibility
						proxyQuery.Context = append(proxyQuery.Context, proxyContext)
						proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)

					}
				} else if m == "Vod" {
					var proxyQuery models.ProxySearchQuery
					var proxyContext models.ProxySearchQueryContext
					proxyContext.Metatype = m
					proxyContext.Text = match[6]
					proxyContext.Context = "title"
					if len(*ass.SearchConfig.AdultRating) > 0 {
						proxyContext.RatingId = *ass.SearchConfig.AdultRating
					}
					if car.CustomContext.Profile != nil &&
						car.CustomContext.Profile.VodEligibility != nil &&
						len(*(car.CustomContext.Profile.VodEligibility)) > 0 {
						proxyContext.VodEligibility = car.CustomContext.Profile.VodEligibility
						proxyContext.Eligibility = car.CustomContext.Profile.VodEligibilityAsInt
					}
					proxyQuery.Context = append(proxyQuery.Context, proxyContext)
					proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
				}
			}
		}
	} else {
		return nil
	}

	return nil
}

func convertTVGuideToProxyParams(car *models.CloudApiRequest, proxySearchParams *models.ProxySearchQueries, st *models.SearchType) error {
	st.SubType = "TV_GUIDE_NO_CHANNEL"

	added := false

	if car.CustomContext.Profile != nil &&
		car.CustomContext.Profile.Market != nil &&
		len(*(car.CustomContext.Profile.Market)) > 0 {
		var proxyQuery models.ProxySearchQuery
		var proxyContext models.ProxySearchQueryContext
		proxyContext.Metatype = "Channel"
		if len(car.StructuredQuery.ChannelName) > 0 {
			proxyContext.Text = car.StructuredQuery.ChannelName
			proxyContext.Context = "title"
			st.SubType = "TV_GUIDE"
		} else if len(car.StructuredQuery.ChannelNumber) > 0 {
			intVar, err := strconv.Atoi(car.StructuredQuery.ChannelNumber)
			if err == nil {
				proxyContext.Number = intVar
			} else {
				proxyContext.Text = car.StructuredQuery.ChannelNumber
				proxyContext.Context = "title"
			}
		}
		market := fmt.Sprintf("%s,100", *car.CustomContext.Profile.Market)
		proxyContext.ChannelMarket = &market
		proxyQuery.Context = append(proxyQuery.Context, proxyContext)
		proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
		added = true
	}
	if car.CustomContext.Profile != nil &&
		car.CustomContext.Profile.SubMarket != nil &&
		len(*(car.CustomContext.Profile.SubMarket)) > 0 {
		var proxyQuery models.ProxySearchQuery
		var proxyContext models.ProxySearchQueryContext
		proxyContext.Metatype = "Channel"
		if len(car.StructuredQuery.ChannelName) > 0 {
			proxyContext.Text = car.StructuredQuery.ChannelName
			proxyContext.Context = "title"
			st.SubType = "TV_GUIDE"
		} else if len(car.StructuredQuery.ChannelNumber) > 0 {
			intVar, err := strconv.Atoi(car.StructuredQuery.ChannelNumber)
			if err == nil {
				proxyContext.Number = intVar
			} else {
				proxyContext.Text = car.StructuredQuery.ChannelNumber
				proxyContext.Context = "title"
			}
			st.SubType = "TV_GUIDE"
		}
		proxyContext.ChannelSubMarket = car.CustomContext.Profile.SubMarket
		proxyContext.Eligibility = car.CustomContext.Profile.Eligibility
		proxyQuery.Context = append(proxyQuery.Context, proxyContext)
		proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
		added = true
	}

	if !added {
		var proxyQuery models.ProxySearchQuery
		var proxyContext models.ProxySearchQueryContext
		proxyContext.Metatype = "Channel"
		if len(car.StructuredQuery.ChannelName) > 0 {
			proxyContext.Text = car.StructuredQuery.ChannelName
			proxyContext.Context = "title"
			st.SubType = "TV_GUIDE"
		} else if len(car.StructuredQuery.ChannelNumber) > 0 {
			intVar, err := strconv.Atoi(car.StructuredQuery.ChannelNumber)
			if err == nil {
				proxyContext.Number = intVar
			} else {
				proxyContext.Text = car.StructuredQuery.ChannelNumber
				proxyContext.Context = "title"
			}
			st.SubType = "TV_GUIDE"
		}
		proxyQuery.Context = append(proxyQuery.Context, proxyContext)
		proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
	}

	return nil
}

func convertStartRecordingToProxyParams(car *models.CloudApiRequest, proxySearchParams *models.ProxySearchQueries, ass *config.Assistant, st *models.SearchType) error {
	var proxyQuery models.ProxySearchQuery

	var proxyContext models.ProxySearchQueryContext

	if len(car.StructuredQuery.ChannelName) > 0 {
		proxyContext.Text = car.StructuredQuery.ChannelName
		proxyContext.Metatype = "Channel"
		proxyContext.Context = "title"
		st.SubType = "START_RECORDING_CHANNEL"
		proxyQuery.Context = append(proxyQuery.Context, proxyContext)
	} else if len(car.StructuredQuery.ChannelNumber) > 0 {
		intVar, err := strconv.Atoi(car.StructuredQuery.ChannelNumber)
		if err == nil {
			proxyContext.Number = intVar
		} else {
			proxyContext.Text = car.StructuredQuery.ChannelNumber
			proxyContext.Context = "title"
		}
		proxyContext.Metatype = "Channel"
		st.SubType = "START_RECORDING_CHANNEL"
		proxyQuery.Context = append(proxyQuery.Context, proxyContext)
	} else {
		st.SubType = "START_RECORDING_SCHEDULE"
		for _, ent := range car.StructuredQuery.Entities {
			ok, metatypes := getMetatypes(ass, ent.EntityType)
			if ok {
				for _, metatype := range metatypes {
					proxyContext.Metatype = metatype.TypeName
					mtype := metatype.Type
					proxyContext.Type = &mtype
					if metatype.Now {
						proxyContext.AtDate = time.Now().UTC().Unix()
					}
					if ent.ContentID != nil && ass.SearchConfig.EnableSearchByID && !st.CanFallback {
						proxyContext.PlatformID = *ent.ContentID
						st.CanFallback = true
					} else {
						proxyContext.Context = "title"
						proxyContext.Text = ent.Title
						st.CanFallback = false
					}
					proxyQuery.Context = append(proxyQuery.Context, proxyContext)
				}
			}
		}
	}

	proxySearchParams.Queries = append(proxySearchParams.Queries, proxyQuery)
	return nil
}
