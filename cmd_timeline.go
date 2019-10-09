package main

import (
	"time"

	"github.com/asdine/storm/q"
)

// addTimelineEntry adds a new entry to the timeline based on the last timestamp
func (u *urlTester) addTimelineEntry(id, status int) (diff int64, err error) {

	var (
		entry timeline
	)

	entry.MonitorID = id
	entry.Status = status
	entry.Timestamp = time.Now().Unix()

	if status == statusUp {
		entry.Downtime = entry.Timestamp - u.lastStatus[id].Timestamp
		diff = entry.Downtime
	}

	err = u.db.Save(&entry)

	u.Lock()
	u.lastStatus[id] = entry
	u.Unlock()

	return
}

// getLastTimelineEntry returns the last entry or a fake one to ensure one will be created
func (u *urlTester) getLastTimelineEntry(id int) (entry timeline) {

	var err error

	// get the last timeline entry for the monitor
	var entries []timeline
	err = u.db.Select(q.Eq("MonitorID", id)).OrderBy("Timestamp").Limit(1).Reverse().Find(&entries)
	if len(entries) == 1 {
		entry = entries[0]
	}

	if err != nil {
		// create a new fake entry
		entry = timeline{
			MonitorID: id,
			Timestamp: time.Now().Unix(),
		}
	}

	return
}
