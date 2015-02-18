package absensi

import "time"

type Attandance struct {
	Id        AttandanceId
	TimeIn    time.Time
	TimeOut   time.Time
	WorkHours float64
}
type AttandanceId struct {
	EmployeeId string
	Date       time.Time
}

func (atd *Attandance) Calculate(evt Event) {
	if evt.Type == In {
		if evt.Time.Before(atd.TimeIn) {
			atd.TimeIn = evt.Time
		}

	} else {
		if evt.Time.After(atd.TimeOut) {
			atd.TimeOut = evt.Time
		}

	}
	if atd.TimeIn != NullDate && atd.TimeOut != NullDate {
		atd.WorkHours = atd.TimeOut.Sub(atd.TimeIn).Hours()
	}
}
func (id *AttandanceId) ToKey() string {
	tgl := id.Date.Format("2006-01-02")
	return id.EmployeeId + "_" + tgl
}
