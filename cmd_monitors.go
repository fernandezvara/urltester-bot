package main

import (
	"fmt"
	"log"
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

	if !m.Private() {
		return
	}
	u.saveHistory(m)

	var (
		scheds  []schedule
		status  string
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
			status = statusText(u.lastStatus[sched.ID].Status)
			message = fmt.Sprintf("%s%d - %s %s (%d)\n%s for %s\n", message, sched.ID, sched.Method, sched.URL, sched.ExpectedStatus, status, secondsToHuman(diff))
		}

		message = fmt.Sprintf("%s\n", message)
	}

	u.bot.Send(m.Sender, message, tb.NoPreview)

}

// monitors retuns the 'visible' monitors available on the system
// own monitors + public monitors defined by others
func (u *urlTester) monitors(m *tb.Message) {

	if !m.Private() {
		return
	}
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
			message = fmt.Sprintf("%s%d - %s %s (%d) (yours)", message, sched.ID, sched.Method, sched.URL, sched.ExpectedStatus)
		} else {
			message = fmt.Sprintf("%s%d - %s %s (%d)", message, sched.ID, sched.Method, sched.URL, sched.ExpectedStatus)
		}
		if alreadyOnIntArray(sched.Subscriptors, m.Sender.ID) == true {
			message = fmt.Sprintf("%s (subscribed)", message)
		}
		message = fmt.Sprintf("%s\n", message)
	}

	u.bot.Send(m.Sender, message, tb.NoPreview)

}

func (u *urlTester) newmonitor(m *tb.Message) {

	var (
		sched schedule
	)

	if !m.Private() {
		return
	}
	u.saveHistory(m)

	// verify format
	method, url, interval, private, expectedStatus, err := u.cleanPayload(m.Payload, true)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
		return
	}

	// monitor exists? look for the private urls of the current user or public ones
	query := u.db.Select(q.Or(
		q.And(q.Eq("UserID", m.Sender.ID), q.Eq("Method", method), q.Eq("URL", url), q.Eq("Private", true)),
		q.And(q.Eq("Method", method), q.Eq("URL", url), q.Eq("Private", false)),
	))

	var scheds []schedule
	err = query.Find(&scheds)
	if err != nil && err != storm.ErrNotFound { // not found is not an error but the desired state
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
		return
	}

	if len(scheds) > 0 {
		u.bot.Send(m.Sender, "Error:\nMethod and URL is already monitored.")
		var message string
		for _, s := range scheds {
			if s.UserID == m.Sender.ID {
				message = fmt.Sprintf("%s%d - %s %s (%d) (yours)\n", message, s.ID, s.Method, s.URL, s.ExpectedStatus)
			} else {
				message = fmt.Sprintf("%s%d - %s %s (%d)\n", message, s.ID, s.Method, s.URL, s.ExpectedStatus)
			}
		}
		u.bot.Send(m.Sender, message, tb.NoPreview)
		return
	}

	// insert in db
	sched.UserID = m.Sender.ID
	sched.Private = private
	sched.Method = method
	sched.URL = url
	sched.ExpectedStatus = expectedStatus
	sched.Every = interval
	sched.Subscriptors = []int{m.Sender.ID}

	err = u.db.Save(&sched)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
		return
	}

	// create scheduled task
	err = u.addJob(sched)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
		return
	}

	log.Println("Monitor added.", sched)

	u.bot.Send(m.Sender, "Monitor added.")

}

func (u *urlTester) subscribe(m *tb.Message) {

	var (
		sched   schedule
		message string
		err     error
	)

	sched, message = u.getScheduleByIDString(m.Payload)
	if message != "" {
		u.bot.Send(m.Sender, message, tb.NoPreview)
		return
	}

	if alreadyOnIntArray(sched.Subscriptors, m.Sender.ID) {
		u.bot.Send(m.Sender, "Already subscribed.")
		return
	}

	sched.Subscriptors = append(sched.Subscriptors, m.Sender.ID)

	err = u.db.Save(&sched)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
		return
	}

	u.bot.Send(m.Sender, "Successfully subscribed.")

}

func (u *urlTester) unsubscribe(m *tb.Message) {

	var (
		sched   schedule
		message string
		err     error
	)

	sched, message = u.getScheduleByIDString(m.Payload)
	if message != "" {
		u.bot.Send(m.Sender, message, tb.NoPreview)
		return
	}

	if alreadyOnIntArray(sched.Subscriptors, m.Sender.ID) {
		// remove from array
		sched.Subscriptors = removeFromIntArray(sched.Subscriptors, m.Sender.ID)
		err = u.db.Save(&sched)
		if err != nil {
			u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
			return
		}
		u.bot.Send(m.Sender, "Unsubscribed.")
		return

	}

	u.bot.Send(m.Sender, "Not subscribed to the requested monitor.")

}
