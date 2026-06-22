package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditStore interface {
	Log(entry AuditEntry) error
	GetByEntity(entity string, entityID uint) ([]AuditEntry, error)
}

type AuditEntry struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"  json:"id"`
	Timestamp time.Time          `bson:"timestamp"       json:"timestamp"`
	Action    string             `bson:"action"          json:"action"`
	Entity    string             `bson:"entity"          json:"entity"`
	EntityID  uint               `bson:"entity_id"       json:"entity_id"`
	ActorID   *uint              `bson:"actor_id"        json:"actor_id"`
	Details   map[string]any     `bson:"details"         json:"details"`
	IP        string             `bson:"ip"              json:"ip"`
	UserAgent string             `bson:"user_agent"      json:"user_agent"`
}
