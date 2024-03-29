package main

import (
	"fmt"
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
	if err != nil {
		return err
	}
	u.db.Init(&history{})
	u.db.Init(&schedule{})
	u.db.Init(&user{})
	u.db.Init(&timeline{})

	// schedule map
	u.schedules = make(map[int]*scheduler.Job)
	u.lastStatus = make(map[int]timeline)
	u.commands = u.buildCommands()

	// set up bot
	u.bot, err = tb.NewBot(tb.Settings{
		Token:    u.token,
		Poller:   &tb.LongPoller{Timeout: 10 * time.Second},
		Reporter: collector,
	})

	if err != nil {
		return err
	}

	u.bot.Handle("/about", func(m *tb.Message) {
		u.bot.Send(m.Sender, fmt.Sprintf("*Version:*\n%s\n*Repo URL:*\n[%s]\n*Commit:*\n'%s'\n*Build date:*\n%s", Version, RepoURL, Commit, BuildDate), tb.NoPreview, tb.ModeMarkdown)
	})
	u.bot.Handle(tb.OnText, u.handler)

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
	log.Printf("Version: %s, Commit: '%s', Build date: %s", Version, Commit, BuildDate)
	u.bot.Start()
	return nil

}
