package main

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/asdine/storm"
	"github.com/fernandezvara/scheduler"
	tb "gopkg.in/tucnak/telebot.v2"
)

// errors
var (
	errInvalidPayload = errors.New("Invalid payload: <method> <url> <expected_code>")
	errInvalidMethod  = errors.New("Invalid method: Only GET, POST, PUT and OPTIONS allowed")
)

// consts
var (
	allowedMethods = []string{"GET", "POST", "PUT", "OPTIONS"}
)

// urlTester is the main struct that holds the service parts that need to interact
type urlTester struct {
	db        *storm.DB
	bot       *tb.Bot
	dbpath    string
	token     string
	schedules map[int]*scheduler.Job
	sync.RWMutex
}

// schedule is the definition of a recurrent monitoring job
type schedule struct {
	ID             int    `json:"id" storm:"id,increment"`
	UserID         int    `json:"user_id" storm:"index"`
	Private        bool   `json:"private" storm:"index"`
	Method         string `json:"method" storm:"index"`
	URL            string `json:"url" storm:"index"`
	ExpectedStatus int    `json:"expected_status"`
	Every          string `json:"every"`
	Paused         bool   `json:"paused"`
	Subscriptors   []int  `json:"subscriptors"`
}

// history saves the user interactions with the bot by user
type history struct {
	ID      int       `json:"id" storm:"id,increment"`
	When    time.Time `json:"when"`
	UserID  int       `json:"user_id" storm:"index"`
	Message string    `json:"message"`
}

type telegramUser struct {
	id int
}

func (t telegramUser) Recipient() string {
	return strconv.Itoa(t.id)
}
