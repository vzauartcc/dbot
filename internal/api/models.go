package zauapi

import "time"

type User struct {
	CID          int      `json:"cid"`
	FirstName    string   `json:"fname"`
	LastName     string   `json:"lname"`
	DiscordID    string   `json:"discord"`
	Rating       int      `json:"rating"`
	Roles        []string `json:"roleCodes"`
	IsVisitor    bool     `json:"vis"`
	IsMember     bool     `json:"member"`
	HomeFacility string   `json:"homeFacility"`
	CertCodes    []string `json:"certCodes"`
}

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

type ActivityPeriod struct {
	Unit                 string    `json:"unit"`
	PeriodsInYear        int       `json:"periodsInYear"`
	PeriodLenghtInMonths int       `json:"periodLength"`
	CurrentPeriod        int       `json:"currentPeriod"`
	StartOfCurrentPeriod time.Time `json:"startOfCurrent"`
	EndOfCurrentPeriod   time.Time `json:"endOfCurrent"`
}

type OnlineData struct {
	Pilots      any                `json:"-"`
	Controllers []OnlineController `json:"atc"`
}

type OnlineController struct {
	CID       int       `json:"cid"`
	Name      string    `json:"name"`
	Rating    int       `json:"-"`
	Position  string    `json:"pos"`
	LogonTime time.Time `json:"timeStart"`
	Atis      *string   `json:"-"`
	Frequency string    `json:"-"`

	RatingShort string `json:"-"`
	RatingLong  string `json:"-"`
}

type Staff struct {
	ATM  StaffPosition `json:"atm"`
	DATM StaffPosition `json:"datm"`
	TA   StaffPosition `json:"ta"`
	EC   StaffPosition `json:"ec"`
	FE   StaffPosition `json:"fe"`
	WM   StaffPosition `json:"wm"`
	Ins  StaffPosition `json:"-"`
	Mtr  StaffPosition `json:"-"`
	IA   StaffPosition `json:"-"`
}

type StaffPosition struct {
	Title string `json:"-"`
	Code  string `json:"-"`
	Users []User `json:"users"`
}
