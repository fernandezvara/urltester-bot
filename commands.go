package main

const (
	helpMonitorID  = "Monitor ID"
	helpStatusCode = "Desired HTTP status code"
	helpInterval   = "Time between checks. Format <amount><unit>. Units = s: Second, m: Minute, h: Hour. Example: 1m"
	helpMethod     = "HTTP Method to use."
	helpURL        = "Full URL to monitor. Ex: https://google.com "
	helpPrivate    = "Sets the monitor as private to the owner."
	helpText       = "Text to search for on request's body"
	helpTimeout    = "Timeot that will trigger a monitor"
)

func (u *urlTester) buildCommands() (commands map[string]command) {

	commands = make(map[string]command)

	commands["/start"] = command{
		fn:        u.start,
		isPrivate: true,
		forUsers:  false,
		forAdmins: false,
		helpShort: "starts the bot asking for permissions from an administrator",
		helpLong:  "starts the bot asking for permissions from an administrator",
		payload:   []payloadPart{},
	}

	commands["/help"] = command{
		fn:        u.help,
		isPrivate: true,
		forUsers:  false,
		forAdmins: false,
		helpShort: "this help",
		helpLong:  "all the commands available for the current user are shown",
		payload:   []payloadPart{},
	}

	commands["/history"] = command{
		fn:        u.history,
		isPrivate: true,
		forUsers:  false,
		forAdmins: false,
		helpShort: "shows last commands sent to the bot",
		helpLong:  "shows last commands sent to the bot",
		payload:   []payloadPart{},
	}

	commands["/summary"] = command{
		fn:        u.summary,
		isPrivate: true,
		forUsers:  true,
		forAdmins: false,
		helpShort: "shows current status of monitors subscribed to",
		helpLong:  "shows current status of monitors subscribed to",
		payload:   []payloadPart{},
	}

	commands["/monitors"] = command{
		fn:        u.monitors,
		isPrivate: true,
		forUsers:  true,
		forAdmins: false,
		helpShort: "shows current defined monitors",
		helpLong:  "shows current defined monitors",
		payload:   []payloadPart{},
	}

	commands["/newmonitor"] = command{
		fn:        u.newmonitor,
		isPrivate: true,
		forUsers:  true,
		forAdmins: false,
		helpShort: "creates a new monitor with basic monitor settings",
		helpLong:  "creates a new monitor with basic monitor settings",
		payload: []payloadPart{
			payloadPart{
				arg:   "method",
				typ:   typeString,
				help:  helpMethod,
				valid: allowedMethods,
			},
			payloadPart{
				arg:  "url",
				typ:  typeString,
				help: helpURL,
			},
			payloadPart{
				arg:  "statuscode",
				typ:  typeInt,
				help: helpStatusCode,
			},
			payloadPart{
				arg:  "interval",
				typ:  typeTimeExp,
				help: helpInterval,
			},
			payloadPart{
				arg:  "private",
				typ:  typeBool,
				help: helpPrivate,
			},
		},
	}

	commands["/remove"] = command{
		fn:        u.remove,
		isPrivate: true,
		forUsers:  true,
		forAdmins: false,
		helpShort: "removes a monitor",
		helpLong:  "removes a monitor",
		payload: []payloadPart{
			payloadPart{
				arg:  "id",
				typ:  typeInt,
				help: helpMonitorID,
			},
		},
	}

	commands["/setinterval"] = command{
		fn:        u.setinterval,
		isPrivate: true,
		forUsers:  true,
		forAdmins: false,
		helpShort: "updates the interval between monitor calls",
		helpLong:  "updates the interval between monitor calls",
		payload: []payloadPart{
			payloadPart{
				arg:  "id",
				typ:  typeInt,
				help: helpMonitorID,
			},
			payloadPart{
				arg:  "interval",
				typ:  typeTimeExp,
				help: helpInterval,
			},
		},
	}

	commands["/setstatuscode"] = command{
		fn:        u.setstatuscode,
		isPrivate: true,
		forUsers:  true,
		forAdmins: false,
		helpShort: "updates the expected HTTP status code for the monitor",
		helpLong:  "updates the expected HTTP status code for the monitor",
		payload: []payloadPart{
			payloadPart{
				arg:  "id",
				typ:  typeInt,
				help: helpMonitorID,
			},
			payloadPart{
				arg:  "statuscode",
				typ:  typeInt,
				help: helpStatusCode,
			},
		},
	}

	commands["/settext"] = command{
		fn:        u.settext,
		isPrivate: true,
		forUsers:  true,
		forAdmins: false,
		helpShort: "updates the expected text to be found on every monitor request",
		helpLong:  "updates the expected text to be found on every monitor request",
		payload: []payloadPart{
			payloadPart{
				arg:  "id",
				typ:  typeInt,
				help: helpMonitorID,
			},
			payloadPart{
				arg:  "text",
				typ:  typeString,
				help: helpText,
			},
		},
	}

	commands["/settimeout"] = command{
		fn:        u.settimeout,
		isPrivate: true,
		forUsers:  true,
		forAdmins: false,
		helpShort: "updates the expected timeout to be found on every monitor request",
		helpLong:  "updates the expected timeout to be found on every monitor request",
		payload: []payloadPart{
			payloadPart{
				arg:  "id",
				typ:  typeInt,
				help: helpMonitorID,
			},
			payloadPart{
				arg:  "timeout",
				typ:  typeTimeExp,
				help: helpTimeout,
			},
		},
	}

	commands["/subscribe"] = command{
		fn:        u.subscribe,
		isPrivate: true,
		forUsers:  true,
		forAdmins: false,
		helpShort: "subscribes to monitor status",
		helpLong:  "subscribes to monitor status",
		payload: []payloadPart{
			payloadPart{
				arg:  "id",
				typ:  typeInt,
				help: helpMonitorID,
			},
		},
	}

	commands["/unsubscribe"] = command{
		fn:        u.unsubscribe,
		isPrivate: true,
		forUsers:  true,
		forAdmins: false,
		helpShort: "unsubscribes from monitor status",
		helpLong:  "unsubscribes from monitor status",
		payload: []payloadPart{
			payloadPart{
				arg:  "id",
				typ:  typeInt,
				help: helpMonitorID,
			},
		},
	}

	commands["/test"] = command{
		fn:        u.test,
		isPrivate: true,
		forUsers:  true,
		forAdmins: false,
		helpShort: "test a URL and its HTTP status",
		helpLong:  "test a URL and its HTTP status",
		payload: []payloadPart{
			payloadPart{
				arg:   "method",
				typ:   typeString,
				help:  helpMethod,
				valid: allowedMethods,
			},
			payloadPart{
				arg:  "url",
				typ:  typeString,
				help: helpURL,
			},
			payloadPart{
				arg:  "statuscode",
				typ:  typeInt,
				help: helpURL,
			},
		},
	}

	commands["/testfull"] = command{
		fn:        u.testFull,
		isPrivate: true,
		forUsers:  true,
		forAdmins: false,
		helpShort: "test a URL and its HTTP status, returning its data",
		helpLong:  "test a URL and its HTTP status, returning its data",
		payload: []payloadPart{
			payloadPart{
				arg:   "method",
				typ:   typeString,
				help:  helpMethod,
				valid: allowedMethods,
			},
			payloadPart{
				arg:  "url",
				typ:  typeString,
				help: helpURL,
			},
			payloadPart{
				arg:  "statuscode",
				typ:  typeInt,
				help: helpURL,
			},
		},
	}

	commands["/users"] = command{
		fn:        u.users,
		isPrivate: true,
		forUsers:  false,
		forAdmins: true,
		helpShort: "returns users and its current status",
		helpLong:  "returns users and its current status",
		payload:   []payloadPart{},
	}

	commands["/grant"] = command{
		fn:        u.grant,
		isPrivate: true,
		forUsers:  false,
		forAdmins: true,
		helpShort: "grants access to a user",
		helpLong:  "grants access to a user",
		payload: []payloadPart{
			payloadPart{
				arg:  "id",
				typ:  typeInt,
				help: helpMonitorID,
			},
		},
	}

	commands["/revoke"] = command{
		fn:        u.revoke,
		isPrivate: true,
		forUsers:  false,
		forAdmins: true,
		helpShort: "revokes access to a user",
		helpLong:  "revokes access to a user",
		payload: []payloadPart{
			payloadPart{
				arg:  "id",
				typ:  typeInt,
				help: helpMonitorID,
			},
		},
	}
	return commands

}
