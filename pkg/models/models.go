package models

import (
	"database/sql"
	"time"
)

const (
	StatusAvailable = "available" // доступен для дежурства
	StatusSick      = "sick"      // болеет
	StatusVacation  = "vacation"  // в отпуске
	StatusFired     = "fired"     // уволен
)

type Employee struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
	Status   string `json:"status"` // может принимать значения StatusAvailable, StatusSick, StatusVacation, StatusFired
}

type DutyHistory struct {
	ID       int       `json:"id"`
	UserID   int       `json:"user_id"`
	Name     string    `json:"name"`
	Nickname string    `json:"nickname"`
	DutyDate time.Time `json:"duty_date"`
	DutyType string    `json:"duty_type"` // может принимать значения express_release, instances_release, support
}

type DutySummary struct {
	UserID        int          `json:"user_id"`
	Name          string       `json:"name"`
	Nickname      string       `json:"nickname"`
	Status        string       `json:"status"`
	DutyType      string       `json:"duty_type"`
	DutyTypeCount int          `json:"duty_type_count"`
	LastDutyDate  sql.NullTime `json:"last_duty_date"`
}

type Command struct {
	Type     string
	Statuses map[string]string
	FullName string
	Nickname string
}
