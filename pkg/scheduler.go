package pkg

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

// findEmployeeForInstances находит подходящего сотрудника для дежурства типа "Instances release".
func findEmployeeForInstances(employees *[]Employee, expressEmployee *Employee) (Employee, error) {
	// Сортировка списка сотрудников сначала по числу дежурств, затем по дате последнего дежурства.
	sort.Slice(*employees, func(i, j int) bool {
		if (*employees)[i].InstancesDutyCount == (*employees)[j].InstancesDutyCount {
			return (*employees)[i].ReleaseLastDuty.Before((*employees)[j].ReleaseLastDuty)
		}
		return (*employees)[i].InstancesDutyCount < (*employees)[j].InstancesDutyCount
	})

	for _, employee := range *employees {
		// Исключаем сотрудника, который болеет, в отпуске или уволен.
		if employee.Status == StatusVacation || employee.Status == StatusSick || employee.Status == StatusFired {
			continue
		}

		// Исключаем сотрудника, который был назначен на "Express Release".
		if employee.Id == expressEmployee.Id {
			continue
		}

		// Проверяем, что прошло не менее 2 недель с момента последнего дежурства сотрудника.
		if time.Since(employee.ReleaseLastDuty).Hours() >= 14*24 {
			employee.ReleaseLastDuty = nextMonday().AddDate(0, 0, 3)
			employee.InstancesDutyCount++
			return employee, nil
		}
	}

	// Если дошли до конца списка и никого не подобрали тогда берем первого, с наименьшим количеством дежурств в Instances release
	// Сортировка списка сотрудников по возрастанию числа дежурств.
	sort.Slice(*employees, func(i, j int) bool {
		return (*employees)[i].InstancesDutyCount < (*employees)[j].InstancesDutyCount
	})

	for _, employee := range *employees {
		// Исключаем сотрудника, который болеет, в отпуске или уволен.
		if employee.Status == StatusVacation || employee.Status == StatusSick || employee.Status == StatusFired {
			continue
		}

		// Исключаем сотрудника, который был назначен на "Express Release".
		if employee.Id == expressEmployee.Id {
			continue
		}

		employee.ReleaseLastDuty = nextMonday().AddDate(0, 0, 3)
		employee.InstancesDutyCount++
		return employee, nil
	}

	return Employee{}, errors.New("не найдено подходящего сотрудника для дежурства 'Instances release'")
}

// findEmployeeForSupport находит подходящего сотрудника для дежурства типа "Support".
func findEmployeeForSupport(employees *[]Employee, dayInWeek int) (Employee, error) {
	// Сортировка списка сотрудников сначала по числу дежурств, затем по дате последнего дежурства.
	sort.Slice(*employees, func(i, j int) bool {
		if (*employees)[i].SupportDutyCount == (*employees)[j].SupportDutyCount {
			return (*employees)[i].SupportLastDuty.Before((*employees)[j].SupportLastDuty)
		}
		return (*employees)[i].SupportDutyCount < (*employees)[j].SupportDutyCount
	})

	for i, employee := range *employees {
		// Исключаем сотрудника, который болеет, в отпуске или уволен.
		if employee.Status == StatusVacation || employee.Status == StatusSick || employee.Status == StatusFired {
			continue
		}

		// Проверяем, что прошло не менее 7 дней с момента последнего дежурства сотрудника.
		if time.Since(employee.SupportLastDuty).Hours() >= 7*24 {
			employee.SupportLastDuty = nextMonday().AddDate(0, 0, dayInWeek)
			employee.SupportDutyCount = employee.SupportDutyCount + 2
			// и удаляем его из списка сотрудников, чтобы он не попал в дежурство на следующий день
			*employees = append((*employees)[:i], (*employees)[i+1:]...)
			return employee, nil
		}
	}

	// Если дошли до конца списка и никого не подобрали тогда берем первого, с наименьшим количеством дежурств в Support
	// Сортировка списка сотрудников по возрастанию числа дежурств.
	sort.Slice(*employees, func(i, j int) bool {
		return (*employees)[i].SupportDutyCount < (*employees)[j].SupportDutyCount
	})

	for i, employee := range *employees {
		// Исключаем сотрудника, который болеет, в отпуске или уволен.
		if employee.Status == StatusVacation || employee.Status == StatusSick || employee.Status == StatusFired {
			continue
		}

		employee.SupportLastDuty = nextMonday().AddDate(0, 0, dayInWeek)
		employee.SupportDutyCount = employee.SupportDutyCount + 2
		// и удаляем его из списка сотрудников, чтобы он не попал в дежурство на следующий день
		*employees = append((*employees)[:i], (*employees)[i+1:]...)
		return employee, nil
	}

	return Employee{}, errors.New("не найдено подходящего сотрудника для дежурства 'Support'")
}

