package main

import (
	"fmt"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

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

// saveHistory stores on the database the user interaction
func (u *urlTester) saveHistory(m *tb.Message) {
	u.db.Save(&history{
		When:    time.Now(),
		UserID:  m.Sender.ID,
		Message: m.Text,
	})
}
