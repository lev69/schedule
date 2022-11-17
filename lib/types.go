package lib

import (
	"fmt"
	"strconv"
	"time"
)

type UID uint32

type UserInfo struct {
	Name string
}

type User struct {
	Id UID `json:"UserId"`
	UserInfo
	meetings map[MeetingId]bool
}

type Presence int

const (
	Unknown Presence = iota
	Accepted
	Rejected
)

type Participant struct {
	UserId UID
	Status Presence
}

var presenceToNames map[Presence]string
var namesToPresence map[string]Presence

func init() {
	presenceToNames = map[Presence]string{
		Unknown:  "Unknown",
		Accepted: "Accepted",
		Rejected: "Rejected",
	}
	namesToPresence = map[string]Presence{
		"Unknown":  Unknown,
		"Accepted": Accepted,
		"Rejected": Rejected,
	}
}

func (p Presence) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(presenceToNames[p])), nil
}

func ParsePresence(s string) (Presence, error) {
	p, ok := namesToPresence[s]
	var err error
	if !ok {
		err = fmt.Errorf("unknown value: %s: %v", s, ErrParse)
	}
	return p, err
}

type Period int

const (
	Once Period = iota
	EveryDay
	EveryWeek
	EveryMonth
	// EveryDayOfWeekOfMonth
	EveryYear
)

var periodToNames map[Period]string
var namesToPeriod map[string]Period

func init() {
	periodToNames = map[Period]string{
		Once:       "Once",
		EveryDay:   "EveryDay",
		EveryWeek:  "EveryWeek",
		EveryMonth: "EveryMonth",
		// EveryDayOfWeekOfMonth: "EveryDayOfWeekOfMonth",
		EveryYear: "EveryYear",
	}
	namesToPeriod = map[string]Period{
		"Once":       Once,
		"EveryDay":   EveryDay,
		"EveryWeek":  EveryWeek,
		"EveryMonth": EveryMonth,
		// "EveryDayOfWeekOfMonth": EveryDayOfWeekOfMonth,
		"EveryYear": EveryYear,
	}
}

func (p Period) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", periodToNames[p])), nil
}

func ParsePeriod(s string) (Period, error) {
	p, ok := namesToPeriod[s]
	var err error
	if !ok {
		err = fmt.Errorf("unknown value: %s: %v", s, ErrParse)
	}
	return p, err
}

type MeetingId uint32

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(d.String())), nil
}

type MeetingInfo struct {
	Members        []Participant
	CreatorId      UID
	FirstOccurence time.Time
	Duration       Duration
	Repeat         Period
}

type Meeting struct {
	Id MeetingId `json:"MeetingId"`
	MeetingInfo
}
