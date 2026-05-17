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
