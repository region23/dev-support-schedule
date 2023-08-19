package handlers

import (
	"database/sql"
	"dev-support-schedule/pkg/models"
	"dev-support-schedule/pkg/scheduler"
	"dev-support-schedule/pkg/storage"
	"dev-support-schedule/pkg/utils"
	"log"
	"time"
)

type CommandHandler struct {
	db *sql.DB
}

func NewCommandHandler(db *sql.DB) *CommandHandler {
	return &CommandHandler{
		db: db,
	}
}

func (ch *CommandHandler) GenerateSchedule(save bool, weekStartDateStr string) string {
	// получить из базы расписание
	employees, err := storage.LoadEmployees(ch.db)
	if err != nil {
		// Логирование ошибки
		log.Println(err)
		return "Не удалось загрузить список сотрудников"
	}

	if len(employees) == 0 {
		return "Список сотрудников пуст\nДобавь сотрудников с помощью команды */schedule add @nickname ФИО*"
	}

	// определить какой сейчас квартал и вернуть дату начала квартала
	quarterStartDate := utils.GetQuarterStartDate()

	var startDate time.Time

	if weekStartDateStr != "" {
		startDate, err = time.Parse("2006-01-02", weekStartDateStr)
		if err != nil {
			// Логирование ошибки
			log.Println(err)
			return "Неверный формат даты начала недели"
		}
		startDate = utils.GetWeekStartDate(startDate)
	} else {
		// если дата начала недели не задана, то начало следующей недели
		startDate = utils.NextMonday()
	}

	weekEndDate := startDate.AddDate(0, 0, 6) // пятница следующей недели

	dutySummary, err := storage.GetDutySummary(ch.db, false, quarterStartDate, weekEndDate)
	if err != nil {
		// Логирование ошибки
		log.Println(err)
		return "Не удалось загрузить список сотрудников"
	}

	messageForBot, schedule, err := scheduler.GetSchedule(dutySummary, startDate)
	if err != nil {
		// Логирование ошибки
		log.Println(err)
		return "Не удалось сформировать расписание"
	}

	if save {
		// сохранить в базу расписание
		err = storage.AddScheduleToDutyHistory(ch.db, schedule)
		if err != nil {
			// Логирование ошибки
			log.Println(err)
			return "Не удалось сохранить расписание"
		}

		messageForBot += "\n\n✅ *Расписание сохранено в базу*"
	}

	return messageForBot
}

func (ch *CommandHandler) AllEmployees() string {
	// определить какой сейчас квартал и вернуть дату начала квартала
	quarterStartDate := utils.GetQuarterStartDate()
	weekEndDate := utils.GetCurrentWeekEndDate()

	dutyStats, err := storage.GetDutySummary(ch.db, true, quarterStartDate, weekEndDate)
	if err != nil {
		// Логирование ошибки
		log.Println(err)
		return "Не удалось загрузить список сотрудников"
	}

	employees := scheduler.AllEmployees(&dutyStats)

	if len(employees) == 0 {
		return "\nСписок сотрудников пуст\nДобавь сотрудников с помощью команды */schedule add @nickname ФИО*"
	}

	return employees
}

func (ch *CommandHandler) AddEmployee(nickname string, name string) string {
	emp := models.Employee{
		Nickname: nickname,
		Name:     name,
	}

	err := storage.AddEmployee(ch.db, emp)
	if err != nil {
		// Логирование ошибки
		log.Println(err)
		return "Не удалось добавить сотрудника"
	}

	return "Сотрудник добавлен"
}

func (ch *CommandHandler) UpdateEmployeeStatus(cmd models.Command) string {
	employees := make([]models.Employee, 0, len(cmd.Statuses))

	for nickname, status := range cmd.Statuses {
		emp := models.Employee{
			Nickname: nickname,
			Status:   status,
		}
		employees = append(employees, emp)
	}

	err := storage.UpdateEmployees(ch.db, employees)
	if err != nil {
		// Логирование ошибки
		log.Println(err)
		return "Не удалось обновить статусы сотрудников"
	}

	return "Статусы сотрудников обновлены"
}
