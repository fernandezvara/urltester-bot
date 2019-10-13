package main

import (
	"fmt"
	"time"

	"github.com/asdine/storm"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (u *urlTester) history(m *tb.Message) {

	var (
		err       error
		histories []history
		message   string
	)

	u.saveHistory(m)
	message = "-- History --\n"
	err = u.db.Find("UserID", m.Sender.ID, &histories, storm.Limit(20), storm.Reverse()) // the last 20 messages or less
	if err != nil {
		u.bot.Send(m.Sender, "there was an error retrieving information")
		u.bot.Send(m.Sender, err.Error())
		return
	}

	for i := len(histories); i > 0; i-- {
		message = fmt.Sprintf("%s%s - %s\n", message, histories[i-1].When.Format("02/01/2006 15:04:05"), histories[i-1].Message)
	}

	u.bot.Send(m.Sender, message, tb.NoPreview)

}

// saveHistory stores on the database the user interaction
func (u *urlTester) saveHistory(m *tb.Message) {
	u.db.Save(&history{
		When:    time.Now(),
		UserID:  m.Sender.ID,
		Message: m.Text,
	})
}
