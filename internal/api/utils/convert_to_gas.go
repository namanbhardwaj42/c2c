// convertToGas.go

package utils

import (
	handlers "c2c/internal/api/handlers"
	"c2c/internal/config"
	models "c2c/internal/models"
	"c2c/internal/tools"

	"fmt"
	"net/http"
	"time"
)

func ConvertProxyItemsToC(gasResult *models.GasResponse,
	e *handlers.Env,
	car *models.CloudApiRequest,
	st *models.SearchType,
	proxy_items []interface{},
	status_code int) error {

	var err error
	gasResult.DREX = nil
	gasResult.OMQR = nil
	gasResult.FLEX = nil
	switch {
	case st.Type == "OperationMedia":
		if len(proxy_items) == 0 || status_code == http.StatusNoContent {
			err := models.C2CError{ErrorCode: "NO_SEARCH_RESULTS_FOUND"}
			switch {
			case st.SubType == "ENTITY_SEARCH":
				err.ErrorCode = "ENTITY_NOT_AVAILABLE_FOR_SEARCH"
			case st.SubType == "PLAY":
				err.ErrorCode = "ENTITY_NOT_AVAILABLE_FOR_PLAYBACK"
			case st.SubType == "PLAY_TVM":
				err.ErrorCode = "ENTITY_NOT_AVAILABLE_FOR_PLAYBACK"
			}
			return &err
		}

		var omqr models.OperatorMediaQueryResponse
		if err = convertToGASResult(
			proxy_items,
			car,
			&e.Config.Assistant,
			e.RatingMap.GetRatingSystem(e.Config.RatingType),
			st,
			&omqr); err != nil {
			tools.Logger.Error("Can't convert to gas: ", e)
		}
		gasResult.OMQR = &omqr
	case st.Type == "DirectExecution":
		var drex models.DirectExecution
		if err = convertToDREX(proxy_items, car, st, &drex); err != nil {
			tools.Logger.Error("Can't convert to drex: ", err)
		}
		gasResult.DREX = &drex
	}
	return err
}

func isTonight(dateTimeSpec *models.DateTimeSpec) (bool, error) {

	if len(dateTimeSpec.Point) == 0 && len(dateTimeSpec.Range.Begin) > 0 {
		// Convert to datetime
		layout := "2006-01-02T15:04:05-07:00"
		date, err := time.Parse(layout, dateTimeSpec.Range.Begin)
		if err != nil {
			return false, err
		}
		// date = date.In(loc)
		tools.Logger.Debugf("Parsing %s to %s == %d", dateTimeSpec.Range.Begin, date, date.UTC().Unix())

		// now
		now := time.Now().In(date.Location())

		// 17h creation
		tonight := time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, date.Location())
		tools.Logger.Debugf("%s is tonight (%s)? %v", date, tonight, date.After(tonight) || date == tonight)

		return date.After(tonight) || date == tonight, err
	}
	return false, nil
}

func atTime(dateTimeSpec *models.DateTimeSpec) (int64, error) {

	var err error
	if len(dateTimeSpec.Point) != 0 {
		// Convert to datetime
		layout := "2006-01-02T15:04:05-07:00"
		date, err := time.Parse(layout, dateTimeSpec.Point)
		if err != nil {
			return 0, err
		}

		return date.UTC().Unix(), nil
	} else {
		err = new(models.C2CError)
		err.(*models.C2CError).ErrorCode = "No valid Date in point"
	}
	return 0, err
}

func convertToGASResult(
	proxy_items []interface{},
	car *models.CloudApiRequest,
	assConfig *config.Assistant,
	ratingSystem *models.RatingSystem,
	st *models.SearchType,
	omqr *models.OperatorMediaQueryResponse) error {
	var gasResultList models.GasResultList
	for _, proxy_item := range proxy_items {

		var gas_item models.GasItem
		switch v := proxy_item.(type) {
		case models.EpgChannelV1:
		case models.EpgChannelV3:
		case models.EpgScheduleV1:
			epg_item := proxy_item.(models.EpgScheduleV1)
			gas_item = (&epg_item).ConvertToGasItem(car, assConfig, ratingSystem, st)
		case models.EpgScheduleV3:
			epg_item := proxy_item.(models.EpgScheduleV3)
			gas_item = (&epg_item).ConvertToGasItem(car, assConfig, ratingSystem, st)
		case models.VodContentV1:
			vod_item := proxy_item.(models.VodContentV1)
			gas_item = (&vod_item).ConvertToGasItem(car, assConfig, ratingSystem, st)
		case models.VodContentV3:
			vod_item := proxy_item.(models.VodContentV3)
			gas_item = (&vod_item).ConvertToGasItem(car, assConfig, ratingSystem, st)
		default:
			tools.Logger.Debugf("Can't handle %s type", v)

		}

		gasResultList.Items = append(gasResultList.Items, gas_item)
	}
	omqr.QueryIntent = st.SubType
	omqr.ResultList = gasResultList

	return nil
}

