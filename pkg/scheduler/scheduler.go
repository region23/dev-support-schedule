package scheduler

import (
	"dev-support-schedule/pkg/models"
	"errors"
	"fmt"
	"time"

	rt "github.com/ivahaev/russian-time"
)

const (
	DutyTypeInstancesRelease = "instances_release"
	DutyTypeExpressRelease   = "express_release"
	DutyTypeSupport          = "support"
	NoDuty                   = "no_duty"
)

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
func findEmployeeForReleases(employees []models.DutySummary, findForInstances bool, expressEmployee *models.DutyHistory, startDate time.Time) (models.DutyHistory, error) {
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
		if employee.DutyTypeCount == 0 || (employee.LastDutyDate.Valid && startDate.Sub(employee.LastDutyDate.Time).Hours() >= 13*24) {
			return models.DutyHistory{
				UserID:   employee.UserID,
				Name:     employee.Name,
				Nickname: employee.Nickname,
				DutyDate: startDate,
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
func findEmployeeForSupport(employees []models.DutySummary, startDate time.Time) ([]models.DutyHistory, error) {
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
		dutyDate := startDate.AddDate(0, 0, len(assignedEmployees))
		if employee.DutyTypeCount == 0 || (employee.LastDutyDate.Valid && dutyDate.Sub(employee.LastDutyDate.Time).Hours() >= 6*24) {
			selectedEmployee := models.DutyHistory{
				UserID:   employee.UserID,
				Name:     employee.Name,
				Nickname: employee.Nickname,
				DutyDate: dutyDate,
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
func GetSchedule(employees []models.DutySummary, startDate time.Time) (string, []models.DutyHistory, error) {
	endDate := startDate.AddDate(0, 0, 6)

	var schedule []models.DutyHistory

	expressEmployee, err := findEmployeeForReleases(employees, false, nil, startDate)
	if err != nil {
		return "", nil, err
	}

	schedule = append(schedule, expressEmployee)

	instancesEmployee, err := findEmployeeForReleases(employees, true, &expressEmployee, startDate)
	if err != nil {
		return "", nil, err
	}

	schedule = append(schedule, instancesEmployee)

	supportEmployees, err := findEmployeeForSupport(employees, startDate)
	if err != nil {
		return "", nil, err
	}

	schedule = append(schedule, supportEmployees...)

	supportSchedule := ""
	for _, supportEmploye := range supportEmployees {
		supportSchedule += fmt.Sprintf("@%s – %s\n", supportEmploye.Nickname, rt.Time(supportEmploye.DutyDate).Weekday().String())
	}

	result := fmt.Sprintf(
		"Всем привет! 👾\n**Расписание для саппорт и релиз инженеров с %d %s по %d %s**\n**Релизы**\n@%s – Express Release\n@%s – Instances release\n\n**Саппорт**\n%s\n\nЛюбезно сгенерировано автоматически 🤖\nP.S. Если заметите аномалии, дайте знать – алгоритм требует донастройки 😉",
		startDate.Day(), rt.Time(startDate).Month().StringInCase(), endDate.Day(), rt.Time(endDate).Month().StringInCase(),
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
			result += fmt.Sprintf("%s (%d), последнее %s | ", getDutyTypeName(employee.DutyType), employee.DutyTypeCount, employee.LastDutyDate.Time.Format("01.02.2006"))
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
