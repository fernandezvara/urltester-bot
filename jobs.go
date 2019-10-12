package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fernandezvara/scheduler"
	tb "gopkg.in/tucnak/telebot.v2"
)

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
		log.Println("Monitor with no subscriptors, skipping... ", expected, err, args)
		return
	}

	_, _, resultStatusCode, expected, err = u.sendRequest(method, urlString, statusCode)

	if expected {
		if u.lastStatus[sched.ID].Status != statusUp {
			diff, err = u.addTimelineEntry(sched.ID, statusUp)
			for _, sub := range sched.Subscriptors {
				u.bot.Send(telegramUser{id: sub}, fmt.Sprintf("RESOLVED: %s %s (%d):\n\nDowntime: %s\n", sched.Method, sched.URL, sched.ExpectedStatus, secondsToHuman(diff)), tb.NoPreview)
			}
		}
	} else {
		if u.lastStatus[sched.ID].Status != statusDown {
			_, _ = u.addTimelineEntry(sched.ID, statusDown)
			for _, sub := range sched.Subscriptors {
				if err != nil {
					u.bot.Send(telegramUser{id: sub}, fmt.Sprintf("PROBLEM: (id:%d) %s %s (%d):\nerror: %s", sched.ID, sched.Method, sched.URL, sched.ExpectedStatus, err.Error()), tb.NoPreview)
				} else {
					u.bot.Send(telegramUser{id: sub}, fmt.Sprintf("PROBLEM: (id:%d) %s %s (%d):\nres status: %d", sched.ID, sched.Method, sched.URL, sched.ExpectedStatus, resultStatusCode), tb.NoPreview)
				}
			}
			return
		}

		for _, sub := range sched.Subscriptors {
			u.bot.Send(telegramUser{id: sub}, fmt.Sprintf("PROBLEM: (id:%d) %s %s (%d):\nres status: %d\nDowntime: %s\n", sched.ID, sched.Method, sched.URL, sched.ExpectedStatus, resultStatusCode, secondsToHuman(time.Now().Unix()-u.lastStatus[sched.ID].Timestamp)), tb.NoPreview)
		}
	}

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
