package absensi

import "time"

type Event struct {
	Id       string
	Location string
	Time     time.Time
	Type     EventType
}
type EventType int

const (
	In EventType = iota
	Out
)
const timeCutPoint = 8

var NullDate = time.Date(1999, time.January, 1, 0, 0, 0, 0, time.UTC)

func (evt *Event) CreateAttandanceId() AttandanceId {
	y, m, d := evt.CreateAttandanceDate().Date()
	return AttandanceId{evt.Id, time.Date(y, m, d, 0, 0, 0, 0, time.UTC)}
}
func (evt *Event) CreateAttandance() Attandance {
	id := evt.CreateAttandanceId()
	if evt.Type == In {
		return Attandance{id, evt.Time, NullDate, 0}
	} else {
		return Attandance{id, NullDate, evt.Time, 0}
	}
}
func (evt *Event) CreateAttandanceDate() time.Time {
	y, m, d := evt.Time.Date()
	if evt.Type == In {
		return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	} else {
		actualTime := evt.Time.Hour()*60 + evt.Time.Minute()
		cutTime := timeCutPoint * 60
		if actualTime < cutTime {
			return time.Date(y, m, d-1, 0, 0, 0, 0, time.UTC)
		} else {
			return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
		}
	}
}
