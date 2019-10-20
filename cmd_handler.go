package main

import (
	"fmt"
	"log"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (u *urlTester) handler(m *tb.Message) {

	var (
		cmdString string
		returns   []interface{}
		err       error
	)

	if m.Text == "" {
		return
	}

	log.Printf("Command received: User: %s %s (%d). Text: '%s'", m.Sender.FirstName, m.Sender.LastName, m.Sender.ID, m.Text)

	cmdString, returns, err = u.payloadReader(m.Text)
	if err != nil {
		u.explainError(m, cmdString, err)
		return
	}

	cmd := u.commands[cmdString]

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

	cmd.fn(m, returns)

}

func (u *urlTester) explainError(m *tb.Message, cmdString string, err error) {

	switch err {
	case errInvalidPayload:
		// get the command and it to the help func
		u.help(m, []interface{}{cmdString})
	case errCommandNotFound:
		u.bot.Send(m.Sender, fmt.Sprintf("'%s' command not found.\n Use /help for more info.", cmdString))
	default:
		u.explainError(m, "", err)
	}

}
