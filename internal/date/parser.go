package date

import (
	"strings"
	"time"
)

func Parse(input string) (*time.Time, error) {
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "none" || input == "" {
		return nil, nil
	}

	now := time.Now()

	switch input {
	case "today":
		d := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, now.Location())
		return &d, nil
	case "tomorrow":
		tomorrow := now.AddDate(0, 0, 1)
		d := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 23, 59, 0, 0, now.Location())
		return &d, nil
	case "next week":
		nextWeek := now.AddDate(0, 0, 7)
		d := time.Date(nextWeek.Year(), nextWeek.Month(), nextWeek.Day(), 9, 0, 0, 0, now.Location())
		return &d, nil
	case "in 3 days":
		future := now.AddDate(0, 0, 3)
		d := time.Date(future.Year(), future.Month(), future.Day(), 23, 59, 0, 0, now.Location())
		return &d, nil
	case "in a week":
		future := now.AddDate(0, 0, 7)
		d := time.Date(future.Year(), future.Month(), future.Day(), 23, 59, 0, 0, now.Location())
		return &d, nil
	}

	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04",
		"01-02-2006",
		"01/02/2006",
		"Jan 2",
		"Jan 2 2006",
		"January 2",
		"January 2 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, input); err == nil {
			if t.Year() == 0 {
				t = t.AddDate(now.Year(), 0, 0)
			}
			return &t, nil
		}
	}

	return nil, nil
}
