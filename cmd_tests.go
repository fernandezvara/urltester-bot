package main

import (
	"fmt"
	"log"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (u *urlTester) test(m *tb.Message) {
	u.testURL(m, false)
}

func (u *urlTester) testFull(m *tb.Message) {
	u.testURL(m, true)
}

func (u *urlTester) testURL(m *tb.Message, full bool) {

	var (
		headerString string
		message      string
	)

	if !m.Private() || !u.accessGranted(m.Sender) {
		return
	}
	u.saveHistory(m)

	method, urlString, _, _, expectedStatus, err := u.cleanPayload(m.Payload, false)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
	}

	body, headers, resultCode, expected, err := u.sendRequest(method, urlString, expectedStatus)
	log.Println(method, urlString, expectedStatus, resultCode, expected, err)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
		return
	}

	headerString = headersToString(headers)

	if full {
		message = fmt.Sprintf("Expected result: %t\n\nhttp status: %d\n\nHeaders:\n%s\n\nBody:\n%s\n", expected, resultCode, headerString, body)
	} else {
		message = fmt.Sprintf("Expected result: %t\n\nhttp status: %d\n\nHeaders:\n%s\n", expected, resultCode, headerString)
	}

	u.bot.Send(m.Sender, message)

}
