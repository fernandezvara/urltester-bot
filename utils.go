package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/asdine/storm"
	tb "gopkg.in/tucnak/telebot.v2"
)

func plural(count int64, singular string) (result string) {
	if (count == 1) || (count == 0) {
		result = fmt.Sprintf("%d %s", count, singular)
		return
	}

	result = fmt.Sprintf("%d %ss", count, singular)

	return
}

func secondsToHuman(input int64) (result string) {

	years := math.Floor(float64(input) / 60 / 60 / 24 / 7 / 30 / 12)
	seconds := input % (60 * 60 * 24 * 7 * 30 * 12)
	months := math.Floor(float64(seconds) / 60 / 60 / 24 / 7 / 30)
	seconds = input % (60 * 60 * 24 * 7 * 30)
	weeks := math.Floor(float64(seconds) / 60 / 60 / 24 / 7)
	seconds = input % (60 * 60 * 24 * 7)
	days := math.Floor(float64(seconds) / 60 / 60 / 24)
	seconds = input % (60 * 60 * 24)
	hours := math.Floor(float64(seconds) / 60 / 60)
	seconds = input % (60 * 60)
	minutes := math.Floor(float64(seconds) / 60)
	seconds = input % 60

	if years > 0 {
		result = fmt.Sprintf("%s, ", plural(int64(years), "year"))
	}
	if months > 0 {
		result = fmt.Sprintf("%s%s, ", result, plural(int64(months), "month"))
	}

	if weeks > 0 {
		result = fmt.Sprintf("%s%s, ", result, plural(int64(weeks), "week"))
	}

	if days > 0 {
		result = fmt.Sprintf("%s%s, ", result, plural(int64(days), "day"))
	}

	if hours > 0 {
		result = fmt.Sprintf("%s%s, ", result, plural(int64(hours), "hour"))
	}

	if minutes > 0 {
		result = fmt.Sprintf("%s%s, ", result, plural(int64(minutes), "minute"))
	}

	if seconds > 0 {
		result = fmt.Sprintf("%s%s", result, plural(int64(seconds), "second"))
	}

	return
}

func alreadyOnIntArray(arr []int, value int) bool {

	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false

}

func alreadyOnStringArray(arr []string, value string) bool {

	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false

}

func arrStringToString(arr []string) (message string) {

	for _, v := range arr {
		if len(message) == 0 {
			message = fmt.Sprintf("'%s'", v)
		} else {
			message = fmt.Sprintf("%s, '%s'", message, v)
		}
	}

	return
}

func removeFromIntArray(arr []int, value int) (newArr []int) {

	for _, v := range arr {
		if v != value {
			newArr = append(newArr, v)
		}
	}
	return

}

func statusText(id int) string {
	switch id {
	case statusDown:
		return "DOWN"
	case statusUp:
		return "UP"
	case statusStarted:
		return "STARTED"
	case statusStopped:
		return "STOPPED"
	default:
		return ""
	}
}

func (u *urlTester) accessGranted(tbUser *tb.User) (authorized bool) {

	_, authorized = u.userInfo(tbUser.ID)
	return

}

func (u *urlTester) userInfo(id int) (tgUser user, authorized bool) {

	var (
		err error
	)

	err = u.db.One("ID", id, &tgUser)
	// on error user will be unauthorized (if the user didn't /start'ed)
	if err != nil {
		u.bot.Send(telegramUser{id}, "Access not allowed. \nPlease use /start to ask for permissions.")
		return
	}

	authorized = tgUser.Authorized
	return

}

// isUserAdmin returns if the current user is allowed as administrator
func (u *urlTester) isUserAdmin(id int) bool {

	for _, adminID := range u.admins {
		if adminID == id {
			return true
		}
	}

	return false

}

func (u *urlTester) sendMessageAndNotifyAdmins(userID int, message string) {

	log.Println(userID, ":", message)
	u.bot.Send(telegramUser{userID}, message)
	u.sendMessageToAdmins(message)

}

func (u *urlTester) sendMessageToAdmins(message string) {

	for _, adminID := range u.admins {
		u.bot.Send(telegramUser{adminID}, message)
	}

}

func (u *urlTester) getScheduleByID(id int) (sched schedule, err error) {

	err = u.db.One("ID", id, &sched)
	if err != nil {
		if err == storm.ErrNotFound {
			err = errors.New("ID not found")
			return
		}
		err = errors.New("unexpected error")
		return
	}

	return

}

// evaluateTimeExp verifies that the expression to evaluate meets the requirements
// time expressions:
// (n)s = n seconds
// (n)m = n minutes
// (n)h = n hours
func evaluateTimeExp(exp string) (int, string, bool) {

	var matcher = regexp.MustCompile(`^([0-9]+)([a-zA-Z])$`)

	parts := matcher.FindAllStringSubmatch(exp, -1)

	if len(parts) == 1 {
		if len(parts[0]) == 3 {

			i, _ := strconv.Atoi(parts[0][1])
			switch parts[0][2] {
			case "s", "S", "m", "M", "h", "H":
				return i, strings.ToLower(parts[0][2]), true
			default:
				return 0, "", false
			}

		}
	}
	return 0, "", false

}

func headersToString(headers map[string]string) (headersString string) {

	for k, v := range headers {
		headersString = fmt.Sprintf("%s%s: %s\n", headersString, k, v)
	}
	return

}
