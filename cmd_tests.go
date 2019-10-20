package main

import (
	"fmt"
	"log"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (u *urlTester) test(m *tb.Message, returns []interface{}) {
	u.testURL(m, false, returns)
}

func (u *urlTester) testFull(m *tb.Message, returns []interface{}) {
	u.testURL(m, true, returns)
}

func (u *urlTester) testURL(m *tb.Message, full bool, returns []interface{}) {

	var (
		headerString string
		message      string

		method     string
		urlString  string
		statusCode int
		err        error
	)

	u.saveHistory(m)

	method = returns[0].(string)
	urlString = returns[1].(string)
	statusCode = returns[2].(int)

	duration, body, headers, resultCode, _, _, _, expected, err := u.sendRequest(method, urlString, statusCode, "", "")
	log.Println(method, urlString, statusCode, resultCode, expected, err)
	if err != nil {
		u.explainError(m, "", err)
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
