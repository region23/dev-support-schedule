package utils

import "time"

func NextMonday() time.Time {
	today := time.Now().Weekday()
	var daysToAdd int

	switch today {
	case time.Monday:
		daysToAdd = 0
	case time.Tuesday:
		daysToAdd = 6
	case time.Wednesday:
		daysToAdd = 5
	case time.Thursday:
		daysToAdd = 4
	case time.Friday:
		daysToAdd = 3
	case time.Saturday:
		daysToAdd = 2
	case time.Sunday:
		daysToAdd = 1
	}

	nextMonday := time.Now().AddDate(0, 0, daysToAdd)
	return nextMonday.Truncate(24 * time.Hour)
}

// getQuarterStartDate возвращает дату начала текущего квартала.
func GetQuarterStartDate() time.Time {
	// получить текущую дату
	today := time.Now()

	// определить квартал
	quarter := (today.Month() - 1) / 3

	// вернуть дату начала квартала
	return time.Date(today.Year(), time.Month(quarter*3+1), 1, 0, 0, 0, 0, time.UTC)

}

// вернуть дату конца текущей недели
func GetCurrentWeekEndDate() time.Time {
	today := time.Now()
	weekDay := today.Weekday()
	daysToAdd := 7 - int(weekDay)
	return today.AddDate(0, 0, daysToAdd)
}

// GetWeekStartDate возвращает дату начала недели
func GetWeekStartDate(dt time.Time) time.Time {
	// Week starts from Sunday in Go, so Monday is 1.
	// If it's Sunday, we need to go back 6 days to get to previous Monday.
	// Otherwise, just go back by (dt.Weekday() - 1) days to get to current week's Monday.
	offset := int(dt.Weekday()) - 1
	if offset < 0 {
		offset = 6
	}
	return dt.AddDate(0, 0, -offset)
}
