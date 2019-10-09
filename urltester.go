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
	"github.com/fernandezvara/scheduler"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (u *urlTester) start() error {

	var (
		err error
	)

	log.Println("Starting API ...")
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
			log.Println("Interrupt request received:", sig.String())
			u.db.Close() // stopping database gratefully
			log.Println("db closed")
			u.bot.Stop() // stopping bot gratefully
			log.Println("bot closed")
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

	log.Println("Starting Bot ...")
	u.bot.Start()
	return nil

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

func (u *urlTester) sendRequest(method, url string, expectedStatus int) (body string, headers map[string]string, httpStatus int, expected bool, err error) {

	var (
		client *http.Client
		req    *http.Request
		res    *http.Response
	)

	// init headers
	headers = make(map[string]string)

	client = &http.Client{
		Timeout: 2 * time.Minute,
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
