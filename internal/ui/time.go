package ui

import "time"

func FormatDate(date time.Time) string {
	formattedDate := date.Format("Jan 2, 2006")

	// If today, return "Today"
	if formattedDate == time.Now().Format("Jan 2, 2006") {
		return "Today"
	}

	// If yesterday, return "Yesterday"
	if formattedDate == time.Now().AddDate(0, 0, -1).Format("Jan 2, 2006") {
		return "Yesterday"
	}

	return formattedDate
}
