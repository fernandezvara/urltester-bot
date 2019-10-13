package main

import (
	"fmt"
	"log"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (u *urlTester) argsToString(payload []payloadPart) (message string) {

	for _, p := range payload {
		message = fmt.Sprintf("%s <%s>", message, p.arg)
	}

	return

}

func (u *urlTester) showHelp(m *tb.Message, cmdString string, long bool) {

	var (
		anonMessage   string
		usersMessage  string
		adminsMessage string
		message       string
	)

	// build all commands message
	if cmdString == "" {
		for key, value := range u.commands {
			if key != "/start" { // start command would not need help
				if value.forUsers == false && value.forAdmins == false {
					anonMessage = fmt.Sprintf("%s%s %s - %s\n", anonMessage, key, u.argsToString(value.payload), value.helpShort)
				}
			}
		}
		for key, value := range u.commands {
			if value.forUsers == true && value.forAdmins == false {
				usersMessage = fmt.Sprintf("%s%s %s - %s\n", usersMessage, key, u.argsToString(value.payload), value.helpShort)
			}
		}
		if u.isUserAdmin(m.Sender.ID) {
			for key, value := range u.commands {
				if value.forAdmins == true {
					adminsMessage = fmt.Sprintf("%s%s %s - %s\n", adminsMessage, key, u.argsToString(value.payload), value.helpShort)
				}
			}
		}

		// anonymous commands
		message = fmt.Sprintf(`*HELP*
%s
`, anonMessage)

		// users commands
		if u.accessGranted(m.Sender) == true {
			message = fmt.Sprintf(`%s*USER COMMANDS*
%s
`, message, usersMessage)
		}

		// admins commands
		if u.isUserAdmin(m.Sender.ID) {
			message = fmt.Sprintf(`%s

*ADMIN COMMANDS*
%s
`, message, adminsMessage)
		}

		if _, err := u.bot.Send(m.Sender, message, tb.ModeMarkdown); err != nil {
			log.Println(err)
		}
		return
	}

	// one command help

}

func (u *urlTester) help(m *tb.Message) {

	// 	message := `*HELP*
	// /monitors - returns defined monitors
	// /summary - shows all monitor you are subscribed and returns its current status
	// /newmonitor <method> <url> <expected_http_status> <interval> <private> - Adds a new monitor
	// /remove <id> - removes a monitor
	// /subscribe <id> - subscribes you to the desired monitor
	// /unsubscribe <id> - removes a subscription
	// /test <method> <url> <expected_http_status> - Send a test request for an URL
	// /testfull - Sends a test request and returns the body
	// /history - returns your command history
	// /help - this text

	// *BETA commands*
	// /setinterval - <id> <newinterval>
	// /setstatuscode - <id> <statuscode>
	// /settext - <id> <expectedtext>
	// /settimeout - <id> <timeout>
	// `

	// 	if u.isUserAdmin(m.Sender.ID) {
	// 		message = fmt.Sprintf(`%s
	// *ADMIN COMMANDS*
	// /grant <id> - Grant permissions for and user by ID
	// /revoke <id> - Revoke permissions for an user by ID
	// /users - List users of the bot and its authorization status
	// `, message)
	// 	}

	// 	u.bot.Send(m.Sender, message, tb.ModeMarkdown)

	u.showHelp(m, "", false)

}
