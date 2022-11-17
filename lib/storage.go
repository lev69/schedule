package lib

import (
	"sync"
	"time"
)

type userStorage struct {
	sync.Mutex
	m     map[UID]User
	maxId UID
}

var users = userStorage{
	m: make(map[UID]User),
}

// returns list of users registered in the system
func userList() ([]User, error) {
	users.Lock()
	defer users.Unlock()
	values := make([]User, 0, len(users.m))
	for _, usr := range users.m {
		values = append(values, usr)
	}
	return values, nil
}

// looks up the user by given Id.
// Possible errors:
//
//	ErrNotExist User with given Id is not found.
func userFindById(id UID) (User, error) {
	users.Lock()
	defer users.Unlock()
	var e error
	usr, ok := users.m[id]
	if !ok {
		e = ErrNotExist
	}
	return usr, e
}

// creates user and returns theirs id.
// This implementation does not return errors
func userAdd(u UserInfo) (UID, error) {
	users.Lock()
	defer users.Unlock()
	users.maxId++
	id := users.maxId
	users.m[id] = User{id, u, map[MeetingId]bool{}}
	return id, nil
}

func createUser(name string) (UID, error) {
	return userAdd(UserInfo{name})
}

type meetingStorage struct {
	sync.Mutex
	m     map[MeetingId]Meeting
	maxId MeetingId
}

var meetings = meetingStorage{
	m: make(map[MeetingId]Meeting),
}

// returns list of scheduled meetings
func meetingList() ([]Meeting, error) {
	meetings.Lock()
	defer meetings.Unlock()
	values := make([]Meeting, 0, len(meetings.m))
	for _, m := range meetings.m {
		values = append(values, m)
	}
	return values, nil
}

// looks up the meeting by given Id.
// Possible errors:
//
//	ErrNotExist Meeting with given Id is not found.
func meetingFindById(id MeetingId) (Meeting, error) {
	meetings.Lock()
	defer meetings.Unlock()
	var e error
	m, ok := meetings.m[id]
	if !ok {
		e = ErrNotExist
	}
	return m, e
}

func meetingAdd(m MeetingInfo) (MeetingId, error) {
	users.Lock()
	defer users.Unlock()
	meetings.Lock()
	defer meetings.Unlock()
	meetings.maxId++
	id := meetings.maxId
	meetings.m[id] = Meeting{id, m}
	for _, member := range m.Members {
		if _, ok := users.m[member.UserId]; ok {
			users.m[member.UserId].meetings[id] = true
		}
	}
	return id, nil
}

func meetingUpdate(m Meeting) error {
	meetings.Lock()
	defer meetings.Unlock()
	if _, ok := meetings.m[m.Id]; !ok {
		return ErrNotExist
	}
	meetings.m[m.Id] = m
	// TODO: remove meetings from the User.meetings map if the Meeting.Members has been changed
	return nil
}

func getUserMeetings(id UID) ([]Meeting, error) {
	users.Lock()
	defer users.Unlock()
	meetings.Lock()
	defer meetings.Unlock()
	meetIds := users.m[id].meetings
	meets := make([]Meeting, 0, len(meetIds))
	for meetId := range meetIds {
		meets = append(meets, meetings.m[meetId])
	}
	return meets, nil
}

func createMeeting(creator UID, members []Participant, startAt time.Time, duration Duration, repeat Period) (MeetingId, error) {
	return meetingAdd(MeetingInfo{CreatorId: creator, Members: members, FirstOccurence: startAt, Duration: duration, Repeat: repeat})
}
