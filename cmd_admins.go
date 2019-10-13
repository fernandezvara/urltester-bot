package main

import (
	"fmt"

	tb "gopkg.in/tucnak/telebot.v2"
)

// grantOrRevoke modifies the authorized status for a user
func (u *urlTester) users(m *tb.Message) {

	u.saveHistory(m)

	var (
		users []user
		err   error
	)

	err = u.db.All(&users)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error: %s", err.Error()))
		return
	}

	var message string
	for _, thisUser := range users {
		message = fmt.Sprintf("%s%d - %s %s", message, thisUser.ID, thisUser.FirstName, thisUser.LastName)
		if thisUser.Username != "" {
			message = fmt.Sprintf("%s @%s", message, thisUser.Username)
		}
		if thisUser.Authorized {
			message = fmt.Sprintf("%s *(autorized)*", message)
		}
		if u.isUserAdmin(thisUser.ID) {
			message = fmt.Sprintf("%s _(admin)_", message)
		}
		message = fmt.Sprintf("%s\n", message)
	}

	u.bot.Send(m.Sender, message, tb.ModeMarkdown)

}

func (u *urlTester) grant(m *tb.Message) {

	u.grantOrRevoke(m, true)

}

func (u *urlTester) revoke(m *tb.Message) {

	u.grantOrRevoke(m, false)

}

// grantOrRevoke modifies the authorized status for a user
func (u *urlTester) grantOrRevoke(m *tb.Message, action bool) {

	u.saveHistory(m)

	var (
		id      int
		returns []interface{}
		err     error
		tgUser  user
	)

	returns, err = u.payloadReader(m.Text)
	if err != nil {
		u.bot.Send(m.Sender, err.Error())
		return
	}

	id = returns[0].(int)

	err = u.db.One("ID", id, &tgUser)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error: %s", err.Error()))
		return
	}

	tgUser.Authorized = action
	err = u.db.Save(&tgUser)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error: %s", err.Error()))
		return
	}

	if action == true {
		u.bot.Send(m.Sender, "Permissions granted")
		u.bot.Send(telegramUser{id}, "Permissions granted\nUse /help to know the bot allowed actions.")
		return
	}

	u.bot.Send(m.Sender, "Permissions revoked")

}
