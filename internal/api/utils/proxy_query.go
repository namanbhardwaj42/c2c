// searchProxy.go

package utils

import (
	"bytes"
	handlers "c2c/internal/api/handlers"
	models "c2c/internal/models"
	"c2c/internal/tools"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type VersionResponse struct {
	ApiVersion string `json:"api_version"`
}

func SearchProxy(e *handlers.Env, params *models.ProxySearchQueries) (*http.Response, error) {
	base, err := url.Parse(e.Config.ProxyUrl)
	if err != nil {
		return nil, err
	}

	base.Path += "/search"

	url_params := url.Values{}
	url_params.Add("order", fmt.Sprintf("%s_%s", e.Config.Assistant.SearchConfig.OrderBy, e.Config.Assistant.SearchConfig.GroupOrder))
	url_params.Add("content_ordering", e.Config.Assistant.SearchConfig.ContentOrdering)
	url_params.Add("limit", fmt.Sprintf("%d", e.Config.Assistant.SearchConfig.Limit))
	url_params.Add("dataset", strings.Join(*e.Config.Assistant.SearchConfig.Datasets, ","))
	if e.Config.Assistant.SearchConfig.EpgOrderingStartDate {
		url_params.Add("epg_ordering", "start_asc")
	}
	if e.Config.Assistant.SearchConfig.EpgLimit > 0 {
		url_params.Add("epg_limit", fmt.Sprintf("%d", e.Config.Assistant.SearchConfig.EpgLimit))
	}
	if e.Config.Assistant.SearchConfig.VodLimit > 0 {
		url_params.Add("vod_limit", fmt.Sprintf("%d", e.Config.Assistant.SearchConfig.VodLimit))
	}
	base.RawQuery = url_params.Encode()

	tools.Logger.Infof(base.String())

	json_data, err := json.Marshal(params)
	if err != nil {
		tools.Logger.Error(err)
	}

	tools.Logger.Infof("new buffer %s", string(json_data))

	resp, err := http.Post(
		base.String(),
		"application/json",
		bytes.NewBuffer(json_data))

	return resp, err
}

func VersionProxy(e *handlers.Env) (string, error) {

	base, err := url.Parse(e.Config.ProxyUrl)
	if err != nil {
		tools.Logger.Error(err)
	}

	base.Path += "/version"

	resp, err := http.Get(base.String())
	if err != nil {
		tools.Logger.Error(err)
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		tools.Logger.Error(err)
		return "", err
	}

	defer resp.Body.Close()

	var target VersionResponse

	if err := json.Unmarshal(body, &target); err != nil {
		tools.Logger.Error(err)
		return "", err
	}

	return target.ApiVersion, nil
}