// findEmployeeForExpress находит подходящего сотрудника для дежурства типа "Express Release".
func findEmployeeForExpress(employees *[]Employee) (Employee, error) {
	// Сортировка списка сотрудников сначала по числу дежурств, затем по дате последнего дежурства.
	sort.Slice(*employees, func(i, j int) bool {
		if (*employees)[i].ExpressDutyCount == (*employees)[j].ExpressDutyCount {
			return (*employees)[i].ReleaseLastDuty.Before((*employees)[j].ReleaseLastDuty)
		}
		return (*employees)[i].ExpressDutyCount < (*employees)[j].ExpressDutyCount
	})

	for _, employee := range *employees {
		// Исключаем сотрудника, который болеет, в отпуске или уволен.
		if employee.Status == StatusVacation || employee.Status == StatusSick || employee.Status == StatusFired {
			continue
		}

		// Проверяем, что прошло не менее 2 недель с момента последнего дежурства сотрудника.
		if time.Since(employee.ReleaseLastDuty).Hours() >= 14*24 {
			employee.ReleaseLastDuty = nextMonday().AddDate(0, 0, 3)
			employee.ExpressDutyCount = employee.ExpressDutyCount + 2
			return employee, nil
		}
	}

	// Если дошли до конца списка и никого не подобрали тогда берем первого, с наименьшим количеством дежурств в Express release
	// Сортировка списка сотрудников по возрастанию числа дежурств.
	sort.Slice(*employees, func(i, j int) bool {
		return (*employees)[i].ExpressDutyCount < (*employees)[j].ExpressDutyCount
	})

	for _, employee := range *employees {
		// Исключаем сотрудника, который болеет, в отпуске или уволен.
		if employee.Status == StatusVacation || employee.Status == StatusSick || employee.Status == StatusFired {
			continue
		}

		employee.ReleaseLastDuty = nextMonday().AddDate(0, 0, 3)
		employee.ExpressDutyCount++
		return employee, nil
	}

	return Employee{}, errors.New("не найдено подходящего сотрудника для дежурства 'Express Release'")
}

func nextMonday() time.Time {
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

// GetSchedule формирует и возвращает расписание на предстоящую неделю.
func GetSchedule(employees *[]Employee) (string, *[]Employee, error) {
	// узнать какой сегодня день недели и прибавить столько дней, чтобы получить понедельник

	startDate := nextMonday()             // начало следующей недели
	endDate := startDate.AddDate(0, 0, 4) // пятница следующей недели

	fmt.Println("startDate: ", startDate, " | endDate: ", endDate)

	var schedule []Employee

	expressEmployee, err := findEmployeeForExpress(employees)
	if err != nil {
		return "", nil, err
	}
	fmt.Println("---expressEmployee---")
	fmt.Println(*employees)
	fmt.Println(expressEmployee)

	instancesEmployee, err := findEmployeeForInstances(employees, &expressEmployee)
	if err != nil {
		return "", nil, err
	}
	fmt.Println("---instancesEmployee---")
	fmt.Println(*employees)
	fmt.Println(instancesEmployee)

	// Обновляем статусы и счетчики для выбранных дежурных
	err = updateEmployeeInList(employees, &expressEmployee)
	if err != nil {
		return "", nil, err
	}
	expressEmployee.SupportDutyCount = 0
	expressEmployee.ExpressDutyCount = 1
	expressEmployee.InstancesDutyCount = 0
	schedule = append(schedule, expressEmployee)

	err = updateEmployeeInList(employees, &instancesEmployee)
	if err != nil {
		return "", nil, err
	}
	instancesEmployee.SupportDutyCount = 0
	instancesEmployee.ExpressDutyCount = 0
	instancesEmployee.InstancesDutyCount = 1
	schedule = append(schedule, instancesEmployee)

	supportSchedule := ""
	weekdays := [5]string{"понедельник", "вторник", "среда", "четверг", "пятница"}

	// создаем копию списка сотрудников, чтобы не менять исходный
	employeesCopy := make([]Employee, len(*employees))
	copy(employeesCopy, *employees)

	for dayInWeek, day := range weekdays {
		supportEmployee, err := findEmployeeForSupport(&employeesCopy, dayInWeek)
		if err != nil {
			return "", nil, err
		}
		fmt.Println("---supportEmployee---")
		fmt.Println(*employees)
		fmt.Println(supportEmployee)

		err = updateEmployeeInList(employees, &supportEmployee)
		if err != nil {
			return "", nil, err
		}
		supportEmployee.SupportDutyCount = 1
		supportEmployee.ExpressDutyCount = 0
		supportEmployee.InstancesDutyCount = 0
		schedule = append(schedule, supportEmployee)
		supportSchedule += fmt.Sprintf("%s - %s\n", supportEmployee.Name, day)
	}

	result := fmt.Sprintf(
		"Всем привет! 👾\n**Расписание для саппорт и релиз инженеров с %s по %s**\n**Релизы**\n%s - Express Release\n%s - Instances release\n\n**Саппорт**\n%s\n\nЛюбезно сгенерировано автоматически 🤖\nP.S. Если заметите аномалии, дайте знать - алгоритм требует донастройки 😉",
		startDate.Format("2 January"), endDate.Format("2 January"),
		expressEmployee.Name, instancesEmployee.Name, supportSchedule)

	return result, &schedule, nil
}

// AllEmployees возвращает отформатированную строку со списком всех сотрудников, с их статусами и счетчиками дежурств.
func AllEmployees(employees *[]Employee) string {
	result := ""

	for _, employee := range *employees {
		if employee.Status == StatusFired {
			continue
		}

		result += fmt.Sprintf(
			"%d. %s – %s | Support: %d | Instances release: %d | Express Release: %d\n",
			employee.Id, employee.Name, employee.Status, employee.SupportDutyCount, employee.InstancesDutyCount, employee.ExpressDutyCount)
	}

	return result
}
