// proxyParser.go

package utils

import (
	handlers "c2c/internal/api/handlers"
	models "c2c/internal/models"
	"c2c/internal/tools"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func ParseRespProxy(e *handlers.Env, resp *http.Response) ([]interface{}, int, error) {

	if resp.StatusCode == http.StatusOK {

		body, err := ioutil.ReadAll(resp.Body)

		defer resp.Body.Close()

		if err != nil {
			return nil, resp.StatusCode, err
		}

		var items []interface{}

		var search_resp models.SearchResponse
		if err := json.Unmarshal(body, &search_resp); err != nil {
			tools.Logger.Errorf("Resp %s", err)
			return nil, resp.StatusCode, err
		}

		if e.ProxyVersion == 1 {
			if err := parseRestProxyV1(search_resp.ContentsV1, &items); err != nil {
				tools.Logger.Errorf("%s", err)
				return nil, resp.StatusCode, err
			}
		} else if e.ProxyVersion == 3 {
			if err := parseRestProxyV3(search_resp.ContentsV1, &items); err != nil {
				tools.Logger.Errorf("%s", err)
				return nil, resp.StatusCode, err
			}
		} else {
			tools.Logger.Error("No proxy version available")
		}

		if len(items) == 0 {
			tools.Logger.Error("No contents found")
		}

		return items, resp.StatusCode, nil
	} else if resp.StatusCode == http.StatusNoContent {
		var items []interface{}
		return items, resp.StatusCode, nil
	}

	return nil, resp.StatusCode, nil
}

func parseRestProxyV1(contents *[]interface{}, items *[]interface{}) error {
	for _, item := range *contents {
		item_as_json, _ := json.Marshal(item)
		var item_as_content models.Content
		if err := json.Unmarshal(item_as_json, &item_as_content); err != nil {
			tools.Logger.Error("test %s", err)
		}

		switch {
		case item_as_content.Metatype == "Channel":
			var item_as_channel models.EpgChannelV1
			if err := json.Unmarshal(item_as_json, &item_as_channel); err != nil {
				tools.Logger.Error("Channel parsing %s", err)
			} else {
				*items = append(*items, item_as_channel)
			}
		case item_as_content.Metatype == "Schedule":
			var item_as_schedule models.EpgScheduleV1
			if err := json.Unmarshal(item_as_json, &item_as_schedule); err != nil {
				tools.Logger.Error("Schedule parsing %s", err)
			} else {
				*items = append(*items, item_as_schedule)
			}
		case item_as_content.Metatype == "Vod":
			var item_as_vod models.VodContentV1
			if err := json.Unmarshal(item_as_json, &item_as_vod); err != nil {
				tools.Logger.Error("Vod %s", err)
			} else {
				*items = append(*items, item_as_vod)
			}
		default:
			tools.Logger.Error("Not a valid content metatype %s", item_as_content.Metatype)
		}
	}
	return nil
}

func parseRestProxyV3(contents *[]interface{}, items *[]interface{}) error {
	for _, item := range *contents {
		item_as_json, _ := json.Marshal(item)
		var item_as_content models.Content
		if err := json.Unmarshal(item_as_json, &item_as_content); err != nil {
			tools.Logger.Error("test %s", err)
		}

		switch {
		case item_as_content.Metatype == "Channel":
			var item_as_channel models.EpgChannelV3
			if err := json.Unmarshal(item_as_json, &item_as_channel); err != nil {
				tools.Logger.Error("Channel parsing %s", err)
			} else {
				*items = append(*items, item_as_channel)
			}
		case item_as_content.Metatype == "Schedule":
			var item_as_schedule models.EpgScheduleV3
			if err := json.Unmarshal(item_as_json, &item_as_schedule); err != nil {
				tools.Logger.Error("Schedule parsing %s", err)
			} else {
				*items = append(*items, item_as_schedule)
			}
		case item_as_content.Metatype == "Vod":
			var item_as_vod models.VodContentV3
			if err := json.Unmarshal(item_as_json, &item_as_vod); err != nil {
				tools.Logger.Error("Vod %s", err)
			} else {
				*items = append(*items, item_as_vod)
			}
		}
	}
	return nil
}