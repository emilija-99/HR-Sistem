package types

import "time"

type AbsenceStore interface {
	// absence types
	GetAllAbsenceTypes() (*AbsenceResponse, error)

	// absence requests
	CreateRequest(req AbsenceRequest) (*AbsenceRequest, error)
	GetRequestByID(id int64) (*AbsenceRequest, error)
	GetRequestsByEmployee(employeeID uint) ([]AbsenceRequest, error)
	GetAllRequests() ([]AbsenceRequest, error)
	UpdateRequest(req AbsenceRequest) (*AbsenceRequest, error)
	UpdateRequestStatus(id int64, status string, approvedBy *uint) (*AbsenceRequest, error)

	// public holidays (business-day calculation)
	GetHolidays(employeeID uint, from, to string) (map[string]bool, error)

	// leave balance
	GetBalanceByEmployee(employeeID uint) ([]LeaveBalanceSummary, error)
	AddLedgerEntry(entry LeaveBalanceEntry) (*LeaveBalanceEntry, error)

	// leave policies
	GetPolicies() ([]LeavePolicy, error)
	GetActivePolicy(employeeID, absenceTypeID uint, atDate string) (*LeavePolicy, error)
	AssignEmployeePolicy(assignment EmployeePolicyAssignment) error
	GetEmployeePolicyAssignments(employeeID uint) ([]EmployeePolicyAssignment, error)
	GetAvailableForType(employeeID, absenceTypeID uint) (float64, error)
	RolloverYear(year int) (*RolloverReport, error)

	// scheduled jobs
	TryMarkScheduledRun(job, period, subject string) (bool, error)
	RunMonthlyAccrual(period string) (*RolloverReport, error)
	RunExpiration() (*RolloverReport, error)
}

// ── Absence types ─────────────────────────────────────────────

type AbsenceTypes struct {
	Id       uint   `json:"id"`
	TypeName string `json:"type_name"`
	Code     string `json:"code"`
	IsPaid   bool   `json:"is_paid"`
	Status   string `json:"status"`
}

type AbsenceResponse struct {
	Data []AbsenceTypes `json:"data"`
}

// ── Absence requests ──────────────────────────────────────────

