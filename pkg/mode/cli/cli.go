package cli

import (
	"bufio"
	"database/sql"
	"dev-support-schedule/pkg/handlers"
	"dev-support-schedule/pkg/models"
	"fmt"
	"os"
	"strings"
)

type CLI struct {
	db                     *sql.DB
	googleClientSecretPath string
	googleTokenPath        string
}

func NewCLI(db *sql.DB, googleClientSecretPath, googleTokenPath string) *CLI {
	return &CLI{
		db:                     db,
		googleClientSecretPath: googleClientSecretPath,
		googleTokenPath:        googleTokenPath,
	}
}

func (c *CLI) Start() {
	c.displayMenu()

	for {
		var choice int
		fmt.Scan(&choice)
		//choice = 2
		c.choiceSwitcher(choice)
	}
}

func (c *CLI) displayMenu() {
	fmt.Printf("\n--------------------------------------------------------------------------------------\n")
	fmt.Println("Добро пожаловать в программу расписания дежурств!")
	fmt.Println("Выберите действие:")
	fmt.Println("1. Команда и её статусы")
	fmt.Println("2. Сформировать расписание на следующую неделю")
	fmt.Println("3. Сформировать и сохранить в базе расписание на следующую неделю")
	fmt.Println("4. Обновить статусы сотрудников (на больничном, в отпуске, доступен для дежурства, уволен)")
	fmt.Println("5. Добавить сотрудника")
	fmt.Println("6. Показать это меню")
	fmt.Println("7. Выход")
	fmt.Printf("\n--------------------------------------------------------------------------------------\n")
}

func (c *CLI) choiceSwitcher(choice int) {
	ch := handlers.NewCommandHandler(c.db)

	switch choice {
	case 1:
		fmt.Println(ch.AllEmployees())
	case 2:
		fmt.Println("Введите дату начала недели, на которую надо сформировать расписание, в формате YYYY-MM-DD")
		fmt.Println("Или нажмите Enter, чтобы сформировать расписание на следующую неделю")
		var weekStartDate string
		fmt.Scanln(&weekStartDate)
		fmt.Println(ch.GenerateSchedule(false, weekStartDate))
	case 3:
		fmt.Println("Введите дату начала недели, на которую надо сформировать расписание, в формате YYYY-MM-DD")
		fmt.Println("Или нажмите Enter, чтобы сформировать расписание на следующую неделю")
		var weekStartDate string
		fmt.Scanln(&weekStartDate)
		fmt.Println(ch.GenerateSchedule(true, weekStartDate))
	case 4:
		fmt.Println("Введите статусы сотрудников в формате @nickname status, @nickname2 status, ...")
		fmt.Println("Возможные статусы: available, sick, vacation, fired")
		command, err := readStatuses()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(ch.UpdateEmployeeStatus(command))
	case 5:
		fmt.Println("Введите нового сотрудника в формате @nickname ФИО")
		command, err := readEmployee()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(ch.AddEmployee(command.Nickname, command.FullName))
	case 6:
		c.displayMenu()
	case 7:
		fmt.Println("Выход из программы...")
		os.Exit(0)
	default:
		fmt.Println("Неизвестный выбор. Пожалуйста, попробуйте снова.")
	}
}

func readStatuses() (models.Command, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return models.Command{}, err
	}

	// Удалите символ новой строки в конце
	input = input[:len(input)-1]

	// Разбор строки на пары
	pairs := strings.Split(input, ",")
	result := make(map[string]string)

	for _, pair := range pairs {
		fields := strings.Fields(strings.TrimSpace(pair))
		if len(fields) != 2 {
			return models.Command{}, fmt.Errorf("неверный формат пары: %s", pair)
		}

		nick := strings.TrimPrefix(fields[0], "@")
		status := fields[1]
		result[nick] = status
	}

	// Вывод результата
	return models.Command{Statuses: result}, nil
}

func readEmployee() (models.Command, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return models.Command{}, err
	}

	// Удалите символ новой строки в конце
	input = input[:len(input)-1]

	// Разбор строки на пары
	fields := strings.Fields(strings.TrimSpace(input))
	if len(fields) < 3 {
		return models.Command{}, fmt.Errorf("неверный формат строки: %s", input)
	}

	nick := strings.TrimPrefix(fields[0], "@")
	name := strings.Join(fields[1:], " ")

	// Вывод результата
	return models.Command{Nickname: nick, FullName: name}, nil
}
