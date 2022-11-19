package lib

import (
	"encoding/json"
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
	return []byte(strconv.Quote(p.String())), nil
}

func (p Presence) String() string {
	s, ok := presenceToNames[p]
	if ok {
		return s
	}
	return "Unknown"
}

func (p *Presence) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*p, err = ParsePresence(s)
	return err
}

func ParsePresence(s string) (Presence, error) {
	p, ok := namesToPresence[s]
	var err error
	if !ok {
		err = fmt.Errorf("unknown value: %s: %w", s, ErrParse)
	}
	return p, err
}

type Period int

const (
	Once Period = iota
	EveryDay
	EveryWeek
	EveryMonth
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
		EveryYear:  "EveryYear",
	}
	namesToPeriod = map[string]Period{
		"Once":       Once,
		"EveryDay":   EveryDay,
		"EveryWeek":  EveryWeek,
		"EveryMonth": EveryMonth,
		"EveryYear":  EveryYear,
	}
}

func (p Period) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", p.String())), nil
}

func (p *Period) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*p, err = ParsePeriod(s)
	return err
}

func (p Period) String() string {
	return periodToNames[p]
}

func ParsePeriod(s string) (Period, error) {
	p, ok := namesToPeriod[s]
	var err error
	if !ok {
		err = fmt.Errorf("unknown value: %s: %w", s, ErrParse)
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

func (p *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	d, err := time.ParseDuration(s)
	*p = Duration{d}
	return err
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
