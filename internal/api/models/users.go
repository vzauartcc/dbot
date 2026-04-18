package models

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
