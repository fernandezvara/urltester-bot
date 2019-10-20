package main

import (
	"fmt"

	tb "gopkg.in/tucnak/telebot.v2"
)

// start : response when the user makes the first connection to the bot
// the client will send a /start on the first contact
func (u *urlTester) start(m *tb.Message, returns []interface{}) {

	u.saveHistory(m)

	if u.accessGranted(m.Sender) {
		u.bot.Send(m.Sender, "You already have permissions to use this bot")
		return
	}

	var (
		tgUser  user
		err     error
		isAdmin bool
	)

	tgUser, _ = u.userInfo(m.Sender.ID)
	if tgUser.FirstName != "" || tgUser.LastName != "" || tgUser.LanguageCode != "" {
		u.bot.Send(m.Sender, "Access already requested.")
		return
	}

	tgUser.ID = m.Sender.ID
	tgUser.FirstName = m.Sender.FirstName
	tgUser.LastName = m.Sender.LastName
	tgUser.LanguageCode = m.Sender.LanguageCode
	tgUser.IsBot = m.Sender.IsBot

	// if the user is on admins array it will be automatically authorized to use the bot
	if u.isUserAdmin(m.Sender.ID) {
		tgUser.Authorized = true
		isAdmin = true
	}

	err = u.db.Save(&tgUser)
	if err != nil {
		u.sendMessageAndNotifyAdmins(m.Sender.ID, fmt.Sprintf("There was an error:\n%s", err.Error()))
		return
	}

	if isAdmin {
		u.bot.Send(m.Sender, "Welcome Admin!")
		return
	}

	u.bot.Send(m.Sender, "An Administration grant has been requested.")
	u.bot.Send(m.Sender, fmt.Sprintf("This is an open source project. You can setup your own bot: %s", RepoURL))
	u.sendMessageToAdmins(fmt.Sprintf("A new user request access: %s %s (%d)", tgUser.FirstName, tgUser.LastName, tgUser.ID))

}
