package storage

import (
	"database/sql"
	"dev-support-schedule/pkg/models"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatalf("Failed to open database: %v\n", err)
	}

	// Создание таблицы для сотрудников
	createEmployeesTable := `
	CREATE TABLE IF NOT EXISTS employees (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		nickname TEXT NOT NULL,
		status TEXT NOT NULL
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_nickname ON employees(nickname);
	`
	_, err = db.Exec(createEmployeesTable)
	if err != nil {
		log.Fatalf("Failed to create employees table: %v\n", err)
	}

	// Создание таблицы для истории дежурств
	createDutyHistoryTable := `
	CREATE TABLE IF NOT EXISTS duty_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		name TEXT NOT NULL,
		nickname TEXT NOT NULL,
		duty_date DATETIME NOT NULL,
		duty_type TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES employees(id)
	);
	`
	_, err = db.Exec(createDutyHistoryTable)
	if err != nil {
		log.Fatalf("Failed to create duty_history table: %v\n", err)
	}

	// При необходимости добавьте создание других таблиц

	return db
}

func LoadEmployees(db *sql.DB) ([]models.Employee, error) {
	rows, err := db.Query("SELECT id, name, nickname, status FROM employees")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []models.Employee
	for rows.Next() {
		var e models.Employee
		err := rows.Scan(&e.ID, &e.Name, &e.Nickname, &e.Status)
		if err != nil {
			return nil, err
		}
		employees = append(employees, e)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return employees, nil
}

// AddEmployee добавляет нового сотрудника в базу данных
func AddEmployee(db *sql.DB, e models.Employee) error {
	_, err := db.Exec("INSERT INTO employees(name, nickname, status) VALUES (?, ?, ?)", e.Name, e.Nickname, models.StatusAvailable)
	if err != nil {
		return err
	}
	return nil
}

func UpdateEmployees(db *sql.DB, employees []models.Employee) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Подготовка запроса для вставки или обновления записи
	stmt, err := tx.Prepare(`
	UPDATE employees SET status = ? WHERE nickname = ?;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range employees {
		_, err := stmt.Exec(e.Status, e.Nickname)
		if err != nil {
			tx.Rollback() // Откат транзакции в случае ошибки
			return err
		}
	}

	return tx.Commit() // Завершение транзакции
}

// GetDutySummary
func GetDutySummary(db *sql.DB, orderById bool, startDate, endDate time.Time) ([]models.DutySummary, error) {
	var query = `
	SELECT 
		e.id AS user_id, 
		e.name, 
		e.nickname, 
		e.status,
		COALESCE(dh.duty_type, 'no_duty') AS duty_type, 
		COALESCE(COUNT(dh.duty_type), 0) AS duty_type_count,
		MAX(dh.duty_date) AS last_duty_date
	FROM employees e
	LEFT JOIN duty_history dh ON e.id = dh.user_id AND dh.duty_date BETWEEN ? AND ?
	GROUP BY e.id, e.name, e.nickname, dh.duty_type
	`

	if orderById {
		query += " ORDER BY e.id ASC;"
	} else {
		query += " ORDER BY duty_type_count ASC, last_duty_date ASC;"
	}

	startDateStr := startDate.UTC().Format("2006-01-02")
	endDateStr := endDate.UTC().Format("2006-01-02")

	rows, err := db.Query(query, startDateStr, endDateStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.DutySummary

	for rows.Next() {
		var s models.DutySummary
		var lastDutyDateString sql.NullString

		if err := rows.Scan(&s.UserID, &s.Name, &s.Nickname, &s.Status, &s.DutyType, &s.DutyTypeCount, &lastDutyDateString); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		if lastDutyDateString.Valid {
			parsedTime, err := time.Parse("2006-01-02", lastDutyDateString.String)
			if err != nil {
				return nil, err
			}

			s.LastDutyDate = sql.NullTime{
				Time:  parsedTime,
				Valid: true,
			}
		} else {
			s.LastDutyDate = sql.NullTime{
				Valid: false,
			}
		}

		stats = append(stats, s)
	}

	return stats, rows.Err()
}

// AddScheduleToDutyHistory добавляет недельное расписание в историю дежурств
func AddScheduleToDutyHistory(db *sql.DB, schedule []models.DutyHistory) error {
	query := `
		INSERT INTO duty_history (user_id, name, nickname, duty_date, duty_type)
		VALUES (?, ?, ?, ?, ?)	
	`

	for _, e := range schedule {
		// convert Time.Time to string format YYYY-MM-DD
		dateStr := e.DutyDate.UTC().Format("2006-01-02")
		_, err := db.Exec(query, e.UserID, e.Name, e.Nickname, dateStr, e.DutyType)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteDutyHistoryByPeriod удаляет записи из истории дежурств за указанный период
func DeleteDutyHistoryByPeriod(db *sql.DB, startDate, endDate time.Time) error {
	query := `
		DELETE FROM duty_history
		WHERE duty_date BETWEEN ? AND ?
	`
	startDateStr := startDate.UTC().Format("2006-01-02")
	endDateStr := endDate.UTC().Format("2006-01-02")

	_, err := db.Exec(query, startDateStr, endDateStr)
	if err != nil {
		return err
	}

	return nil
}
