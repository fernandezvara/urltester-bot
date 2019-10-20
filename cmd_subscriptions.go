package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func (u *urlTester) subscribe(m *tb.Message, returns []interface{}) {

	var (
		sched schedule

		id  int
		err error
	)

	u.saveHistory(m)

	id = returns[0].(int)

	sched, err = u.getScheduleByID(id)
	if err != nil {
		u.bot.Send(m.Sender, err.Error())
		return
	}

	if alreadyOnIntArray(sched.Subscriptors, m.Sender.ID) {
		u.bot.Send(m.Sender, "Already subscribed.")
		return
	}

	sched.Subscriptors = append(sched.Subscriptors, m.Sender.ID)

	err = u.db.Save(&sched)
	if err != nil {
		u.explainError(m, "", err)
		return
	}

	u.bot.Send(m.Sender, "Successfully subscribed.")

}

func (u *urlTester) unsubscribe(m *tb.Message, returns []interface{}) {

	var (
		sched schedule

		id  int
		err error
	)

	u.saveHistory(m)

	id = returns[0].(int)

	sched, err = u.getScheduleByID(id)
	if err != nil {
		u.bot.Send(m.Sender, err.Error())
		return
	}

	if alreadyOnIntArray(sched.Subscriptors, m.Sender.ID) {
		// remove from array
		sched.Subscriptors = removeFromIntArray(sched.Subscriptors, m.Sender.ID)
		err = u.db.Save(&sched)
		if err != nil {
			u.explainError(m, "", err)
			return
		}
		u.bot.Send(m.Sender, "Unsubscribed.")
		return

	}

	u.bot.Send(m.Sender, "Not subscribed to the requested monitor.")

}
