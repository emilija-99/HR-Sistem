package types

type EmployeeStore interface {
}

type Countries struct {
	CountryID   uint   `json:"country_id"`
	CountryName string `json:"country_name"`
}

type Positions struct {
	ID     uint   `json:"id"`
	DepartmentID uint   `json:"department_id"`
	Title  string `json:"title"`
	Level uint   `json:"level"`
	Description uint   `json:"description"`
	Status  bool   `json:"status"`
}

type Departments struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Description      uint   `json:"description"`
	Status       bool   `json:"status"`
}

type EmployeePosition{
	EmployeePositionID uint `json:"employee_position_id"`
EmployeeID uint `json:"employee_id",
PositionID uint `json:"position_id"`
Sa;ary uint `json:"salary"`
StartDate string `json:"start_date"`
EndDate string `json:"end_date"`

}
