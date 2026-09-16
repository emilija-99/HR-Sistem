package audit

import (
	"context"
	"sort"
	"time"

	types "main/types/audit"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (s *Store) GetRecent(limit int, entity, action string) ([]types.AuditEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	if entity != "" {
		filter["entity"] = entity
	}
	if action != "" {
		filter["action"] = action
	}

	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(int64(limit))
	cursor, err := s.collection.Find(ctx, filter, opts)
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

// DistinctActions returns the sorted set of action values present in the log,
// used to populate filter dropdowns in the UI.
func (s *Store) DistinctActions() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	values, err := s.collection.Distinct(ctx, "action", bson.M{})
	if err != nil {
		return nil, err
	}
	actions := make([]string, 0, len(values))
	for _, v := range values {
		if str, ok := v.(string); ok {
			actions = append(actions, str)
		}
	}
	sort.Strings(actions)
	return actions, nil
}
