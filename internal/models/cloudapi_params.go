// requests.go

package models

import (
	"encoding/json"
	"strings"
)

type DeviceConfig struct {
	DeviceModelId string `json:"deviceModelId"`
}

type CARPerson struct {
	Name  string `json:"name"`
	TmsId string `json:"tmsId"`
}

type Entity struct {
	ContentID  *string `json:"contentId,omitempty"`
	TmsRootID  *string `json:"tmsRootId,omitempty"`
	Title      string  `json:"title"`
	EntityType string  `json:"entityType"`
}

type Profile struct {
	Market              *string `json:"market,omitempty"`
	SubMarket           *string `json:"subMarket,omitempty"`
	VodEligibility      *string `json:"eligibility,omitempty"`
	VodEligibilityAsInt *int    `json:"eligibility,omitempty"`
	Eligibility         *int    `json:"tv_eligibility,omitempty"`
}

type ProfileJson struct {
	Market         *string `json:"market,omitempty"`
	SubMarket      *string `json:"subMarket,omitempty"`
	VodEligibility *string `json:"eligibility,omitempty"`
	Eligibility    *string `json:"tv_eligibility,omitempty"`
}

func (p *Profile) UnmarshalJSON(data []byte) error {
	var res ProfileJson

	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}

	if res.Market != nil {
		p.Market = res.Market
	}

	if res.SubMarket != nil {
		p.SubMarket = res.SubMarket
	}

	if res.VodEligibility != nil {
		p.VodEligibility = res.VodEligibility
		var eligibility int = 0
		values := strings.Split(*res.VodEligibility, "|")
		if contains("OTT", values) {
			eligibility |= 1 // OTT enum value
		}
		if contains("IPTV", values) {
			eligibility |= 2 // IPTV enum value
		}
		if contains("IPTV_HD", values) {
			eligibility |= 4 // IPTV_HD enum value
		}
		if contains("IPTV_SD", values) {
			eligibility |= 8 // IPTV_SD enum value
		}
		p.VodEligibilityAsInt = &eligibility
	}

	if res.Eligibility != nil {
		var eligibility int = 0
		values := strings.Split(*res.Eligibility, "|")
		if contains("OTT", values) {
			eligibility |= 1 // OTT enum value
		}
		if contains("IPTV", values) {
			eligibility |= 2 // IPTV enum value
		}
		if contains("IPTV_HD", values) {
			eligibility |= 4 // IPTV_HD enum value
		}
		if contains("IPTV_SD", values) {
			eligibility |= 8 // IPTV_SD enum value
		}
		p.Eligibility = &eligibility
	}

	return nil
}

type CustomContext struct {
	Emi struct {
		Population string `json:"population"`
	} `json:"emi"`

	Profile *Profile `json:"profile,omitempty"`

	Services *struct {
		Live bool `json:"live"`
		Vod  bool `json:"vod"`
		Dvr  bool `json:"dvr"`
	} `json:"services,omitempty"`

	Entitlements *struct {
		TVods            []string `json:"tvods"`
		SubscriptionsIds []string `json:"subscriptions_ids"`
	} `json:"entitlements,omitempty"`

	Recordings []struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"recordings"`

	CurrentChannel *string   `json:"current_channel,omitempty"`
	ChannelsList   *[]string `json:"channels_list"`
}

type DateTimeSpec struct {
	Point string `json:"point,omitempty"`
	Range struct {
		Begin string `json:"begin,omitempty"`
		End   string `json:"string,omitempty"`
	} `json:"range,omitempty"`
}

type StructuredQuery struct {
	QueryIntent   string       `json:"queryIntent"`
	SearchQuery   string       `json:"searchQuery,omitempty"`
	MediaType     string       `json:"mediaType"`
	Genre         []string     `json:"genre,omitempty"`
	Actor         []CARPerson  `json:"actor,omitempty"`
	Director      []CARPerson  `json:"director,omitempty"`
	Entities      []Entity     `json:"entities,omitempty"`
	ChannelName   string       `json:"channelName,omitempty"`
	ChannelID     string       `json:"channelId,omitempty"`
	ChannelNumber string       `json:"channelNumber,omitempty"`
	DateTimeSpec  DateTimeSpec `json:"dateTimeSpec,omitempty"`
}

type CloudApiRequest struct {
	RequestId         string          `json:"requestId"`
	LanguageCode      string          `json:"languageCode"`
	DeviceConfig      DeviceConfig    `json:"deviceConfig"`
	CustomContext     *CustomContext  `json:"-"`
	CustomContextFull string          `json:"customContext,omitempty"`
	StructuredQuery   StructuredQuery `json:"structuredQuery"`
}

func (car *CloudApiRequest) HasRight(right string) bool {
	if car.CustomContext == nil || car.CustomContext.Services == nil {
		return true
	}

	switch {
	case right == "live":
		return car.CustomContext.Services.Live
	case right == "vod":
		return car.CustomContext.Services.Vod
	case right == "dvr":
		return car.CustomContext.Services.Dvr
	}
	return false
}

func (car *CloudApiRequest) isEntitled(contentID string) bool {
	if car.CustomContext == nil || car.CustomContext.Entitlements == nil {
		return false
	}
	return contains(contentID, car.CustomContext.Entitlements.TVods)
}

func (car *CloudApiRequest) isSubscribed(subscriptionsIds []string) bool {
	if car.CustomContext == nil || car.CustomContext.Entitlements == nil {
		return false
	}

	for _, s := range car.CustomContext.Entitlements.SubscriptionsIds {
		for _, si := range subscriptionsIds {
			if s == si {
				return true
			}
		}
	}
	return false
}

func (car *CloudApiRequest) isRecordStatus(contentID string, status []string) bool {
	if car.CustomContext == nil {
		return false
	}
	for _, r := range car.CustomContext.Recordings {
		if r.ID == contentID && contains(r.Status, status) {
			return true
		}
	}

	return false
}
