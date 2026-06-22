package audit

import (
	"context"
	"time"

	types "main/types/audit"

	"go.mongodb.org/mongo-driver/mongo"
)

type Store struct {
	collection *mongo.Collection
}

func NewStore(db *mongo.Database) *Store {
	return &Store{collection: db.Collection("audit_logs")}
}

func (s *Store) Log(entry types.AuditEntry) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	entry.Timestamp = time.Now()
	_, err := s.collection.InsertOne(ctx, entry)
	return err
}

func (s *Store) GetByEntity(entity string, entityID uint) ([]types.AuditEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := s.collection.Find(ctx, map[string]any{
		"entity":    entity,
		"entity_id": entityID,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	entries := make([]types.AuditEntry, 0)
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}
