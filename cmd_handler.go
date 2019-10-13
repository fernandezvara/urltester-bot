package main

import (
	"log"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (u *urlTester) handler(m *tb.Message) {

	if m.Text == "" {
		return
	}

	log.Printf("Command received: UserID: %d. Text: '%s'", m.Sender.ID, m.Text)
	parts := strings.Split(m.Text, " ")

	cmd, ok := u.commands[parts[0]]
	if !ok {
		u.bot.Send(m.Sender, "Command not found. Use /help.")
		return
	}

	// check if command is private
	if cmd.isPrivate == true {
		if !m.Private() {
			return
		}
	}

	// check if command can be executed anonymously or not
	if cmd.forUsers == true {
		if u.accessGranted(m.Sender) == false {
			return
		}
	}

	// check if command must be executed by an admin
	if cmd.forAdmins == true {
		if u.isUserAdmin(m.Sender.ID) == false {
			return
		}
	}

	cmd.fn(m)

}
