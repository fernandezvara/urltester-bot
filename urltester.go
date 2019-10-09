package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/fernandezvara/scheduler"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (u *urlTester) start() error {

	var (
		err error
	)

	// set up database
	u.db, err = storm.Open(u.dbpath)
	u.db.Init(&history{})
	u.db.Init(&schedule{})

	// schedule map
	u.schedules = make(map[int]*scheduler.Job)
	u.lastStatus = make(map[int]timeline)

	// set up bot
	u.bot, err = tb.NewBot(tb.Settings{
		Token:  u.token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		return err
	}

	u.bot.Handle("/hello", u.hello)
	u.bot.Handle("/monitors", u.monitors)
	u.bot.Handle("/newmonitor", u.newmonitor)
	u.bot.Handle("/subscribe", u.subscribe)
	u.bot.Handle("/unsubscribe", u.unsubscribe)
	u.bot.Handle("/test", u.test)
	u.bot.Handle("/testfull", u.testFull)
	u.bot.Handle("/history", u.history)
	u.bot.Handle("/help", u.help)

	// handle stop
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		for sig := range ch {
			// sig is a ^C, handle it
			fmt.Println("Interrupt request received:", sig.String())
			u.db.Close() // stopping database gratefully
			fmt.Println("db closed")
			u.bot.Stop() // stopping bot gratefully
			fmt.Println("bot closed")
			os.Exit(0)
		}
	}()

	// load up monitors from database
	var schedules []schedule
	err = u.db.All(&schedules)
	if err != nil && err != storm.ErrNotFound {
		return err
	}

	for _, sched := range schedules {
		// add job to the scheduler
		u.addJob(sched)
		// // get the last timeline entry for this monitor
		u.Lock()
		u.lastStatus[sched.ID] = u.getLastTimelineEntry(sched.ID)
		u.Unlock()
	}

	u.bot.Start()
	return nil

}

// addJob adds a job the the scheduler,
// TODO: for update a scheduled job, first delete and recreate, the ID will remain the same
func (u *urlTester) addJob(sched schedule) error {

	var (
		err    error
		amount int
		unit   string
		job    *scheduler.Job
	)

	u.Lock()
	amount, unit, _ = evaluateTimeExp(sched.Every)
	job, err = newScheduledJob(amount, unit).NotImmediately().RunWithArgs(u.executeMonitor, []string{strconv.Itoa(sched.ID), sched.Method, sched.URL, strconv.Itoa(sched.ExpectedStatus)})
	if err != nil {
		return err
	}
	u.schedules[sched.ID] = job
	log.Println("job added", sched.ID, sched.Method, sched.URL, sched.ExpectedStatus, sched.Every)
	u.Unlock()

	return nil
}

// saveHistory stores on the database the user interaction
func (u *urlTester) saveHistory(m *tb.Message) {
	u.db.Save(&history{
		When:    time.Now(),
		UserID:  m.Sender.ID,
		Message: m.Text,
	})
}

func (u *urlTester) help(m *tb.Message) {
	if !m.Private() {
		return
	}

	u.bot.Send(m.Sender, `/monitors - returns defined monitors
/newmonitor <method> <url> <expected_http_status> <interval> <private> - Adds a new monitor
/subscribe <id> - subscribes you to the desired monitor
/unsubscribe <id> - removes a subscription
/test <method> <url> <expected_http_status> - Send a test request for an URL
/testfull - Sends a test request and returns the body
/history - returns your command history
/help - this text
`)
}

func (u *urlTester) hello(m *tb.Message) {
	if !m.Private() {
		return
	}
	u.saveHistory(m)
	fmt.Println(m.Sender)

	u.bot.Send(m.Sender, "hello world")
}

