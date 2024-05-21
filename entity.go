package rip

import "context"

// start EntityProvider OMIT

// EntityProvider provides entities.
// An entity is an identifiable resource. Its id should be marshalable as string.
type EntityProvider[Ent any] interface {

	// Create creates a resource that can be identified (an entity).
	Create(ctx context.Context, ent Ent) (Ent, error)

	// Get gets a entity with its id.
	Get(ctx context.Context, id string) (Ent, error)

	// Update updates an entity.
	Update(ctx context.Context, ent Ent) error

	// Delete deletes a entity with its id.
	Delete(ctx context.Context, id string) error

	// List lists a group of entities.
	List(ctx context.Context, offset, limit int) ([]Ent, error)
}

// end EntityProvider OMIT
