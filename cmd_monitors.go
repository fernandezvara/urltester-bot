package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (u *urlTester) setinterval(m *tb.Message) {

	u.saveHistory(m)

	var (
		id       int
		interval string
		returns  []interface{}
		err      error
		sched    schedule
	)

	returns, err = u.payloadReader(m.Text)
	if err != nil {
		u.bot.Send(m.Sender, err.Error())
		return
	}

	id = returns[0].(int)
	interval = returns[1].(string)

	sched, err = u.getScheduleByID(id)
	if err != nil {
		u.bot.Send(m.Sender, err.Error())
		return
	}

	sched.Every = interval

	err = u.updateSchedule(&sched)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
		return
	}

	u.bot.Send(m.Sender, "Interval updated.")

}
func (u *urlTester) setstatuscode(m *tb.Message) {

	u.saveHistory(m)

	var (
		id         int
		statusCode int
		returns    []interface{}
		err        error
		sched      schedule
	)

	returns, err = u.payloadReader(m.Text)
	if err != nil {
		u.bot.Send(m.Sender, err.Error())
		return
	}

	id = returns[0].(int)
	statusCode = returns[1].(int)

	sched, err = u.getScheduleByID(id)
	if err != nil {
		u.bot.Send(m.Sender, err.Error())
		return
	}

	sched.ExpectedStatus = statusCode

	err = u.updateSchedule(&sched)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
		return
	}

	u.bot.Send(m.Sender, "Status code updated.")

}

func (u *urlTester) settext(m *tb.Message) {

	u.saveHistory(m)

	var (
		id      int
		text    string
		returns []interface{}
		err     error
		sched   schedule
	)

	returns, err = u.payloadReader(m.Text)
	if err != nil {
		u.bot.Send(m.Sender, err.Error())
		return
	}

	id = returns[0].(int)
	text = returns[1].(string)

	sched, err = u.getScheduleByID(id)
	if err != nil {
		u.bot.Send(m.Sender, err.Error())
		return
	}

	sched.ExpectedText = text

	err = u.updateSchedule(&sched)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
		return
	}

	u.bot.Send(m.Sender, "Text updated.")

}

func (u *urlTester) settimeout(m *tb.Message) {

	u.saveHistory(m)

	var (
		id      int
		timeout string
		returns []interface{}
		err     error
		sched   schedule
	)

	returns, err = u.payloadReader(m.Text)
	if err != nil {
		u.bot.Send(m.Sender, err.Error())
		return
	}

	id = returns[0].(int)
	timeout = returns[1].(string)

	sched, err = u.getScheduleByID(id)
	if err != nil {
		u.bot.Send(m.Sender, err.Error())
		return
	}

	sched.ExpectedTimeout = timeout

	err = u.updateSchedule(&sched)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
		return
	}

	u.bot.Send(m.Sender, "Timeout updated updated.")

}

func (u *urlTester) updateSchedule(sched *schedule) (err error) {

	u.Lock()
	u.schedules[sched.ID].Quit <- true
	delete(u.schedules, sched.ID)
	log.Println("job removed", sched)
	u.Unlock()

	err = u.addJob(*sched)
	if err != nil {
		return
	}

	err = u.db.Save(sched)
	if err != nil {
		return
	}

	return
}

func (u *urlTester) newmonitor(m *tb.Message) {

	var (
		sched      schedule
		returns    []interface{}
		method     string
		urlString  string
		statusCode int
		interval   string
		private    bool
		err        error
	)

	u.saveHistory(m)

	returns, err = u.payloadReader(m.Text)
	if err != nil {
		u.bot.Send(m.Sender, err.Error())
		return
	}

	method = returns[0].(string)
	urlString = returns[1].(string)
	statusCode = returns[2].(int)
	interval = returns[3].(string)
	private = returns[4].(bool)

	// monitor exists? look for the private urls of the current user or public ones
	query := u.db.Select(q.Or(
		q.And(q.Eq("UserID", m.Sender.ID), q.Eq("Method", method), q.Eq("URL", urlString), q.Eq("Private", true)),
		q.And(q.Eq("Method", method), q.Eq("URL", urlString), q.Eq("Private", false)),
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
				message = fmt.Sprintf("%s%d - %s [%s] (%d) *(yours)*\n", message, s.ID, s.Method, s.URL, s.ExpectedStatus)
			} else {
				message = fmt.Sprintf("%s%d - %s [%s] (%d)\n", message, s.ID, s.Method, s.URL, s.ExpectedStatus)
			}
		}
		u.bot.Send(m.Sender, message, tb.NoPreview, tb.ModeMarkdown)
		return
	}

	// insert in db
	sched.UserID = m.Sender.ID
	sched.Private = private
	sched.Method = strings.ToUpper(method)
	sched.URL = urlString
	sched.ExpectedStatus = statusCode
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

	// write a fake timeline (zero status)
	// on first interval remote is down it will report the downtime
	newTimeline := timeline{
		MonitorID: sched.ID,
		Timestamp: time.Now().Unix(),
		Status:    statusUp,
		Downtime:  0,
	}

	u.Lock()
	u.lastStatus[sched.ID] = newTimeline
	u.Unlock()

	log.Println("Monitor added.", sched)

	u.bot.Send(m.Sender, "Monitor added.")

}

func (u *urlTester) remove(m *tb.Message) {

	u.saveHistory(m)

	var (
		sched   schedule
		id      int
		returns []interface{}
		err     error
	)

	returns, err = u.payloadReader(m.Text)
	if err != nil {
		u.bot.Send(m.Sender, err.Error())
		return
	}

	id = returns[0].(int)

	sched, err = u.getScheduleByID(id)
	if err != nil {
		u.bot.Send(m.Sender, err.Error())
		return
	}

	if sched.UserID != m.Sender.ID {
		u.bot.Send(m.Sender, "ERROR: You can't delete monitors not created by you.")
		return
	}

	// notify suscribers about this removal
	for subscriptor := range sched.Subscriptors {
		if subscriptor != m.Sender.ID {
			u.bot.Send(telegramUser{subscriptor}, fmt.Sprintf("Monitor %d was removed by its owner. Settings:\nMethod: %s\nURL: %s\nExpected HTTP status: %d\nInterval: %s\n", sched.ID, sched.Method, sched.URL, sched.ExpectedStatus, sched.Every), tb.NoPreview)
		}
	}

	err = u.db.DeleteStruct(&sched)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
		return
	}

	u.Lock()
	u.schedules[sched.ID].Quit <- true
	delete(u.schedules, sched.ID)
	u.Unlock()

	log.Println("Monitor removed.", sched)

	u.bot.Send(m.Sender, "Monitor removed.")

}
