package models

type IronMicResponse struct {
	Results IronMicResult  `json:"results"`
	Period  ActivityPeriod `json:"period"`
}

type IronMicResult struct {
	Center   []IronMicEntry `json:"center"`
	Approach []IronMicEntry `json:"approach"`
	Tower    []IronMicEntry `json:"tower"`
	Ground   []IronMicEntry `json:"ground"`
}

type IronMicEntry struct {
	FirstName    string `json:"fname"`
	LastName     string `json:"lname"`
	Rating       int    `json:"-"`
	TotalSeconds int    `json:"totalSeconds"`
	CID          int    `json:"controller"`
}
