package rip

import "context"

// EntityCreator creates a resource that can be identified (an entity).
type EntityCreator[Ent any] interface {
	Create(ctx context.Context, ent Ent) (Ent, error)
}

// EntityGetter gets a entity with its id.
type EntityGetter[Ent any] interface {
	Get(ctx context.Context, id string) (Ent, error)
}

// EntityUpdater updates an entity.
type EntityUpdater[Ent any] interface {
	Update(ctx context.Context, ent Ent) error
}

// EntityDeleter deletes a entity with its id.
type EntityDeleter interface {
	Delete(ctx context.Context, id string) error
}

// EntityLister lists a group of entities.
type EntityLister[Ent any] interface {
	ListAll(ctx context.Context) ([]Ent, error)
}

// start EntityProvider OMIT

// EntityProvider provides entities.
// An entity is an identifiable resource. Its id should be marshalable as string.
type EntityProvider[Ent any] interface {
	EntityCreator[Ent]
	EntityGetter[Ent]
	EntityUpdater[Ent]
	EntityDeleter
	EntityLister[Ent]
}

// end EntityProvider OMIT
