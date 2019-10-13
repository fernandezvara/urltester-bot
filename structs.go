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
	errInvalidPayload = errors.New("invalid payload")
)

// consts
var (
	allowedMethods = []string{"GET", "POST", "PUT", "OPTIONS"}
)

// statuses, there is a non-zero status to ensure comparations won't match for an empty entry
const (
	statusDown int = iota + 1
	statusUp
	statusStarted
	statusStopped
)

// payload consts
const (
	typeString  string = "string"
	typeInt     string = "int"
	typeBool    string = "bool"
	typeTimeExp string = "time"
)

type command struct {
	fn        func(m *tb.Message) // function to execute
	noHelp    bool                // hide on help command
	isPrivate bool                // command needs to be send privatelly to the bot
	forUsers  bool                // must be user of the bot in order of been able to use it
	forAdmins bool                // must be admin to use it
	helpShort string              // short help text description
	helpLong  string              // long help text
	payload   []payloadPart       // payload parts and validations
}

type payloadPart struct {
	arg   string
	typ   string
	valid []string
	help  string
}

// urlTester is the main struct that holds the service parts that need to interact
type urlTester struct {
	db         *storm.DB
	bot        *tb.Bot
	dbpath     string
	token      string
	schedules  map[int]*scheduler.Job
	lastStatus map[int]timeline
	commands   map[string]command
	admins     []int
	sync.RWMutex
}

// schedule is the definition of a recurrent monitoring job
type schedule struct {
	ID              int    `json:"id" storm:"id,increment"`
	UserID          int    `json:"user_id" storm:"index"`
	Private         bool   `json:"private" storm:"index"`
	Method          string `json:"method" storm:"index"`
	URL             string `json:"url" storm:"index"`
	ExpectedStatus  int    `json:"expected_status,omitempty"`
	ExpectedText    string `json:"expected_text,omitempty"`
	ExpectedTimeout string `json:"expected_timeout,omitempty"`
	Every           string `json:"every"`
	Subscriptors    []int  `json:"subscriptors"`
}

type user struct {
	ID           int    `json:"id" storm:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
	IsBot        bool   `json:"is_bot"`
	Authorized   bool   `json:"authorized"`
}

// history saves the user interactions with the bot by user
type history struct {
	ID      int       `json:"id" storm:"id,increment"`
	When    time.Time `json:"when"`
	UserID  int       `json:"user_id" storm:"index"`
	Message string    `json:"message"`
}

// timeline stores the status of each monitor
type timeline struct {
	ID        int   `json:"id" storm:"id,increment"`
	MonitorID int   `json:"monitor_id"`
	Timestamp int64 `json:"timestamp"`
	Status    int   `json:"status" storm:"index"`
	Downtime  int64 `json:"downtime"`
	duration  time.Duration
	body      string
	headers   map[string]string
}

// telegramUser uses the same interface than *tb.User
type telegramUser struct {
	id int
}

func (t telegramUser) Recipient() string {
	return strconv.Itoa(t.id)
}
