package scheduler

import (
	"dev-support-schedule/pkg/models"
	"errors"
	"fmt"
	"time"
)

const (
	DutyTypeInstancesRelease = "instances_release"
	DutyTypeExpressRelease   = "express_release"
	DutyTypeSupport          = "support"
	NoDuty                   = "no_duty"
)

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
func GetWeekEndDate() time.Time {
	today := time.Now()
	weekDay := today.Weekday()
	daysToAdd := 7 - int(weekDay)
	return today.AddDate(0, 0, daysToAdd)
}

// filterEmployee определяет, подходит ли сотрудник для дежурства на основе условий.
func filterEmployee(employee models.DutySummary, findForInstances bool, expressEmployee *models.DutyHistory) bool {
	if employee.Status == models.StatusVacation || employee.Status == models.StatusSick || employee.Status == models.StatusFired {
		return false
	}

	if findForInstances && employee.UserID == expressEmployee.UserID {
		return false
	}

	// новенький, кто еще ни разу не дежурил
	if employee.DutyType == NoDuty {
		return true
	}

	if findForInstances && employee.DutyType != DutyTypeInstancesRelease {
		return false
	}

	if !findForInstances && employee.DutyType != DutyTypeExpressRelease {
		return false
	}

	return true
}

// findEmployeeForReleases находит подходящего сотрудника для дежурства типа "Instances release".
func findEmployeeForReleases(employees []models.DutySummary, findForInstances bool, expressEmployee *models.DutyHistory) (models.DutyHistory, error) {
	// на вход получили отсортированный список сотрудников по возрастанию числа дежурств и по дате последнего дежурства

	for _, employee := range employees {
		if !filterEmployee(employee, findForInstances, expressEmployee) {
			continue
		}

		dutyType := DutyTypeExpressRelease

		if findForInstances {
			dutyType = DutyTypeInstancesRelease
		}

		// Проверяем, что прошло не менее 2 недель с момента последнего дежурства сотрудника.
		if employee.DutyTypeCount == 0 || (employee.LastDutyDate.Valid && time.Since(employee.LastDutyDate.Time).Hours() >= 14*24) {
			return models.DutyHistory{
				UserID:   employee.UserID,
				Name:     employee.Name,
				Nickname: employee.Nickname,
				DutyDate: NextMonday(),
				DutyType: dutyType,
			}, nil
		}
	}

	var dtName string
	if findForInstances {
		dtName = "Instances release"
	} else {
		dtName = "Express release"
	}

	return models.DutyHistory{}, fmt.Errorf("не найдено подходящего сотрудника для дежурства '%s'", dtName)
}

// findEmployeeForSupport находит подходящего сотрудника для дежурства типа "Support".
func findEmployeeForSupport(employees []models.DutySummary) ([]models.DutyHistory, error) {
	// Пройдите по отсортированному списку и выберите дежурных для следующей недели
	assignedEmployees := []models.DutyHistory{}

	for _, employee := range employees {
		// Исключаем сотрудника, который болеет, в отпуске или уволен.
		if employee.Status == models.StatusVacation || employee.Status == models.StatusSick || employee.Status == models.StatusFired {
			continue
		}

		if employee.DutyType != DutyTypeSupport && employee.DutyType != NoDuty {
			continue
		}

		// Проверяем, что прошло не менее 7 дней с момента последнего дежурства сотрудника.
		if employee.DutyTypeCount == 0 || (employee.LastDutyDate.Valid && time.Since(employee.LastDutyDate.Time).Hours() >= 7*24) {
			selectedEmployee := models.DutyHistory{
				UserID:   employee.UserID,
				Name:     employee.Name,
				Nickname: employee.Nickname,
				DutyDate: NextMonday().AddDate(0, 0, len(assignedEmployees)),
				DutyType: DutyTypeSupport,
			}

			assignedEmployees = append(assignedEmployees, selectedEmployee)
			if len(assignedEmployees) == 5 { // если мы уже выбрали дежурных на всю неделю, прекращаем выбор
				break
			}
		}
	}

	if len(assignedEmployees) == 5 {
		return assignedEmployees, nil
	}

	return assignedEmployees, errors.New("не найдено подходящего сотрудника для дежурства 'Support'")
}

// GetSchedule формирует и возвращает расписание на предстоящую неделю.
func GetSchedule(employees []models.DutySummary) (string, []models.DutyHistory, error) {
	// узнать какой сегодня день недели и прибавить столько дней, чтобы получить понедельник

	startDate := NextMonday()             // начало следующей недели
	endDate := startDate.AddDate(0, 0, 4) // пятница следующей недели

	var schedule []models.DutyHistory

	expressEmployee, err := findEmployeeForReleases(employees, false, nil)
	if err != nil {
		return "", nil, err
	}

	schedule = append(schedule, expressEmployee)

	instancesEmployee, err := findEmployeeForReleases(employees, true, &expressEmployee)
	if err != nil {
		return "", nil, err
	}

	schedule = append(schedule, instancesEmployee)

	supportEmployees, err := findEmployeeForSupport(employees)
	if err != nil {
		return "", nil, err
	}

	schedule = append(schedule, supportEmployees...)

	supportSchedule := ""
	for _, supportEmploye := range supportEmployees {
		supportSchedule += fmt.Sprintf("%s - %s\n", supportEmploye.Nickname, supportEmploye.DutyDate.Format("Monday"))
	}

	result := fmt.Sprintf(
		"Всем привет! 👾\n**Расписание для саппорт и релиз инженеров с %s по %s**\n**Релизы**\n%s - Express Release\n%s - Instances release\n\n**Саппорт**\n%s\n\nЛюбезно сгенерировано автоматически 🤖\nP.S. Если заметите аномалии, дайте знать - алгоритм требует донастройки 😉",
		startDate.Format("2 January"), endDate.Format("2 January"),
		expressEmployee.Nickname, instancesEmployee.Nickname, supportSchedule)

	return result, schedule, nil
}

func AllEmployees(employees *[]models.DutySummary) string {
	currentUserID := 0
	result := ""

	// предполагаем, что у нас здесь отсортированный список сотрудников по ID
	for _, employee := range *employees {
		if currentUserID != employee.UserID {
			if currentUserID != 0 {
				result += "\n" // print a newline before the next user
			}
			result += fmt.Sprintf("%d. %s (%s) | ", employee.UserID, employee.Name, getStatusEmoji(employee.Status))
			currentUserID = employee.UserID
		}

		if employee.LastDutyDate.Valid {
			result += fmt.Sprintf("%s (%d), последнее %s | ", getDutyTypeName(employee.DutyType), employee.DutyTypeCount, employee.LastDutyDate.Time.Format("2006.01.02"))
		} else {
			result += getDutyTypeName(employee.DutyType)
		}
	}

	return result
}

func getStatusEmoji(status string) string {
	switch status {
	case models.StatusSick:
		return "болеет 🤒"
	case models.StatusVacation:
		return "в отпуске 🏖"
	default:
		return "доступен 👍"
	}
}

func getDutyTypeName(dutyType string) string {
	switch dutyType {
	case DutyTypeInstancesRelease:
		return "Instances release"
	case DutyTypeExpressRelease:
		return "Express release"
	case NoDuty:
		return "Не дежурил"
	default:
		return "Support"
	}
}
