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
		returns      []interface{}
		method       string
		urlString    string
		statusCode   int
		err          error
	)

	u.saveHistory(m)

	returns, err = u.payloadReader(m.Text)
	if err != nil {
		u.bot.Send(m.Sender, err.Error())
		return
	}

	method = returns[0].(string)
	urlString = returns[1].(string)
	statusCode = returns[2].(int)

	duration, body, headers, resultCode, expected, err := u.sendRequest(method, urlString, statusCode, "", "")
	log.Println(method, urlString, statusCode, resultCode, expected, err)
	if err != nil {
		u.bot.Send(m.Sender, fmt.Sprintf("There was an error:\n%s", err.Error()))
		return
	}

	headerString = headersToString(headers)

	if full {
		message = fmt.Sprintf("Expected result: %t\nDuration: %s\nhttp status: %d\n\nHeaders:\n%s\n\nBody:\n%s\n", expected, duration.String(), resultCode, headerString, body)
	} else {
		message = fmt.Sprintf("Expected result: %t\nDuration: %s\nhttp status: %d\n\nHeaders:\n%s\n", expected, duration.String(), resultCode, headerString)
	}

	u.bot.Send(m.Sender, message)

}
