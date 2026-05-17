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
		if err := rows.Scan(&a.Id, &a.Code, &a.TypeName, &a.IsPaid, &a.Status); err != nil {
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

func (s *Store) GetAbsenceTypeById(absence_id int64) (*types.AbsenceTypes, error) {
	var a types.AbsenceTypes
	log.Printf("Querying absence type with ID: %d", absence_id)

	query := `SELECT id, code, type_name, is_paid, status FROM absence_types WHERE id = $1`
	err := s.db.QueryRow(query, absence_id).Scan(&a.Id, &a.Code, &a.TypeName, &a.IsPaid, &a.Status)
	log.Printf("Querying absence type with ID: %d, error: %s", absence_id, err)

	if err != nil {
		return nil, err
	}

	return &a, nil
}

func (s *Store) PatchAbsenceTypeById(
	id int64,
	payload types.AbsenceTypePatchRequest,
) (*types.AbsenceTypes, error) {

	var updated types.AbsenceTypes
	query := `
        UPDATE absence_types
        SET
    `

	if payload.TypeName != nil {
		log.Printf("Updating type_name of absence type with ID: %d to new value: %s", id, *payload.TypeName)
		query += `type_name = $1`
		updated.TypeName = *payload.TypeName
	}

	if payload.IsPaid != nil {
		log.Printf("Updating is_paid of absence type with ID: %d to new value: %t", id, *payload.IsPaid)
		query += `, is_paid = $2`
		updated.IsPaid = *payload.IsPaid
	}

	if payload.Status != nil {
		log.Printf("Updating status of absence type with ID: %d to new value: %s", id, *payload.Status)
		query += `, status = $3`
		updated.Status = *payload.Status
	}

	query += `
		WHERE id = $4
        RETURNING id, type_name, code, is_paid, status;`

	err := s.db.QueryRow(
		query,
		payload.TypeName,
		payload.IsPaid,
		payload.Status,
		id,
	).Scan(
		&updated.Id,
		&updated.TypeName,
		&updated.Code,
		&updated.IsPaid,
		&updated.Status,
	)

	if err != nil {
		return nil, err
	}

	log.Printf("Updated absence type with ID: %d, new values: %+v", id, updated)
	return &updated, nil
}

func (s *Store) ChangeStatusOfAbstenceType(id int64, status string) (absence_type *types.AbsenceTypes, error error) {
	log.Printf("Changing status of absence type with ID: %d to new status: %s", id, status)
	query := `
        UPDATE absence_types
        SET
            status    = $1
        WHERE id = $2
        RETURNING id, code, type_name, is_paid, status;
    `

	var updated types.AbsenceTypes

	err := s.db.QueryRow(
		query,
		status,
		id,
	).Scan(
		&updated.Id,
		&updated.Code,
		&updated.TypeName,
		&updated.IsPaid,
		&updated.Status,
	)

	if err != nil {
		return nil, err
	}
	log.Printf("Updated status of absence type with ID: %d, new values: %+v", id, updated)
	return &updated, nil
}
