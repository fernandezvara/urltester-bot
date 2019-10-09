package main

import (
	"fmt"
	"math"
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
	months := math.Floor(float64(input) / 60 / 60 / 24 / 7 / 30)
	days := math.Floor(float64(input) / 60 / 60 / 24)
	hours := math.Floor(float64(input) / 60 / 60)
	minutes := math.Floor(float64(input) / 60)
	seconds := input % 60

	if years > 0 {
		result = fmt.Sprintf("%s, ", plural(int64(years), "year"))
	}
	if months > 0 {
		result = fmt.Sprintf("%s%s, ", result, plural(int64(months), "month"))
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

	// if years > 0 {
	// 	result = plural(int64(years), "year") + plural(int64(months), "month") + plural(int64(weeks), "week") + plural(int64(days), "day") + plural(int64(hours), "hour") + plural(int64(minutes), "minute") + plural(int64(seconds), "second")
	// } else if months > 0 {
	// 	result = plural(int64(months), "month") + plural(int64(weeks), "week") + plural(int64(days), "day") + plural(int64(hours), "hour") + plural(int64(minutes), "minute") + plural(int64(seconds), "second")
	// } else if weeks > 0 {
	// 	result = plural(int64(weeks), "week") + plural(int64(days), "day") + plural(int64(hours), "hour") + plural(int64(minutes), "minute") + plural(int64(seconds), "second")
	// } else if days > 0 {
	// 	result = plural(int64(days), "day") + plural(int64(hours), "hour") + plural(int64(minutes), "minute") + plural(int64(seconds), "second")
	// } else if hours > 0 {
	// 	result = plural(int64(hours), "hour") + plural(int64(minutes), "minute") + plural(int64(seconds), "second")
	// } else if minutes > 0 {
	// 	result = plural(int64(minutes), "minute") + plural(int64(seconds), "second")
	// } else {
	// 	result = plural(int64(seconds), "second")
	// }

	return
}
