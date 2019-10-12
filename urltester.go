package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/asdine/storm"
	"github.com/fernandezvara/scheduler"
	tb "gopkg.in/tucnak/telebot.v2"
)

func collector(err error) {
	log.Println("-- err:")
	log.Println(err)
	log.Println("-- err:")
}

func (u *urlTester) botstart() error {

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
		Token:    u.token,
		Poller:   &tb.LongPoller{Timeout: 10 * time.Second},
		Reporter: collector,
	})

	if err != nil {
		return err
	}

	// start command
	u.bot.Handle("/start", u.start)

	// user commands
	u.bot.Handle("/hello", u.hello)
	u.bot.Handle("/summary", u.summary)
	u.bot.Handle("/monitors", u.monitors)
	u.bot.Handle("/newmonitor", u.newmonitor)
	u.bot.Handle("/remove", u.remove)
	u.bot.Handle("/subscribe", u.subscribe)
	u.bot.Handle("/unsubscribe", u.unsubscribe)
	u.bot.Handle("/test", u.test)
	u.bot.Handle("/testfull", u.testFull)
	u.bot.Handle("/history", u.history)
	u.bot.Handle("/help", u.help)

	// admin commands
	u.bot.Handle("/grant", u.grant)
	u.bot.Handle("/revoke", u.revoke)
	u.bot.Handle("/users", u.users)

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
