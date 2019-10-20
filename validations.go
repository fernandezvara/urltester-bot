package main

import (
	"fmt"
	"strconv"
	"strings"
)

func (u *urlTester) payloadReader(text string) (cmdString string, returns []interface{}, err error) {

	var (
		textParts []string
		parts     []string
		expects   []payloadPart
	)

	// remove double whitespaces before split
	textParts = strings.Split(strings.Join(strings.Fields(strings.TrimSpace(text)), " "), " ")
	cmdString = textParts[0]
	for i, p := range textParts {
		if i != 0 {
			parts = append(parts, p)
		}
	}

	_, ok := u.commands[cmdString]
	if !ok {
		err = errCommandNotFound
		return
	}

	if len(textParts) > 1 && strings.ToUpper(textParts[1]) == "HELP" {
		err = errInvalidPayload
		return
	}

	expects = u.commands[cmdString].payload
	if len(parts) != len(expects) {
		err = errInvalidPayload
		return
	}

	for index, part := range parts {

		// check type

		switch expects[index].typ {
		case typeString:
			// part is already a string, so no need conversion
			// validate input
			if len(expects[index].valid) > 0 {
				if !alreadyOnStringArray(expects[index].valid, strings.ToUpper(part)) {
					err = fmt.Errorf("Error: Key '%s': '%s' not valid.\n Allowed values: [%s]", expects[index].arg, part, arrStringToString(expects[index].valid))
					return
				}
			}

			// seems valid
			returns = append(returns, part)

		case typeTimeExp:

			// validate expression
			var ok bool
			_, _, ok = evaluateTimeExp(part)
			if !ok {
				err = fmt.Errorf("Error: Key '%s': '%s' not a valid time interval expression.\n Example: '1m' for 1 minute\n Allowed values: \n - s: seconds\n - m: minutes\n - h: hours", expects[index].arg, part)
				return
			}

			returns = append(returns, part)

		case typeInt:

			// is int?
			var intValue int
			intValue, err = strconv.Atoi(part)
			if err != nil {
				err = fmt.Errorf("Error: Key '%s': '%s' is not a int value", expects[index].arg, part)
				return
			}

			// seems valid
			returns = append(returns, intValue)

		case typeBool:

			// is bool?
			var boolValue bool
			switch part {
			case "true":
				boolValue = true
			case "false":
				// do not make anything, it's already false
			default:
				err = fmt.Errorf("Error: Key '%s': '%s' is not a boolean value", expects[index].arg, part)
				return
			}

			// seems valid
			returns = append(returns, boolValue)

		}

	}

	return
}