func convertToDREX(proxy_items []interface{}, car *models.CloudApiRequest, st *models.SearchType, drex *models.DirectExecution) error {

	var err error
	err = nil
	switch {
	case st.SubType == "SWITCH_CHANNEL" || st.SubType == "PLAY":
		if len(proxy_items) > 0 {
			switch v := proxy_items[0].(type) {
			case models.EpgChannelV1:
				// tools.Logger.Debugf("Channel v1: %s", proxy_items[0].(models.EpgChannelV1))
				drex.AndroidIntent.ContentUri = fmt.Sprintf("smartui://play_livetv/channelId=%s", proxy_items[0].(models.EpgChannelV1).PlatformID)
			case models.EpgChannelV3:
				drex.AndroidIntent.ContentUri = fmt.Sprintf("smartui://play_livetv/channelId=%s", proxy_items[0].(models.EpgChannelV3).PlatformId)
			case models.VodContentV1:
				drex.AndroidIntent.ContentUri = fmt.Sprintf("smartui://play/vod/%s", proxy_items[0].(models.VodContentV1).PlatformID)
			case models.VodContentV3:
				drex.AndroidIntent.ContentUri = fmt.Sprintf("smartui://play/vod/%s", proxy_items[0].(models.VodContentV3).PlatformId)
			default:
				tools.Logger.Debugf("Can't handle %s type", v)
			}
		} else {
			tools.Logger.Error("No contents found")
			err = new(models.C2CError)
			err.(*models.C2CError).ErrorCode = "CHANNEL_NOT_AVAILABLE"
			return err
		}
	case st.SubType == "TV_GUIDE":
		if len(proxy_items) > 0 {
			switch v := proxy_items[0].(type) {
			case models.EpgChannelV1:
				if ok, err := isTonight(&car.StructuredQuery.DateTimeSpec); err == nil && ok {
					drex.AndroidIntent.ContentUri = fmt.Sprintf("smartui://epg_menu/epg_tonight/%s", proxy_items[0].(models.EpgChannelV1).PlatformID)
				} else if ts, err := atTime(&car.StructuredQuery.DateTimeSpec); err == nil {
					drex.AndroidIntent.ContentUri = fmt.Sprintf(
						"smartui://epg_menu/epg_now/%s/%d",
						proxy_items[0].(models.EpgChannelV1).PlatformID,
						ts)
				} else {
					if err != nil {
						tools.Logger.Debugf("err %s", err)
					}
					drex.AndroidIntent.ContentUri = fmt.Sprintf("smartui://epg_menu/epg_now/%s", proxy_items[0].(models.EpgChannelV1).PlatformID)
				}
			case models.EpgChannelV3:
				if ok, err := isTonight(&car.StructuredQuery.DateTimeSpec); err == nil && ok {
					drex.AndroidIntent.ContentUri = fmt.Sprintf("smartui://epg_menu/epg_tonight/%s", proxy_items[0].(models.EpgChannelV3).PlatformId)
				} else if ts, err := atTime(&car.StructuredQuery.DateTimeSpec); err == nil {
					drex.AndroidIntent.ContentUri = fmt.Sprintf(
						"smartui://epg_menu/epg_now/%s/%d",
						proxy_items[0].(models.EpgChannelV3).PlatformId,
						ts)
				} else {
					if err != nil {
						tools.Logger.Debugf("err %s", err)
					}
					drex.AndroidIntent.ContentUri = fmt.Sprintf("smartui://epg_menu/epg_now/%s", proxy_items[0].(models.EpgChannelV3).PlatformId)
				}
			default:
				tools.Logger.Debugf("Can't handle %s type", v)
			}
		} else {
			tools.Logger.Error("No contents found")
			err = new(models.C2CError)
			err.(*models.C2CError).ErrorCode = "CHANNEL_NOT_AVAILABLE"
			return err
		}
	case st.SubType == "TV_GUIDE_NO_CHANNEL":
		drex.AndroidIntent.ContentUri = "smartui://epg_menu/epg_now"
	case st.SubType == "START_RECORDING_CHANNEL":
		if len(proxy_items) > 0 {
			switch v := proxy_items[0].(type) {
			case models.EpgChannelV1:
				// tools.Logger.Debugf("Channel v1: %s", proxy_items[0].(models.EpgChannelV1))
				drex.AndroidIntent.ContentUri = fmt.Sprintf("smartui://record/%s", proxy_items[0].(models.EpgChannelV1).PlatformID)
			case models.EpgChannelV3:
				// tools.Logger.Debugf("Channel v3: %s", proxy_items[0].(models.EpgChannelV3))
				drex.AndroidIntent.ContentUri = fmt.Sprintf("smartui://record/%s", proxy_items[0].(models.EpgChannelV3).PlatformId)
			default:
				tools.Logger.Debugf("Can't handle %s type", v)
			}
		} else {
			tools.Logger.Error("No contents found")
			err = new(models.C2CError)
			err.(*models.C2CError).ErrorCode = "ENTITY_NOT_AVAILABLE_FOR_RECORDING"
			return err
		}
	case st.SubType == "START_RECORDING_SCHEDULE":
		if len(proxy_items) > 0 {
			switch v := proxy_items[0].(type) {
			case models.EpgScheduleV1:
				tools.Logger.Debugf("Schedule v1: %s", proxy_items[0].(models.EpgScheduleV1))
				drex.AndroidIntent.ContentUri = fmt.Sprintf("smartui://record_content/%s", proxy_items[0].(models.EpgScheduleV1).PlatformID)
			case models.EpgScheduleV3:
				tools.Logger.Debugf("Schedule v3: %s", proxy_items[0].(models.EpgScheduleV3))
			default:
				tools.Logger.Debugf("Can't handle %s type", v)
			}
		} else {
			tools.Logger.Error("No contents found")
			err = new(models.C2CError)
			err.(*models.C2CError).ErrorCode = "ENTITY_NOT_AVAILABLE_FOR_RECORDING"
			return err
		}
	}
	drex.AndroidIntent.Action = "android.intent.action.VIEW"

	return nil
}
