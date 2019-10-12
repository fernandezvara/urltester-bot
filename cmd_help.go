package main

import (
	"fmt"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (u *urlTester) help(m *tb.Message) {

	if !m.Private() {
		return
	}

	message := `*HELP*
/monitors - returns defined monitors
/summary - shows all monitor you are subscribed and returns its current status
/newmonitor <method> <url> <expected_http_status> <interval> <private> - Adds a new monitor
/remove <id> - removes a monitor
/subscribe <id> - subscribes you to the desired monitor
/unsubscribe <id> - removes a subscription
/test <method> <url> <expected_http_status> - Send a test request for an URL
/testfull - Sends a test request and returns the body
/history - returns your command history
/help - this text
`

	if u.isUserAdmin(m.Sender.ID) {
		message = fmt.Sprintf(`%s
*ADMIN COMMANDS*
/grant <id> - Grant permissions for and user by ID
/revoke <id> - Revoke permissions for an user by ID
/users - List users of the bot and its authorization status
`, message)
	}

	u.bot.Send(m.Sender, message, tb.ModeMarkdown)
}

func (u *urlTester) hello(m *tb.Message) {
	if !m.Private() {
		return
	}

	u.bot.Send(m.Sender, "hello world")
}
