package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lev69/schedule/lib"
)

func TestMain(m *testing.M) {
	f, _ := os.Create("service.log")
	log.Default().SetOutput(f)
	initRouter()
	code := m.Run()
	lib.ResetStorage()
	f.Close()
	os.Exit(code)
}

func TestGetUserListEmptyStorage(t *testing.T) {
	lib.ResetStorage()
	response := getUserList()
	if expected := http.StatusOK; response.Code != expected {
		t.Errorf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	if expected := "[]"; response.Body.String() != expected {
		t.Errorf("response body: expected: %q, actual: %q\n", expected, response.Body.String())
	}
}

func TestGetUserWrongTag(t *testing.T) {
	lib.ResetStorage()
	req, _ := http.NewRequest("GET", "/user?user_id=2", nil)
	response := executeRequest(req)
	if expected := http.StatusBadRequest; response.Code != expected {
		t.Errorf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	if expected := 0; response.Body.Len() != expected {
		t.Errorf("response body: expected: %d, actual: %d\n", expected, response.Body.Len())
	}
}

func TestGetUserIdEmptyStorage(t *testing.T) {
	lib.ResetStorage()
	response := getUser(2)
	if expected := http.StatusNotFound; response.Code != expected {
		t.Errorf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	if expected := 0; response.Body.Len() != 0 {
		t.Errorf("response body: expected: %d, actual: %d\n", expected, response.Body.Len())
	}
}

func TestCreateAndGetUser(t *testing.T) {
	lib.ResetStorage()

	//create users
	names := []string{"John Doe", "Vincent Vega", "John McClane", "Rick Sanchez"}
	ids := make(map[lib.UID]string, len(names))
	for _, name := range names {
		response := createUser(name)
		if expected := http.StatusOK; response.Code != expected {
			t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
		}
		var id idResult
		if err := json.Unmarshal(response.Body.Bytes(), &id); err != nil {
			t.Fatalf("parse response body: %v", err)
		}
		ids[id.Id] = name
	}

	// get the user by returned id
	for id, name := range ids {
		response := getUser(id)
		if expected := http.StatusOK; response.Code != expected {
			t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
		}
		var user lib.User
		if err := json.Unmarshal(response.Body.Bytes(), &user); err != nil {
			t.Fatalf("parse response body: %v", err)
		}
		if expected := id; user.Id != expected {
			t.Errorf("user id: expected: %v, actual: %v\n", expected, user.Id)
		}
		if expected := name; user.Name != expected {
			t.Errorf("user name: expected: %q, actual: %q\n", expected, user.Name)
		}
	}

	// get user list
	response := getUserList()
	if expected := http.StatusOK; response.Code != expected {
		t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	var userList []lib.User
	if err := json.Unmarshal(response.Body.Bytes(), &userList); err != nil {
		t.Fatalf("parse response body: %v", err)
	}
	if expected := len(ids); len(userList) != expected {
		t.Fatalf("users in list: expected: %v, actual: %v\n", expected, len(userList))
	}
	idsInList := make(map[lib.UID]bool)
	for _, user := range userList {
		idsInList[user.Id] = true
		name, ok := ids[user.Id]
		if !ok {
			t.Errorf("user id returned in list not found: %d", user.Id)
		}
		if expected := name; user.Name != expected {
			t.Errorf("user names for id(%v) are different: %q, %q\n", user.Id, expected, user.Name)
		}
	}
	if expected := len(ids); len(idsInList) != expected {
		t.Errorf("ids in list: expected: %v, actual: %v\n", expected, len(idsInList))
	}

	// get not existing user
	response = getUser(lib.UID(len(ids) + 10))
	if expected := http.StatusNotFound; response.Code != expected {
		t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	if expected := 0; response.Body.Len() != expected {
		t.Errorf("response body: expected: %d, actual: %d\n", expected, response.Body.Len())
	}
}

func TestGetMeetingListEmptyStorage(t *testing.T) {
	lib.ResetStorage()
	response := getMeetingList()
	if expected := http.StatusOK; response.Code != expected {
		t.Errorf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	if expected := "[]"; response.Body.String() != expected {
		t.Errorf("response body: expected: %q, actual: %q\n", expected, response.Body.String())
	}
}

func TestGetMeetingWrongTag(t *testing.T) {
	lib.ResetStorage()
	req, _ := http.NewRequest("GET", "/meeting?meeting_id=2", nil)
	response := executeRequest(req)
	if expected := http.StatusBadRequest; response.Code != expected {
		t.Errorf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	if expected := 0; response.Body.Len() != expected {
		t.Errorf("response body: expected: %d, actual: %d\n", expected, response.Body.Len())
	}
}

func TestGetMeetingIdEmptyStorage(t *testing.T) {
	lib.ResetStorage()
	response := getMeeting(2)
	if expected := http.StatusNotFound; response.Code != expected {
		t.Errorf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	if expected := 0; response.Body.Len() != 0 {
		t.Errorf("response body: expected: %d, actual: %d\n", expected, response.Body.Len())
	}
}

func TestCreateMeetingWithUnexistingUsers(t *testing.T) {
	lib.ResetStorage()
	start := time.Date(2025, time.June, 3, 12, 46, 13, 0, time.UTC)
	duration, _ := time.ParseDuration("2h45m")
	params := meetingParams{creator: 1, members: []lib.UID{1, 2, 3, 4}, start: start, duration: duration, period: lib.Once}
	response := createMeeting(params)
	if expected := http.StatusNotFound; response.Code != expected {
		t.Errorf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	if expected := 0; response.Body.Len() != 0 {
		t.Errorf("response body: expected: %d, actual: %d\n", expected, response.Body.Len())
	}
}

func TestCreateAndGetMeeting(t *testing.T) {
	lib.ResetStorage()

	//create users
	names := []string{"John Doe", "Vincent Vega", "John McClane", "Rick Sanchez"}
	ids := make(map[string]lib.UID, len(names))
	for _, name := range names {
		response := createUser(name)
		if expected := http.StatusOK; response.Code != expected {
			t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
		}
		var id idResult
		if err := json.Unmarshal(response.Body.Bytes(), &id); err != nil {
			t.Fatalf("parse response body: %v", err)
		}
		ids[name] = id.Id
	}

	// create meetings
	paramList := []meetingParams{
		{creator: ids[names[1]], members: []lib.UID{ids[names[1]], ids[names[2]], ids[names[3]]}, start: getTime("2022-12-31T22:00:00Z"), duration: getDuration("4h"), period: lib.Once},
		{creator: ids[names[3]], members: []lib.UID{ids[names[0]], ids[names[1]], ids[names[2]]}, start: getTime("2022-11-20T8:00:00Z"), duration: getDuration("1h"), period: lib.EveryDay},
	}
	meetingIds := make(map[lib.MeetingId]bool)
	for _, params := range paramList {
		response := createMeeting(params)
		if expected := http.StatusOK; response.Code != expected {
			t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
		}
		var id meetingIdResult
		if err := json.Unmarshal(response.Body.Bytes(), &id); err != nil {
			t.Fatalf("parse response body: %v", err)
		}
		meetingIds[id.Id] = true
	}

	// get the meetings by returned id
	for id := range meetingIds {
		response := getMeeting(id)
		if expected := http.StatusOK; response.Code != expected {
			t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
		}
		var meeting lib.Meeting
		if err := json.Unmarshal(response.Body.Bytes(), &meeting); err != nil {
			t.Fatalf("parse response body: %v", err)
		}
		if expected := id; meeting.Id != expected {
			t.Errorf("meeting id: expected: %v, actual: %v\n", expected, meeting.Id)
		}
	}

	// get meeting list
	response := getMeetingList()
	if expected := http.StatusOK; response.Code != expected {
		t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	var meetingList []lib.Meeting
	if err := json.Unmarshal(response.Body.Bytes(), &meetingList); err != nil {
		t.Fatalf("parse response body: %v", err)
	}
	if expected := len(meetingIds); len(meetingList) != expected {
		t.Fatalf("users in list: expected: %v, actual: %v\n", expected, len(meetingList))
	}
	idsInList := make(map[lib.MeetingId]bool)
	for _, meeting := range meetingList {
		idsInList[meeting.Id] = true
		if !meetingIds[meeting.Id] {
			t.Errorf("user id returned in list not found: %d", meeting.Id)
		}
	}
	if expected := len(meetingIds); len(idsInList) != expected {
		t.Errorf("meeting ids in list: expected: %v, actual: %v\n", expected, len(idsInList))
	}

	// get not existing meeting
	response = getMeeting(lib.MeetingId(len(meetingIds) + 10))
	if expected := http.StatusNotFound; response.Code != expected {
		t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	if expected := 0; response.Body.Len() != expected {
		t.Errorf("response body: expected: %d, actual: %d\n", expected, response.Body.Len())
	}
}

func TestSendResponse(t *testing.T) {
	lib.ResetStorage()

	//create users
	const (
		johnDoe     = "John Doe"
		vincentVega = "Vincent Vega"
		johnMcClane = "John McClane"
		rickSanchez = "Rick Sanchez"
	)
	names := []string{johnDoe, vincentVega, johnMcClane, rickSanchez}
	ids := make(map[string]lib.UID, len(names))
	for _, name := range names {
		response := createUser(name)
		if expected := http.StatusOK; response.Code != expected {
			t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
		}
		var id idResult
		if err := json.Unmarshal(response.Body.Bytes(), &id); err != nil {
			t.Fatalf("parse response body: %v", err)
		}
		ids[name] = id.Id
	}

	// create meetings
	paramList := []meetingParams{
		{creator: ids[vincentVega], members: []lib.UID{ids[vincentVega], ids[johnMcClane], ids[rickSanchez]}, start: getTime("2022-12-31T22:00:00Z"), duration: getDuration("4h"), period: lib.Once},
		{creator: ids[rickSanchez], members: []lib.UID{ids[johnDoe], ids[vincentVega], ids[johnMcClane]}, start: getTime("2022-11-20T8:00:00Z"), duration: getDuration("1h"), period: lib.EveryDay},
	}
	meetingIds := make([]lib.MeetingId, 0)
	for _, params := range paramList {
		response := createMeeting(params)
		if expected := http.StatusOK; response.Code != expected {
			t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
		}
		var id meetingIdResult
		if err := json.Unmarshal(response.Body.Bytes(), &id); err != nil {
			t.Fatalf("parse response body: %v", err)
		}
		meetingIds = append(meetingIds, id.Id)
	}

	// send response for 1st meeting
	response := sendPresence(ids[vincentVega], meetingIds[0], lib.Accepted)
	if expected := http.StatusOK; response.Code != expected {
		t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	// send response for 2nd meeting
	response = sendPresence(ids[johnMcClane], meetingIds[1], lib.Rejected)
	if expected := http.StatusOK; response.Code != expected {
		t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}

	// check 1st meeting
	response = getMeeting(meetingIds[0])
	if expected := http.StatusOK; response.Code != expected {
		t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	var meeting lib.Meeting
	if err := json.Unmarshal(response.Body.Bytes(), &meeting); err != nil {
		t.Fatalf("parse response body: %v", err)
	}
	for _, member := range meeting.Members {
		expected := lib.Unknown
		if member.UserId == ids[vincentVega] {
			expected = lib.Accepted
		}
		if member.Status != expected {
			t.Errorf("member presence: expected: %v, actual: %v\n", expected, member.Status)
		}
	}

	// check 2nd meeting
	response = getMeeting(meetingIds[1])
	if expected := http.StatusOK; response.Code != expected {
		t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	if err := json.Unmarshal(response.Body.Bytes(), &meeting); err != nil {
		t.Fatalf("parse response body: %v", err)
	}
	for _, member := range meeting.Members {
		expected := lib.Unknown
		if member.UserId == ids[johnMcClane] {
			expected = lib.Rejected
		}
		if member.Status != expected {
			t.Errorf("member presence: expected: %v, actual: %v\n", expected, member.Status)
		}
	}
}

func TestUserMeetings(t *testing.T) {
	lib.ResetStorage()

	//create users
	const (
		johnDoe     = "John Doe"
		vincentVega = "Vincent Vega"
		johnMcClane = "John McClane"
		rickSanchez = "Rick Sanchez"
	)
	names := []string{johnDoe, vincentVega, johnMcClane, rickSanchez}
	ids := make(map[string]lib.UID, len(names))
	for _, name := range names {
		response := createUser(name)
		if expected := http.StatusOK; response.Code != expected {
			t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
		}
		var id idResult
		if err := json.Unmarshal(response.Body.Bytes(), &id); err != nil {
			t.Fatalf("parse response body: %v", err)
		}
		ids[name] = id.Id
	}

	// create meetings
	paramList := []meetingParams{
		{creator: ids[vincentVega], members: []lib.UID{ids[vincentVega], ids[johnMcClane], ids[rickSanchez]}, start: getTime("2022-12-31T22:00:00Z"), duration: getDuration("4h"), period: lib.Once},
		{creator: ids[rickSanchez], members: []lib.UID{ids[johnDoe], ids[vincentVega], ids[johnMcClane]}, start: getTime("2022-11-20T8:00:00Z"), duration: getDuration("1h"), period: lib.EveryDay},
		{creator: ids[rickSanchez], members: []lib.UID{ids[vincentVega], ids[johnMcClane]}, start: getTime("2022-11-20T8:00:00Z"), duration: getDuration("1h"), period: lib.EveryWeek},
		{creator: ids[rickSanchez], members: []lib.UID{ids[vincentVega], ids[johnMcClane]}, start: getTime("2000-02-29T23:00:00Z"), duration: getDuration("3h"), period: lib.EveryYear},
		{creator: ids[rickSanchez], members: []lib.UID{ids[vincentVega], ids[johnMcClane]}, start: getTime("2022-05-31T23:00:00Z"), duration: getDuration("3h"), period: lib.EveryMonth},
		{creator: ids[rickSanchez], members: []lib.UID{ids[vincentVega], ids[johnMcClane]}, start: getTime("2010-12-30T15:40:00Z"), duration: getDuration("96h"), period: lib.EveryYear},
	}
	meetingIds := make([]lib.MeetingId, 0)
	for _, params := range paramList {
		response := createMeeting(params)
		if expected := http.StatusOK; response.Code != expected {
			t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
		}
		var id meetingIdResult
		if err := json.Unmarshal(response.Body.Bytes(), &id); err != nil {
			t.Fatalf("parse response body: %v", err)
		}
		meetingIds = append(meetingIds, id.Id)
	}

	// check 1st meeting
	expectNoMeeting(t, meetingIds[0], ids[johnMcClane], getTime("2020-12-31T0:00:00Z"), getDuration("24h")*365*2)
	expectNoMeeting(t, meetingIds[0], ids[johnMcClane], getTime("2023-12-31T0:00:00Z"), getDuration("24h")*365*2)
	expectMeeting(t, meetingIds[0], ids[johnMcClane], getTime("2022-12-30T0:00:00Z"), getDuration("96h"))

	// check 2nd meeting
	expectNoMeeting(t, meetingIds[1], ids[johnMcClane], getTime("2020-11-10T0:00:00Z"), getDuration("24h"))
	expectMeeting(t, meetingIds[1], ids[johnMcClane], getTime("2023-12-31T0:00:00Z"), getDuration("24h"))

	// check 3rd meeting
	expectNoMeeting(t, meetingIds[2], ids[johnMcClane], getTime("2022-11-06T7:00:00Z"), getDuration("2h"))
	expectMeeting(t, meetingIds[2], ids[johnMcClane], getTime("2022-11-20T7:00:00Z"), getDuration("2h"))
	expectMeeting(t, meetingIds[2], ids[johnMcClane], getTime("2022-12-11T7:00:00Z"), getDuration("2h"))

	// check 4th meeting
	expectMeeting(t, meetingIds[3], ids[johnMcClane], getTime("2000-03-01T1:00:00Z"), getDuration("2h"))
	expectNoMeeting(t, meetingIds[3], ids[johnMcClane], getTime("2001-03-01T1:00:00Z"), getDuration("2h"))
	expectMeeting(t, meetingIds[3], ids[johnMcClane], getTime("2004-03-01T1:00:00Z"), getDuration("2h"))

	// check 5th meeting
	expectMeeting(t, meetingIds[4], ids[johnMcClane], getTime("2022-06-01T0:30:00Z"), getDuration("30m"))
	expectNoMeeting(t, meetingIds[4], ids[johnMcClane], getTime("2022-07-01T0:30:00Z"), getDuration("30m"))
	expectMeeting(t, meetingIds[4], ids[johnMcClane], getTime("2022-08-01T0:30:00Z"), getDuration("30m"))
	expectMeeting(t, meetingIds[4], ids[johnMcClane], getTime("2022-09-01T0:30:00Z"), getDuration("30m"))
	expectMeeting(t, meetingIds[4], ids[johnMcClane], getTime("2023-01-01T0:30:00Z"), getDuration("30m"))
	expectMeeting(t, meetingIds[4], ids[johnMcClane], getTime("2023-02-01T0:30:00Z"), getDuration("30m"))
	expectNoMeeting(t, meetingIds[4], ids[johnMcClane], getTime("2023-03-01T0:30:00Z"), getDuration("30m"))

	// check 6th meeting
	expectNoMeeting(t, meetingIds[5], ids[johnMcClane], getTime("2010-01-01T12:00:00Z"), getDuration("24h"))
	expectMeeting(t, meetingIds[5], ids[johnMcClane], getTime("2011-01-01T12:00:00Z"), getDuration("24h"))
	expectMeeting(t, meetingIds[5], ids[johnMcClane], getTime("2012-01-01T12:00:00Z"), getDuration("24h"))
}

func TestFindFreeTime(t *testing.T) {
	lib.ResetStorage()

	//create users
	const (
		johnDoe     = "John Doe"
		vincentVega = "Vincent Vega"
		johnMcClane = "John McClane"
		rickSanchez = "Rick Sanchez"
	)
	names := []string{johnDoe, vincentVega, johnMcClane, rickSanchez}
	ids := make(map[string]lib.UID, len(names))
	for _, name := range names {
		response := createUser(name)
		if expected := http.StatusOK; response.Code != expected {
			t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
		}
		var id idResult
		if err := json.Unmarshal(response.Body.Bytes(), &id); err != nil {
			t.Fatalf("parse response body: %v", err)
		}
		ids[name] = id.Id
	}

	// create meetings
	paramList := []meetingParams{
		{creator: ids[johnDoe], members: []lib.UID{ids[johnDoe]}, start: getTime("2022-11-30T12:00:00Z"), duration: getDuration("5h"), period: lib.Once},
		{creator: ids[vincentVega], members: []lib.UID{ids[vincentVega]}, start: getTime("2020-03-04T18:00:00Z"), duration: getDuration("2h"), period: lib.EveryDay},
		{creator: ids[johnMcClane], members: []lib.UID{ids[johnMcClane]}, start: getTime("2022-11-02T19:30:00Z"), duration: getDuration("1h30m"), period: lib.EveryWeek},
		{creator: ids[rickSanchez], members: []lib.UID{ids[rickSanchez]}, start: getTime("2000-02-29T20:30:00Z"), duration: getDuration("45m"), period: lib.EveryDay},
		{creator: ids[rickSanchez], members: []lib.UID{ids[rickSanchez]}, start: getTime("2022-05-01T0:00:00Z"), duration: getDuration("1h"), period: lib.EveryMonth},
	}
	meetingIds := make([]lib.MeetingId, 0)
	for _, params := range paramList {
		response := createMeeting(params)
		if expected := http.StatusOK; response.Code != expected {
			t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
		}
		var id meetingIdResult
		if err := json.Unmarshal(response.Body.Bytes(), &id); err != nil {
			t.Fatalf("parse response body: %v", err)
		}
		meetingIds = append(meetingIds, id.Id)
	}

	// check time 1
	idList := make([]lib.UID, 0, len(ids))
	for _, id := range ids {
		idList = append(idList, id)
	}
	startTime := getTime("2022-11-30T09:00:00Z")
	response := findFreeTime(idList, startTime, getDuration("1h30m"))
	if expected := http.StatusOK; response.Code != expected {
		t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	var foundTime time.Time
	if err := json.Unmarshal(response.Body.Bytes(), &foundTime); err != nil {
		t.Fatalf("parse response body: %v", err)
	}
	if !foundTime.Equal(startTime) {
		t.Errorf("time not equal expected: expected: %v, actual: %v", startTime, foundTime)
	}

	// check time 2
	startTime = getTime("2022-11-30T16:00:00Z")
	response = findFreeTime(idList, startTime, getDuration("1h30m"))
	if expected := http.StatusOK; response.Code != expected {
		t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	if err := json.Unmarshal(response.Body.Bytes(), &foundTime); err != nil {
		t.Fatalf("parse response body: %v", err)
	}
	if !foundTime.Equal(getTime("2022-11-30T21:15:00Z")) {
		t.Errorf("time not equal expected: expected: %v, actual: %v", startTime, foundTime)
	}

	// check time 3
	startTime = getTime("2022-11-01T0:00:00Z")
	response = findFreeTime(idList, startTime, getDuration("23h"))
	if expected := http.StatusNotFound; response.Code != expected {
		t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
}

//
// helper functions
//

func expectMeeting(t *testing.T, meetingId lib.MeetingId, userId lib.UID, start time.Time, duration time.Duration) {
	response := getUserMeetings(userId, start, duration)
	if expected := http.StatusOK; response.Code != expected {
		t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	var meets []lib.MeetingId
	if err := json.Unmarshal(response.Body.Bytes(), &meets); err != nil {
		t.Fatalf("parse response body: %v", err)
	}
	found := false
	for _, meet := range meets {
		if meet == meetingId {
			found = true
		}
	}
	if !found {
		t.Errorf("meeting %d should be listed in period %v - %v", meetingId, start, start.Add(duration))
	}
}

func expectNoMeeting(t *testing.T, meetingId lib.MeetingId, userId lib.UID, start time.Time, duration time.Duration) {
	response := getUserMeetings(userId, start, duration)
	if expected := http.StatusOK; response.Code != expected {
		t.Fatalf("response code: expected: %d, actual: %d\n", expected, response.Code)
	}
	var meets []lib.MeetingId
	if err := json.Unmarshal(response.Body.Bytes(), &meets); err != nil {
		t.Fatalf("parse response body: %v", err)
	}
	for _, meet := range meets {
		if meet == meetingId {
			t.Errorf("meeting %d shouldn't be listed in period %v - %v", meetingId, start, start.Add(duration))
		}
	}
}

func createUser(name string) *httptest.ResponseRecorder {
	var payload bytes.Buffer
	fmt.Fprintf(&payload, "name=%s", name)
	req, _ := http.NewRequest("POST", "/user", &payload)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return executeRequest(req)
}

func getUser(id lib.UID) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", fmt.Sprintf("/user?id=%d", id), nil)
	return executeRequest(req)
}

func getUserList() *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", "/user", nil)
	return executeRequest(req)
}

type meetingParams struct {
	creator  lib.UID
	members  []lib.UID
	start    time.Time
	duration time.Duration
	period   lib.Period
}

func createMeeting(p meetingParams) *httptest.ResponseRecorder {
	var payload bytes.Buffer
	fmt.Fprintf(&payload, "creator_id=%d&member_ids=", p.creator)
	for i, id := range p.members {
		if i > 0 {
			fmt.Fprintf(&payload, ",")
		}
		fmt.Fprintf(&payload, "%d", id)
	}
	fmt.Fprintf(&payload, "&start_at=%s", p.start.Format(time.RFC3339))
	fmt.Fprintf(&payload, "&duration=%v", p.duration)
	fmt.Fprintf(&payload, "&period=%v", p.period)

	req, _ := http.NewRequest("POST", "/meeting", &payload)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return executeRequest(req)
}

func getMeeting(id lib.MeetingId) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", fmt.Sprintf("/meeting?id=%d", id), nil)
	return executeRequest(req)
}

func getMeetingList() *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", "/meeting", nil)
	return executeRequest(req)
}

func sendPresence(user lib.UID, meeting lib.MeetingId, status lib.Presence) *httptest.ResponseRecorder {
	var payload bytes.Buffer
	fmt.Fprintf(&payload, "user_id=%d&meeting_id=%d&presence=%v", user, meeting, status)
	req, _ := http.NewRequest("PUT", "/response", &payload)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return executeRequest(req)
}

func getUserMeetings(id lib.UID, start time.Time, duration time.Duration) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", fmt.Sprintf("/user_meetings?id=%d&start_at=%s&duration=%v", id, start.Format(time.RFC3339), duration), nil)
	return executeRequest(req)
}

func findFreeTime(users []lib.UID, start time.Time, duration time.Duration) *httptest.ResponseRecorder {
	var url strings.Builder
	fmt.Fprintf(&url, "/find_free_time?id=")
	for i, id := range users {
		if i > 0 {
			fmt.Fprintf(&url, ",")
		}
		fmt.Fprintf(&url, "%d", id)
	}
	fmt.Fprintf(&url, "&start_at=%s", start.Format(time.RFC3339))
	fmt.Fprintf(&url, "&duration=%v", duration)

	req, _ := http.NewRequest("GET", url.String(), nil)
	return executeRequest(req)
}

type idResult struct {
	Id lib.UID
}

type meetingIdResult struct {
	Id lib.MeetingId
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)
	return rr
}

func getTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func getDuration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	return d
}
