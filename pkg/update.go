package pkg

import (
	"errors"
	"fmt"
	"time"
)

// UpdateEmployeeStatus обновляет статус сотрудника по его имени.
func UpdateEmployeeStatus(employees *[]Employee, id int, status string) error {
	for i, employee := range *employees {
		if employee.Id == id {
			(*employees)[i].Status = status
			return nil
		}
	}
	return fmt.Errorf("сотрудник с Id: %d не найден", id)
}

// UpdateEmployeeInList обновляет информацию о сотруднике в списке сотрудников.
func updateEmployeeInList(employees *[]Employee, updatedEmployee *Employee) error {
	for i, emp := range *employees {
		if emp.Name == updatedEmployee.Name {
			(*employees)[i] = *updatedEmployee
			return nil
		}
	}
	return errors.New("сотрудник с именем " + updatedEmployee.Name + " не найден")
}

// AddScheduleToHistory добавляет расписание на неделю в DutyHistoryStorage, чтобы сохранить исторические данные.
// В эту функцию надо передавать слайс только с теми дежурными, кто дежурит на предстоящей неделе.
func AddScheduleToHistory(employees *[]Employee, storage *DutyHistoryStorage) {
	// Сохраняем текущие значения счетчиков в истории
	currentHistory := DutyHistory{
		Date:      nextMonday(), // Дата начала недели для которой сформировали расписание
		Employees: make([]Employee, len(*employees)),
	}
	copy(currentHistory.Employees, *employees)

	// проверить, есть ли в storage.History.Dates дата начала недели для которой сформировали расписание и если есть, то удаляем этот элемент
	for i, history := range storage.History {
		if history.Date == currentHistory.Date {
			storage.History = append(storage.History[:i], storage.History[i+1:]...)
			break
		}
	}
	storage.History = append(storage.History, currentHistory)
}

func ResetDutyCounters(employees *[]Employee, storage *DutyHistoryStorage) bool {
	reseted := false

	// проверить существует ли storage и storage.LastResetDate
	if storage == nil {
		storage = &DutyHistoryStorage{}
	}

	// Если storage.LastResetDate не установлена, то устанавливаем ее на текущую дату
	if storage.LastResetDate.IsZero() {
		storage.LastResetDate = time.Now()
	}

	// Проверяем, прошло ли 90 дней с момента последнего сброса счетчиков
	if time.Since(storage.LastResetDate).Hours() >= 90*24 {
		// Сброс счетчиков для каждого сотрудника
		for i := range *employees {
			(*employees)[i].SupportDutyCount = 0
			(*employees)[i].ExpressDutyCount = 0
			(*employees)[i].InstancesDutyCount = 0
		}

		storage.LastResetDate = time.Now()
		reseted = true
	}

	return reseted
}

func getMaxId(employees *[]Employee) int {
	maxId := 0
	for _, employee := range *employees {
		if employee.Id > maxId {
			maxId = employee.Id
		}
	}
	return maxId
}

func AddNewEmployee(employees *[]Employee, name string) {
	newEmployee := Employee{
		Id:                 getMaxId(employees) + 1,
		Name:               name,
		SupportLastDuty:    time.Now().Add(-14 * 24 * time.Hour).Truncate(24 * time.Hour), // устанавливаем на 14 дней назад
		ReleaseLastDuty:    time.Now().Add(-14 * 24 * time.Hour).Truncate(24 * time.Hour), // устанавливаем на 14 дней назад
		Status:             StatusAvailable,
		SupportDutyCount:   0,
		ExpressDutyCount:   0,
		InstancesDutyCount: 0,
	}

	*employees = append(*employees, newEmployee)
}
