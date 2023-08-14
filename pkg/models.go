package pkg

import "time"

const (
	StatusAvailable = "available" // доступен для дежурства
	StatusSick      = "sick"      // болеет
	StatusVacation  = "vacation"  // в отпуске
	StatusFired     = "fired"     // уволен
)

type Employee struct {
	Id                 int       `json:"id"`
	Name               string    `json:"name"`
	SupportLastDuty    time.Time `json:"support_last_duty"`
	ReleaseLastDuty    time.Time `json:"release_last_duty"`
	Status             string    `json:"status"` // может принимать значения StatusAvailable, StatusSick, StatusVacation, StatusFired
	SupportDutyCount   int       `json:"support_duty_count"`
	ExpressDutyCount   int       `json:"express_duty_count"`
	InstancesDutyCount int       `json:"instances_duty_count"`
}

type DutyHistory struct {
	Date      time.Time `json:"date"`
	Employees []Employee
}

type DutyHistoryStorage struct {
	History       []DutyHistory
	LastResetDate time.Time
}
