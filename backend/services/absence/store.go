package absence

import (
	"database/sql"
	"log"
	types "main/types/absence"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetAllAbsenceTypes() (*types.AbsenceResponse, error) {
	var absenceList []types.AbsenceTypes
	log.Print("-- Querying all absence types error")

	query := `SELECT * FROM absence_types;`

	rows, err := s.db.Query(query)
	log.Print("Querying all absence types error %s", err)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var a types.AbsenceTypes
		if err := rows.Scan(&a.AbsenceTypeID, &a.Code, &a.TypeName, &a.IsPaid); err != nil {
			return nil, err
		}
		absenceList = append(absenceList, a)
	}

	log.Print("Absence types retrieved successfully: ", absenceList)
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &types.AbsenceResponse{
		Data: absenceList,
	}, nil

}
