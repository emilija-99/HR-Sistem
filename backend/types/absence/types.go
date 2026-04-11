package types

type AbsenceStore interface {
	GetAllAbsenceTypes() (*AbsenceResponse, error)
}

type AbsenceTypes struct {
	AbsenceTypeID uint   `json:"absence_type_id"`
	TypeName      string `json:"type_name"`
	Code          string `json:"code"`
	IsPaid        bool   `json:"is_paid"`
}

type AbsenceResponse struct {
	Data []AbsenceTypes `json:"data"`
}
