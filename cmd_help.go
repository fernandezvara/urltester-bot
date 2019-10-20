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

func (u *urlTester) help(m *tb.Message, returns []interface{}) {

	var (
		anonMessage   string
		usersMessage  string
		adminsMessage string
		message       string
		cmdString     string
		cmd           command
	)

	switch len(returns) {
	case 1:

		cmdString = returns[0].(string)
		cmd = u.commands[cmdString]

		// one command help
		message = fmt.Sprintf("%s %s\n\n%s", cmdString, u.argsToString(cmd.payload), cmd.helpLong)

		if len(cmd.payload) > 0 {
			message = fmt.Sprintf("%s\n\n*Options*:\n", message)
		}
		for _, payloadPart := range cmd.payload {
			message = fmt.Sprintf("%s*<%s:%s>* - %s\n", message, payloadPart.arg, payloadPart.typ, payloadPart.help)
			if len(payloadPart.valid) > 0 {
				message = fmt.Sprintf("%sAllowed:", message)
				for i, p := range payloadPart.valid {
					if i == 0 {
						message = fmt.Sprintf("%s %s", message, p)
					} else {
						message = fmt.Sprintf("%s,%s", message, p)
					}
				}
				message = fmt.Sprintf("%s\n", message)
			}
		}

		if _, err := u.bot.Send(m.Sender, message, tb.ModeMarkdown); err != nil {
			log.Println(err)
		}
		return

	default:

		// build all commands message
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
			message = fmt.Sprintf(`%s*ADMIN COMMANDS*
%s
`, message, adminsMessage)
		}

		message = fmt.Sprintf("%s\n\nFor a longer explanation of a command, use:\n/command help\n", message)

		if _, err := u.bot.Send(m.Sender, message, tb.ModeMarkdown); err != nil {
			log.Println(err)
		}
		return

	}
	// one command help

}
