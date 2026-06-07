// Package event defines the Event domain type — an immutable record of a change.
// Pure domain model — no filesystem I/O (history log writing is done by the store layer).
package event

import "time"

// Type is a machine-readable event type identifier.
type Type string

// Well-known event types.
const (
	ObjectCreated         Type = "object.created"
	ObjectDeleted         Type = "object.deleted"
	ObjectDuplicated      Type = "object.duplicated"
	DocumentUpdated       Type = "document.updated"
	MetadataUpdated       Type = "metadata.updated"
	RelationAdded         Type = "relation.added"
	RelationRemoved       Type = "relation.removed"
	ArtifactAttached      Type = "artifact.attached"
	ArtifactDetached      Type = "artifact.detached"
	LifecycleTransitioned Type = "lifecycle.transitioned"
	CheckpointCreated     Type = "checkpoint.created"
	ReleasePublished      Type = "release.published"
)

// Event is an atomic, append-only record of something that happened to an object.
type Event struct {
	ID       string         `json:"event_id"`
	Time     time.Time      `json:"time"`
	Actor    string         `json:"actor"`
	Type     Type           `json:"type"`
	ObjectID string         `json:"object_id"`
	Payload  map[string]any `json:"payload,omitempty"`
}

// New creates a new Event with the current time.
func New(id string, actor string, typ Type, objectID string, payload map[string]any) Event {
	return Event{
		ID:       id,
		Time:     time.Now().UTC(),
		Actor:    actor,
		Type:     typ,
		ObjectID: objectID,
		Payload:  payload,
	}
}
