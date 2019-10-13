package main

import (
	"fmt"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (u *urlTester) visibleMonitors(userID int) (scheds []schedule, err error) {

	query := u.db.Select(q.Or(
		q.And(q.Eq("UserID", userID), q.Eq("Private", true)),
		q.And(q.Eq("Private", false)),
	))

	err = query.Find(&scheds)
	return

}

func (u *urlTester) summary(m *tb.Message) {

	u.saveHistory(m)

	var (
		scheds  []schedule
		message string
		diff    int64
		err     error
	)

	scheds, err = u.visibleMonitors(m.Sender.ID)
	if err != nil && err != storm.ErrNotFound {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
		return
	}

	for _, sched := range scheds {
		if alreadyOnIntArray(sched.Subscriptors, m.Sender.ID) {
			diff = time.Now().Unix() - u.lastStatus[sched.ID].Timestamp
			message = fmt.Sprintf("%s*%d* - %s [%s] (%d)\n*%s* for %s\n\n", message, sched.ID, sched.Method, sched.URL, sched.ExpectedStatus, statusText(u.lastStatus[sched.ID].Status), secondsToHuman(diff))
		}
	}

	u.bot.Send(m.Sender, message, tb.NoPreview, tb.ModeMarkdown)

}

// monitors retuns the 'visible' monitors available on the system
// own monitors + public monitors defined by others
func (u *urlTester) monitors(m *tb.Message) {

	u.saveHistory(m)

	var (
		scheds  []schedule
		message string
		err     error
	)

	scheds, err = u.visibleMonitors(m.Sender.ID)
	if err != nil && err != storm.ErrNotFound {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
		return
	}

	for _, sched := range scheds {
		if sched.UserID == m.Sender.ID {
			message = fmt.Sprintf("%s%d - %s [%s] (%d) *(yours)*", message, sched.ID, sched.Method, sched.URL, sched.ExpectedStatus)
		} else {
			message = fmt.Sprintf("%s%d - %s [%s] (%d)", message, sched.ID, sched.Method, sched.URL, sched.ExpectedStatus)
		}
		if alreadyOnIntArray(sched.Subscriptors, m.Sender.ID) == true {
			message = fmt.Sprintf("%s *(subscribed)*", message)
		}
		message = fmt.Sprintf("%s\n", message)
	}

	u.bot.Send(m.Sender, message, tb.NoPreview, tb.ModeMarkdown)

}