type AbsenceRequest struct {
	ID            uint       `json:"id"`
	EmployeeID    uint       `json:"employee_id"`
	AbsenceTypeID uint       `json:"absence_type_id"`
	StartDate     string     `json:"start_date"`
	EndDate       string     `json:"end_date"`
	TotalDays     float64    `json:"total_days"`
	Reason        *string    `json:"reason,omitempty"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	CreatedBy     uint       `json:"created_by"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`
	ApprovedBy    *uint      `json:"approved_by,omitempty"`

	// joined fields (read-only)
	TypeName  string `json:"type_name,omitempty"`
	IsPaid    *bool  `json:"is_paid,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

type CreateAbsenceRequestPayload struct {
	AbsenceTypeID uint    `json:"absence_type_id" validate:"required"`
	StartDate     string  `json:"start_date"      validate:"required"`
	EndDate       string  `json:"end_date"        validate:"required"`
	Reason        *string `json:"reason"`
	Status        *string `json:"status"` // optional: omit → PENDING, "DRAFT" → DRAFT
}

// UpdateAbsenceRequestPayload — partial edit of a DRAFT or PENDING request
// owned by the caller.
type UpdateAbsenceRequestPayload struct {
	AbsenceTypeID *uint   `json:"absence_type_id"`
	StartDate     *string `json:"start_date"`
	EndDate       *string `json:"end_date"`
	Reason        *string `json:"reason"`
}

// ── Leave balance ─────────────────────────────────────────────

// LeaveBalanceSummary is the aggregated available days per absence type
type LeaveBalanceSummary struct {
	AbsenceTypeID uint    `json:"absence_type_id"`
	TypeName      string  `json:"type_name"`
	Code          string  `json:"code"`
	IsPaid        bool    `json:"is_paid"`
	GrantedDays   float64 `json:"granted_days"`   // sum of positive entries (ACCRUAL, CARRY_OVER, CANCELLED, MANUAL_ADJUST+)
	UsedDays      float64 `json:"used_days"`      // abs of CONSUMED
	AvailableDays float64 `json:"available_days"` // granted - used
}

// LeaveBalanceEntry is a single ledger row
type LeaveBalanceEntry struct {
	ID            uint      `json:"id"`
	EmployeeID    uint      `json:"employee_id"`
	AbsenceTypeID uint      `json:"absence_type_id"`
	EntryType     string    `json:"entry_type"` // ACCRUAL, CARRY_OVER, CONSUMED, CANCELLED, MANUAL_ADJUST, EXPIRATION
	Days          float64   `json:"days"`
	AccrualYear   int       `json:"accrual_year"`
	ExpiresAt     *string   `json:"expires_at,omitempty"`
	ReferenceID   *uint     `json:"reference_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`

	// joined (read-only)
	TypeName string `json:"type_name,omitempty"`
}

// GrantBalancePayload — admin grants days (annual accrual)
type GrantBalancePayload struct {
	EmployeeID    uint    `json:"employee_id"     validate:"required"`
	AbsenceTypeID uint    `json:"absence_type_id" validate:"required"`
	Days          float64 `json:"days"            validate:"required"`
	AccrualYear   int     `json:"accrual_year"`
	ExpiresAt     *string `json:"expires_at"`
}

// AdjustBalancePayload — admin manual adjustment (+/-)
type AdjustBalancePayload struct {
	EmployeeID    uint    `json:"employee_id"     validate:"required"`
	AbsenceTypeID uint    `json:"absence_type_id" validate:"required"`
	Days          float64 `json:"days"            validate:"required"`
	AccrualYear   int     `json:"accrual_year"`
}

// ── Leave policies ────────────────────────────────────────────

type LeavePolicy struct {
	ID                   uint     `json:"id"`
	Name                 string   `json:"name"`
	GrantPolicy          string   `json:"grant_policy"`
	DaysPerPeriod        *float64 `json:"days_per_period,omitempty"`
	PeriodMonths         *int     `json:"period_months,omitempty"`
	AllowCarryOver       bool     `json:"allow_carry_over"`
	CarryOverMaxDays     *float64 `json:"carry_over_max_days,omitempty"`
	CarryOverExpiryMonth *int     `json:"carry_over_expiry_month,omitempty"`
	CarryOverExpiryDay   *int     `json:"carry_over_expiry_day,omitempty"`
	RequiresBalance      bool     `json:"requires_balance"`
}

type EmployeePolicyAssignment struct {
	ID            uint    `json:"id"`
	EmployeeID    uint    `json:"employee_id"`
	AbsenceTypeID uint    `json:"absence_type_id"`
	PolicyID      uint    `json:"policy_id"`
	ValidFrom     string  `json:"valid_from"`
	ValidTo       *string `json:"valid_to,omitempty"`

	// joined (read-only)
	AbsenceTypeName string `json:"absence_type_name,omitempty"`
	PolicyName      string `json:"policy_name,omitempty"`
}

type AssignEmployeePolicyPayload struct {
	EmployeeID    uint    `json:"employee_id"     validate:"required"`
	AbsenceTypeID uint    `json:"absence_type_id" validate:"required"`
	PolicyID      uint    `json:"policy_id"       validate:"required"`
	ValidFrom     string  `json:"valid_from"      validate:"required"`
	ValidTo       *string `json:"valid_to"`
}

type RolloverRequest struct {
	Year int `json:"year"`
}

type RolloverReport struct {
	Year       int              `json:"year"`
	Accrued    []RolloverAction `json:"accrued"`
	Carried    []RolloverAction `json:"carried"`
	Expired    []RolloverAction `json:"expired"`
	SkippedCnt int              `json:"skipped_count"`
}

type RolloverAction struct {
	EmployeeID    uint    `json:"employee_id"`
	AbsenceTypeID uint    `json:"absence_type_id"`
	TypeName      string  `json:"type_name"`
	Days          float64 `json:"days"`
}
