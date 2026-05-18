package types

type AbsenceStore interface {
	GetAllAbsenceTypes() (*AbsenceResponse, error)
	GetAbsenceTypeById(absence_id int64) (*AbsenceTypes, error)
	PatchAbsenceTypeById(absence_id int64, absence AbsenceTypePatchRequest) (*AbsenceTypes, error)
	ChangeStatusOfAbstenceType(absence_id int64, status string) (*AbsenceTypes, error)
}

type AbsenceTypes struct {
	Id       uint   `json:"id"`
	TypeName string `json:"type_name"`
	Code     string `json:"code"`
	IsPaid   bool   `json:"is_paid"`
	Status   string `json:"status"`
}

type AbsenceGetByIdRequest struct {
	Id uint `json:"id" validate:"required"`
}

type AbsenceResponse struct {
	Data []AbsenceTypes `json:"data"`
}

type AbsenceTypePatchRequest struct {
	TypeName *string `json:"type_name,omitempty"`
	Code     *string `json:"code,omitempty"`
	IsPaid   *bool   `json:"is_paid,omitempty"`
	Status   *string `json:"status,omitempty"`
}

type ChangeStatusRequest struct {
	Status string `json:"status"`
}

// REQUEST for absence request creation
type AbsenceTypeCreateRequest struct {
	Id            uint    `json:"id"`
	EmployeeId    *string `json:"employee_id"`
	AbsenceTypeId *string `json:"absence_type_id"`
	StartDate     *string `json:"start_date"`
	EndDate       *string `json:"end_date"`
	TotalDays     *string `json:"total_days"`
	Reason        *string `json:"reason"`
	Status        *string `json:"status"` // 'PENDING','APPROVED', 'REJECTED', 'CANCELLED','DRAFT'
	CreatedAt     *string `json:"created_at"`
	CreatedBy     *string `json:"created_by"`
	ApprovalDate  *string `json:"approval_date"`
	ApprovedBy    *string `json:"approved_by"`
}
