package main

import (
	"fmt"

	tb "gopkg.in/tucnak/telebot.v2"
)

// users returns current users of the bot and statuses
func (u *urlTester) users(m *tb.Message, returns []interface{}) {

	u.saveHistory(m)

	var (
		users []user
		err   error
	)

	err = u.db.All(&users)
	if err != nil {
		u.explainError(m, "", err)
		return
	}

	var message string
	for _, thisUser := range users {
		message = fmt.Sprintf("%s%d - %s %s", message, thisUser.ID, thisUser.FirstName, thisUser.LastName)
		if thisUser.Username != "" {
			message = fmt.Sprintf("%s @%s", message, thisUser.Username)
		}
		if thisUser.Authorized {
			message = fmt.Sprintf("%s *(authorized)*", message)
		}
		if u.isUserAdmin(thisUser.ID) {
			message = fmt.Sprintf("%s _(admin)_", message)
		}
		message = fmt.Sprintf("%s\n", message)
	}

	u.bot.Send(m.Sender, message, tb.ModeMarkdown)

}

func (u *urlTester) grant(m *tb.Message, returns []interface{}) {

	u.grantOrRevoke(m, true, returns)

}

func (u *urlTester) revoke(m *tb.Message, returns []interface{}) {

	u.grantOrRevoke(m, false, returns)

}

// grantOrRevoke modifies the authorized status for a user
func (u *urlTester) grantOrRevoke(m *tb.Message, action bool, returns []interface{}) {

	u.saveHistory(m)

	var (
		id int64

		err    error
		tgUser user
	)

	id = returns[0].(int64)

	err = u.db.One("ID", id, &tgUser)
	if err != nil {
		u.explainError(m, "", err)
		return
	}

	tgUser.Authorized = action
	err = u.db.Save(&tgUser)
	if err != nil {
		u.explainError(m, "", err)
		return
	}

	if action {
		u.bot.Send(m.Sender, "Permissions granted")
		u.bot.Send(telegramUser{id}, "Permissions granted\nUse /help to know the bot allowed actions.")
		return
	}

	u.bot.Send(m.Sender, "Permissions revoked")

}