func (u *urlTester) monitors(m *tb.Message) {

	if !m.Private() {
		return
	}
	u.saveHistory(m)

	query := u.db.Select(q.Or(
		q.And(q.Eq("UserID", m.Sender.ID), q.Eq("Private", true)),
		q.And(q.Eq("Private", false)),
	))

	var (
		scheds  []schedule
		message string
		err     error
	)

	err = query.Find(&scheds)
	if err != nil {
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

	fmt.Println("sched")
	fmt.Println(sched)

	// create scheduled task
	err = u.addJob(sched)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
		return
	}

	fmt.Println(m.Sender)
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

func (u *urlTester) getScheduleByIDString(idString string) (sched schedule, message string) {
	var (
		parts []string
		id    int
		err   error
	)

	parts = strings.Split(idString, " ")
	if len(parts) != 1 {
		message = "Please write an ID to subscribe to."
		return
	}

	id, err = strconv.Atoi(parts[0])
	if err != nil {
		log.Println("ERROR: Unexpected ID:", parts)
		message = "Unexpected ID."
		return
	}

	err = u.db.One("ID", id, &sched)
	if err != nil {
		log.Println("ERROR: ID request with error:\n", parts)
		if err == storm.ErrNotFound {
			message = "ID not found"
			return
		}
		message = "Unexpected error."
		return
	}

	return
}

func alreadyOnIntArray(arr []int, value int) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

func removeFromIntArray(arr []int, value int) (newArr []int) {

	for _, v := range arr {
		if v != value {
			newArr = append(newArr, v)
		}
	}
	return
}

func (u *urlTester) executeMonitor(args []string) {

	var (
		id               int
		method           string
		urlString        string
		statusCode       int
		resultStatusCode int
		expected         bool
		err              error
		sched            schedule
		diff             int64
	)

	id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Println("ERROR: Unexpected Status code. Monitor called with these args:\n", args)
		return
	}
	method = args[1]
	urlString = args[2]
	statusCode, err = strconv.Atoi(args[3])
	if err != nil {
		log.Println("ERROR: Unexpected Status code. Monitor called with these args:\n", args)
		return
	}

	err = u.db.One("ID", id, &sched)
	if err != nil {
		log.Println("ERROR: Unexpected Status code. Monitor called with these args:\n", args)
		return
	}

	if len(sched.Subscriptors) == 0 {
		log.Println("no subscriptors", expected, err, args)
		return
	}

	_, _, resultStatusCode, expected, err = u.sendRequest(method, urlString, statusCode)

	if expected {
		if u.lastStatus[sched.ID].Status != statusUp {
			diff, err = u.addTimelineEntry(sched.ID, statusUp)
			fmt.Println(diff, err)
			for _, sub := range sched.Subscriptors {
				u.bot.Send(telegramUser{id: sub}, fmt.Sprintf("RESOLVED: %s %s (%d):\n\nDowntime: %s\n", sched.Method, sched.URL, sched.ExpectedStatus, secondsToHuman(diff)), tb.NoPreview)
			}
		}
	} else {
		if u.lastStatus[sched.ID].Status != statusDown {
			_, err = u.addTimelineEntry(sched.ID, statusDown)
			for _, sub := range sched.Subscriptors {
				u.bot.Send(telegramUser{id: sub}, fmt.Sprintf("PROBLEM: (id:%d) %s %s (%d):\nrc: %d", sched.ID, sched.Method, sched.URL, sched.ExpectedStatus, resultStatusCode), tb.NoPreview)
			}
			return
		}

		fmt.Println(time.Now().Unix()-u.lastStatus[sched.ID].Timestamp, secondsToHuman(time.Now().Unix()-u.lastStatus[sched.ID].Timestamp))
		for _, sub := range sched.Subscriptors {
			u.bot.Send(telegramUser{id: sub}, fmt.Sprintf("PROBLEM: (id:%d) %s %s (%d):\nrc: %d\nDowntime: %s\n", sched.ID, sched.Method, sched.URL, sched.ExpectedStatus, resultStatusCode, secondsToHuman(time.Now().Unix()-u.lastStatus[sched.ID].Timestamp)), tb.NoPreview)
		}
	}

}

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

func newScheduledJob(amount int, unit string) *scheduler.Job {

	var this *scheduler.Job
	switch unit {
	case "s", "S":
		this = scheduler.Every(amount).Seconds()
	case "m", "M":
		this = scheduler.Every(amount).Minutes()
	case "h", "H":
		this = scheduler.Every(amount).Hours()
	}
	return this

}

