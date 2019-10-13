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

func (u *urlTester) help(m *tb.Message) {

	var (
		anonMessage   string
		usersMessage  string
		adminsMessage string
		message       string
	)

	// build all commands message
	if m.Payload == "" {

		for key, value := range u.commands {
			if !value.noHelp {
				if value.forUsers == false && value.forAdmins == false {
					anonMessage = fmt.Sprintf("%s%s %s - %s\n", anonMessage, key, u.argsToString(value.payload), value.helpShort)
				}
				if value.forUsers == true && value.forAdmins == false {
					usersMessage = fmt.Sprintf("%s%s %s - %s\n", usersMessage, key, u.argsToString(value.payload), value.helpShort)
				}
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
