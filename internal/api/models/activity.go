package models

import "time"

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
