// search_response.go

package models

type Pages struct {
	Current         int `json:"current"`
	Total           int `json:"total"`
	ItemsCount      int `json:"items_count"`
	TotalItemsCount int `json:"total_items_count"`
}

type Build struct {
	Version string `json:"version"`
}

type SearchResponse struct {
	ApiVersion string         `json:"api_version"`
	Build      Build          `json:"build"`
	Pages      Pages          `json:"pages"`
	ContentsV1 *[]interface{} `json:"contents"`
	ContentsV3 *[]interface{} `json:"Results"`
}

type AndroidIntent struct {
	ContentUri  string `json:"contentUri"`
	Action      string `json:"action"`
	PackageName string `json:"packageName,omitempty"`
}

type CallToAction struct {
	Title           string        `json:"title"`
	Type            string        `json:"type"`
	AndroidIntent   AndroidIntent `json:"androidIntent"`
	ContentSource   string        `json:"contentSource"`
	FromResumePoint bool          `json:"fromResumePoint"`
}

type Badging struct {
	StaticBadge struct {
		BadgeID string `json:"badgeId"`
	} `json:"staticBadge,omitempty"`
	RuntimeBadge *struct {
		BadgeID string `json:"badgeId"`
	} `json:"runtimeBadge,omitempty"`
	Type string `json:"type"`
}

type GasPicture struct {
	ImageUrl string `json:"imageUrl"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

type GasItem struct {
	Title            string         `json:"title"`
	Genre            string         `json:"genre"`
	ID               string         `json:"id"`
	TMSID            string         `json:"tmsId,omitempty"`
	Description      string         `json:"description"`
	Rating           string         `json:"rating"`
	Type             string         `json:"entityType"`
	Network          string         `json:"network,omitempty"`
	ReleaseStartYear string         `json:"releaseStartYear,omitempty"`
	ReleaseEndYear   string         `json:"releaseEndYear,omitempty"`
	Duration         int32          `json:"durationSecs"`
	PosterImage      GasPicture     `json:"posterImage"`
	CallToActions    []CallToAction `json:"callToActions"`
	Badgings         []Badging      `json:"badgings"`
}

type GasResultList struct {
	Items []GasItem `json:"items"`
}

type DirectExecution struct {
	AndroidIntent struct {
		ContentUri string `json:"contentUri"`
		Action     string `json:"action"`
	} `json:"androidIntent"`
}

type OperatorMediaQueryResponse struct {
	QueryIntent string        `json:"queryIntent"`
	ResultList  GasResultList `json:"resultList"`
}

type FallbackExecution struct {
	ErrorCode string `json:"error_code"`
}

func (f *FallbackExecution) fill(err error) {
	switch e := err.(type) {
	case *C2CError:
		f.ErrorCode = e.ErrorCode
	default:
		f.ErrorCode = "OPERATOR_INTERNAL_ERROR"
	}
}

type GasResponse struct {
	DREX *DirectExecution            `json:"directExecution,omitempty"`
	OMQR *OperatorMediaQueryResponse `json:"operatorMediaQueryResponse,omitempty"`
	FLEX *FallbackExecution          `json:"fallbackExecution,omitempty"`
}

func (g *GasResponse) SetError(err error) {
	var gasError FallbackExecution
	gasError.fill(err)

	g.DREX = nil
	g.OMQR = nil
	g.FLEX = &gasError
}
