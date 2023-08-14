package main

import (
	"dev-support-schedule/pkg"
	"fmt"
	"os"
)

const (
	employeesFilePath = "data/employees.json"
	historyFilePath   = "data/history.json"
)

func displayMenu() {
	fmt.Printf("\n--------------------------------------------------------------------------------------\n")
	fmt.Println("Добро пожаловать в программу расписания дежурств!")
	fmt.Println("Выберите действие:")
	fmt.Println("1. Сформировать расписание на следующую неделю")
	fmt.Println("2. Обновить статус сотрудников (на больничном, в отпуске, доступен для дежурства)")
	fmt.Println("3. Просмотреть историю дежурств")
	fmt.Println("4. Добавить сотрудников")
	fmt.Println("5. Выход")
	fmt.Printf("\n--------------------------------------------------------------------------------------\n")
}

func main() {
	employees, err := pkg.LoadEmployees(employeesFilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	historyStorage, err := pkg.LoadDutyHistory(historyFilePath)
	if err != nil {
		fmt.Println("При попытке загрузить историю дежурств произошла ошибка. ", err)
	}

	if len(*employees) == 0 {
		fmt.Println("Список сотрудников пуст. Сначала добавьте сотрудников.")
		choiceSwitcher(4, employees, historyStorage)
	}

	for {
		displayMenu()

		var choice int
		fmt.Scan(&choice)
		choiceSwitcher(choice, employees, historyStorage)
	}
}

func choiceSwitcher(choice int, employees *[]pkg.Employee, historyStorage *pkg.DutyHistoryStorage) {
	switch choice {
	case 1:
		scheduleStr, schedule, err := pkg.GetSchedule(employees)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(scheduleStr)

		// Сброс счетчиков и сохранение текущего состояния в историческое хранилище (если прошло более 90 дней с момента последнего сброса)
		reseted := pkg.ResetDutyCounters(employees, historyStorage)
		if reseted {
			fmt.Println("Счетчики дежурств сброшены.")
		}

		err = pkg.SaveEmployees(employeesFilePath, employees)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		pkg.AddScheduleToHistory(schedule, historyStorage)
		err = pkg.SaveDutyHistory(historyFilePath, historyStorage)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case 2:
		// Здесь можно запросить имя сотрудника и новый статус, затем обновить его данные.

		employeesStr := pkg.AllEmployees(employees)
		fmt.Println(employeesStr)
		fmt.Println("Введите ID сотрудника, статус которого хотите изменить:")
		var employeeId int
		fmt.Scan(&employeeId)
		fmt.Println("Введите новый статус сотрудника (available, sick, vacation, fired):")
		var newStatus string
		fmt.Scan(&newStatus)
		err := pkg.UpdateEmployeeStatus(employees, employeeId, newStatus)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = pkg.SaveEmployees(employeesFilePath, employees)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Статус сотрудника успешно обновлен.")
	case 3:
		// Просмотреть историю дежурств
		// вывести когда последний раз обнуляли счетчики
		fmt.Printf("Последний раз счетчики дежурств обнулялись %s\n\n", historyStorage.LastResetDate)

		for _, record := range historyStorage.History {
			fmt.Printf("История за период с %s до %s\n", record.Date, record.Date.AddDate(0, 0, 4))
			for _, employee := range record.Employees {
				fmt.Printf("%s | Last support: %s | Last release: %s (Support: %d, Express: %d, Instances: %d)\n", employee.Name, employee.SupportLastDuty, employee.ReleaseLastDuty, employee.SupportDutyCount, employee.ExpressDutyCount, employee.InstancesDutyCount)
			}
		}
		fmt.Println()
		fmt.Println()
		fmt.Println()
	case 4:
		// Добавление новых сотрудников
		employeesStr := pkg.AllEmployees(employees)
		fmt.Printf("%s\n", employeesStr)

		fmt.Println("Введите имя нового сотрудника:")
		var name, surname string
		n, err := fmt.Scan(&name, &surname)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if n > 1 {
			// склеить в имя и фамилию
			name = name + " " + surname
		}
		pkg.AddNewEmployee(employees, name)
		err = pkg.SaveEmployees(employeesFilePath, employees)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Сотрудник успешно добавлен.")
	case 5:
		fmt.Println("Выход из программы...")
		os.Exit(0)
	default:
		fmt.Println("Неизвестный выбор. Пожалуйста, попробуйте снова.")
	}
}
