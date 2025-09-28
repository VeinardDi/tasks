package pkg

import "time"

func DayStart(t time.Time) time.Time {
	lt := t.In(time.Local)
	y, m, d := lt.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}