// evaluateTimeExp verifies that the expression to evaluate meets the requirements
// time expressions:
// (n)s = n seconds
// (n)m = n minutes
// (n)h = n hours
func evaluateTimeExp(exp string) (int, string, bool) {

	var matcher = regexp.MustCompile(`^([0-9]+)([a-zA-Z])$`)

	parts := matcher.FindAllStringSubmatch(exp, -1)

	if len(parts) == 1 {
		if len(parts[0]) == 3 {

			i, _ := strconv.Atoi(parts[0][1])
			switch parts[0][2] {
			case "s", "S", "m", "M", "h", "H":
				return i, strings.ToLower(parts[0][2]), true
			default:
				return 0, "", false
			}

		}
	}
	return 0, "", false

}

func (u *urlTester) test(m *tb.Message) {
	u.testURL(m, false)
}

func (u *urlTester) testFull(m *tb.Message) {
	u.testURL(m, true)
}

func (u *urlTester) testURL(m *tb.Message, full bool) {

	var (
		headerString string
		message      string
	)

	if !m.Private() {
		return
	}
	u.saveHistory(m)

	method, urlString, _, _, expectedStatus, err := u.cleanPayload(m.Payload, false)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
	}

	body, headers, resultCode, expected, err := u.sendRequest(method, urlString, expectedStatus)
	fmt.Println(method, urlString, expectedStatus, resultCode, expected, err)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
		return
	}

	headerString = headersToString(headers)

	if full {
		message = fmt.Sprintf("Expected result: %t\n\nhttp status: %d\n\nHeaders:\n%s\n\nBody:\n%s\n", expected, resultCode, headerString, body)
	} else {
		message = fmt.Sprintf("Expected result: %t\n\nhttp status: %d\n\nHeaders:\n%s\n", expected, resultCode, headerString)
	}

	u.bot.Send(m.Sender, message)

}

func headersToString(headers map[string]string) (headersString string) {

	for k, v := range headers {
		headersString = fmt.Sprintf("%s%s: %s\n", headersString, k, v)
	}
	return

}

func (u *urlTester) cleanPayload(payload string, isSchedule bool) (method, url, interval string, private bool, statusCode int, err error) {

	var parts int = 3

	if isSchedule == true {
		parts = 5
	}

	payloadParts := strings.Split(payload, " ")
	if len(payloadParts) != parts {
		if isSchedule == true {
			err = errInvalidPayloadNewMonitor
			return
		}
		err = errInvalidPayloadTest
		return
	}

	statusCode, err = strconv.Atoi(payloadParts[2])
	if err != nil {
		return
	}

	method = strings.ToUpper(payloadParts[0])
	if u.methodAllowed(method) == false {
		err = errInvalidMethod
		return
	}

	url = payloadParts[1]

	if isSchedule {
		// interval must be defined as time?
		interval = payloadParts[3]

		if payloadParts[4] == "true" {
			private = true
		}
	}

	return

}

func (u *urlTester) methodAllowed(method string) bool {
	for _, m := range allowedMethods {
		if method == m {
			return true
		}
	}

	return false
}

func (u *urlTester) history(m *tb.Message) {

	var (
		err       error
		histories []history
		message   string
	)

	if !m.Private() {
		return
	}

	message = "-- History --\n"
	err = u.db.Find("UserID", m.Sender.ID, &histories)
	if err != nil {
		u.bot.Send(m.Sender, "there was an error retrieving information")
		u.bot.Send(m.Sender, err.Error())
		return
	}

	for _, h := range histories {
		message = fmt.Sprintf("%s%s - %s\n", message, h.When.Format("02/01/2006 15:04:05"), h.Message)
	}

	u.bot.Send(m.Sender, message, tb.NoPreview)

}

func (u *urlTester) sendRequest(method, url string, expectedStatus int) (body string, headers map[string]string, httpStatus int, expected bool, err error) {

	var (
		client *http.Client
		req    *http.Request
		res    *http.Response
	)

	// init headers
	headers = make(map[string]string)

	client = &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err = http.NewRequest(strings.ToUpper(method), url, nil)
	if err != nil {
		return
	}

	res, err = client.Do(req)
	if err != nil {
		return
	}

	// body
	defer res.Body.Close()
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	body = string(bodyBytes)

	// expectedStatus ok
	if expectedStatus == res.StatusCode {
		expected = true
	}

	for k, v := range res.Header {
		var vstring string
		for _, vv := range v {
			vstring = fmt.Sprintf("%s %s", vstring, vv)
		}

		headers[k] = vstring
	}

	httpStatus = res.StatusCode

	return
}
