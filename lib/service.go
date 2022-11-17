package lib

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	contentTypeTag = "Content-Type"
	mimeJson       = "text/json"
	idTag          = "id"
	nameTag        = "name"
	creatorIdTag   = "creator_id"
	memberIdsTag   = "member_ids"
	startAtTag     = "start_at"
	durationTag    = "duration"
	periodTag      = "period"
	userIdTag      = "user_id"
	meetingIdTag   = "meeting_id"
	presenceTag    = "presence"
)

// general handler for /user path
func UserHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		userGetHandler(w, r)
	case http.MethodPost:
		userPostHandler(w, r)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}

// general handler for /meeting path
func MeetingHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		meetingGetHandler(w, r)
	case http.MethodPost:
		meetingPostHandler(w, r)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}

// general handler for /response path
func ResponseHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		responsePutHandler(w, r)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func UserMeetingsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		userMeetingsGetHandler(w, r)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}

// handler for GET method for /user path
func userGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	params := parameters{idTag: singleValue}
	if !checkArgs(&r.Form, params) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var id UID
	v, idSet := r.Form[idTag]
	if idSet {
		tmp, err := strconv.ParseUint(v[0], 10, 32)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id = UID(tmp)

		usr, err := userFindById(id)
		if err != nil {
			if errors.Is(err, ErrNotExist) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		result, err := json.MarshalIndent(usr, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set(contentTypeTag, mimeJson)
		w.Write(result)
		return
	}

	usrList, err := userList()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	result, err := json.MarshalIndent(usrList, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set(contentTypeTag, mimeJson)
	w.Write(result)
}

// handler for POST method for /user path
func userPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	params := parameters{nameTag: singleValue | parameterRequired}
	if !checkArgs(&r.Form, params) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := createUser(r.FormValue(nameTag))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(struct{ Id UID }{id})
}

// handler for GET method for /meeting path
func meetingGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	params := parameters{idTag: singleValue}
	if !checkArgs(&r.Form, params) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var id MeetingId
	v, idSet := r.Form[idTag]
	if idSet {
		tmp, err := strconv.ParseUint(v[0], 10, 32)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id = MeetingId(tmp)

		meet, err := meetingFindById(id)
		if err != nil {
			if errors.Is(err, ErrNotExist) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		result, err := json.MarshalIndent(meet, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set(contentTypeTag, mimeJson)
		w.Write(result)
		return
	}

	meets, err := meetingList()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	result, err := json.MarshalIndent(meets, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set(contentTypeTag, mimeJson)
	w.Write(result)
}

// handler for POST method for /meeting path
func meetingPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	params := parameters{
		creatorIdTag: singleValue | parameterRequired,
		memberIdsTag: multipleValue | parameterRequired,
		startAtTag:   singleValue | parameterRequired,
		durationTag:  singleValue | parameterRequired,
		periodTag:    singleValue,
	}
	if !checkArgs(&r.Form, params) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	creatorId, err := strconv.ParseUint(r.FormValue(creatorIdTag), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	members := make([]Participant, 0)
	for _, mId := range r.Form[memberIdsTag] {
		for _, num := range strings.Split(mId, ",") {
			id, err := strconv.ParseUint(num, 10, 32)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			members = append(members, Participant{UserId: UID(id), Status: Unknown})
		}
	}

	startAt, err := time.Parse(time.RFC3339, r.FormValue(startAtTag))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	startAt = startAt.UTC()

	dur, err := time.ParseDuration(r.FormValue(durationTag))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	duration := Duration{dur}

	repeat := Once
	if str, ok := r.Form[periodTag]; ok {
		repeat, err = ParsePeriod(str[0])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	id, err := createMeeting(UID(creatorId), members, startAt, duration, repeat)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(struct{ Id MeetingId }{id})
}

// handler for PUT method for /response path
func responsePutHandler(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	params := parameters{
		userIdTag:    singleValue | parameterRequired,
		meetingIdTag: singleValue | parameterRequired,
		presenceTag:  singleValue | parameterRequired,
	}
	if !checkArgs(&r.Form, params) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId, err := strconv.ParseUint(r.FormValue(userIdTag), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	meetingId, err := strconv.ParseUint(r.FormValue(meetingIdTag), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	presence, err := ParsePresence(r.FormValue(presenceTag))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	meeting, err := meetingFindById(MeetingId(meetingId))
	if err != nil {
		if errors.Is(err, ErrNotExist) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	userFound := false
	for i := range meeting.Members {
		if meeting.Members[i].UserId == UID(userId) {
			userFound = true
			meeting.Members[i].Status = presence
			break
		}
	}
	if !userFound {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = meetingUpdate(meeting)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// handler for GET method for /user_meetings path
func userMeetingsGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	params := parameters{
		idTag:       singleValue | parameterRequired,
		startAtTag:  singleValue | parameterRequired,
		durationTag: singleValue | parameterRequired,
	}
	if !checkArgs(&r.Form, params) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(r.FormValue(idTag), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userId := UID(id)

	startAt, err := time.Parse(time.RFC3339, r.FormValue(startAtTag))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	startAt = startAt.UTC()

	duration, err := time.ParseDuration(r.FormValue(durationTag))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	meets, err := getUserMeetings(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	meeting_ids := make([]MeetingId, 0, len(meets))
	for _, meet := range meets {
		meetingStartTime := meet.meetingStartTimeAfter(startAt)
		if !meetingStartTime.IsZero() && meetingStartTime.Before(startAt.Add(duration)) {
			meeting_ids = append(meeting_ids, meet.Id)
		}
	}

	w.Header().Set(contentTypeTag, mimeJson)
	if err = json.NewEncoder(w).Encode(meeting_ids); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func dateToTime(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func clockToTime(hour, min, sec int) time.Time {
	return time.Date(0, 0, 0, hour, min, sec, 0, time.UTC)
}

func (m Meeting) meetingStartTimeAfter(t time.Time) time.Time {
	if m.FirstOccurence.Add(m.Duration.Duration).After(t) {
		return m.FirstOccurence
	}

	switch m.Repeat {
	case Once:
		return time.Time{}
	case EveryDay:
		firstMeetingEndTime := m.FirstOccurence.Add(m.Duration.Duration)
		truncateFactor := time.Hour * 24
		distance := t.Sub(firstMeetingEndTime).Truncate(truncateFactor) + truncateFactor
		return m.FirstOccurence.Add(distance)
	case EveryWeek:
		firstMeetingEndTime := m.FirstOccurence.Add(m.Duration.Duration)
		truncateFactor := time.Hour * 24 * 7
		distance := t.Sub(firstMeetingEndTime).Truncate(truncateFactor) + truncateFactor
		return m.FirstOccurence.Add(distance)
	case EveryMonth:
		yearPeriod, monthPeriod, _ := t.Date()
		yearStart, monthStart, _ := m.FirstOccurence.Date()
		monthDistance := (yearPeriod-yearStart)*12 + int(monthPeriod-monthStart)
		if monthDistance > 0 {
			monthDistance--
		}
		meetingEndTime := m.FirstOccurence.AddDate(0, monthDistance, 0).Add(m.Duration.Duration)
		if !meetingEndTime.After(t) {
			monthDistance++
		}
		meetingStartTime := m.FirstOccurence.AddDate(0, monthDistance, 0)
		// skip month if no such day in the month or meeting ends before the search interval begins
		for (12+int(meetingStartTime.Month())-int(monthStart))%12 != ((12+monthDistance)%12) ||
			!meetingStartTime.Add(m.Duration.Duration).After(t) {
			monthDistance++
			meetingStartTime = m.FirstOccurence.AddDate(0, monthDistance, 0)
		}
		return meetingStartTime
	// case EveryDayOfWeekOfMonth:
	// 	meetingWeek := (m.FirstOccurence.Day() - 1) / 7 + 1
	// 	checkWeek := (t.Day() - 1) / 7 + 1
	// 	return meetingWeek == checkWeek && m.FirstOccurence.Weekday() == t.Weekday()
	case EveryYear:
		yearPeriod := t.Year()
		yearStart := m.FirstOccurence.Year()
		yearDistance := yearPeriod - yearStart
		if yearDistance > 0 {
			yearDistance--
		}
		meetingEndTime := m.FirstOccurence.AddDate(yearDistance, 0, 0).Add(m.Duration.Duration)
		if !meetingEndTime.After(t) {
			yearDistance++
		}
		meetingStartTime := m.FirstOccurence.AddDate(yearDistance, 0, 0)
		// skip year if no such day in the month (Feb 29 and not a leap year) or meeting ends before the search interval begins
		for (meetingStartTime.Month() != m.FirstOccurence.Month() || meetingStartTime.Day() != m.FirstOccurence.Day()) ||
			!meetingStartTime.Add(m.Duration.Duration).After(t) {
			yearDistance++
			meetingStartTime = m.FirstOccurence.AddDate(yearDistance, 0, 0)
		}
		return meetingStartTime
	default:
		return time.Time{}
	}
}

// func (m Meeting) onDate(t time.Time) bool {
// 	switch m.Repeat {
// 	case Once:
// 		meetingDate := clockToDate(m.FirstOccurence.Date())
// 		checkDate := clockToDate(t.Date())
// 		return checkDate == meetingDate
// 	case EveryDay:
// 		return m.FirstOccurence.Add(m.Duration.Duration).After(t) && m.FirstOccurence.Add(m.Duration.Duration).After(t)
// 	case EveryWeek:
// 		return m.FirstOccurence.Weekday() == t.Weekday()
// 	case EveryDayOfMonth:
// 		return m.FirstOccurence.Day() == t.Day()
// 	case EveryDayOfWeekOfMonth:
// 		meetingWeek := (m.FirstOccurence.Day()-1)/7 + 1
// 		checkWeek := (t.Day()-1)/7 + 1
// 		return meetingWeek == checkWeek && m.FirstOccurence.Weekday() == t.Weekday()
// 	case EveryYear:
// 		return m.FirstOccurence.Day() == t.Day() && m.FirstOccurence.Month() == t.Month()
// 	default:
// 		return false
// 	}
// }

// func (m Meeting) nextSinceTime(t time.Time) time.Time {
// 	switch m.Repeat {
// 	case Once:
// 		if m.FirstOccurence.Add(m.Duration.Duration).After(t) {
// 			return m.FirstOccurence
// 		}
// 		return time.Time{}
// 	case EveryDay:
// 		if clockToTime(m.FirstOccurence.Add(m.Duration.Duration).Clock()).After(clockToTime(t.Clock())) {
// 			return
// 		}
// 		// return true
// 	case EveryWeek:
// 		// return m.FirstOccurence.Weekday() == t.Weekday()
// 	case EveryDayOfMonth:
// 		// return m.FirstOccurence.Day() == t.Day()
// 	case EveryDayOfWeekOfMonth:
// 		// meetingWeek := (m.FirstOccurence.Day() - 1) / 7 + 1
// 		// checkWeek := (t.Day() - 1) / 7 + 1
// 		// return meetingWeek == checkWeek && m.FirstOccurence.Weekday() == t.Weekday()
// 	case EveryYear:
// 		// return m.FirstOccurence.Day() == t.Day() && m.FirstOccurence.Month() == t.Month()
// 	default:
// 		return time.Time{}
// 		// return false
// 	}
// }

type parameterOptions uint

const (
	emptyValue parameterOptions = 1 << iota
	singleValue
	multipleValue
	parameterRequired
)

type parameters map[string]parameterOptions

func checkArgs(args *url.Values, params parameters) bool {
	for k, opt := range params {
		val, ok := (*args)[k]
		if !ok {
			if opt&parameterRequired == 0 {
				continue
			} else {
				return false
			}
		}
		switch len(val) {
		case 0:
			if opt&emptyValue == 0 {
				return false
			}
		case 1:
			if opt&(singleValue|multipleValue) == 0 {
				return false
			}
		default:
			if opt&multipleValue == 0 {
				return false
			}
		}
	}

	for k := range *args {
		if _, ok := params[k]; !ok {
			return false
		}
	}

	return true
}
