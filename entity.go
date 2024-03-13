package rip

import "context"

// NewEntityID is needed to know if an entity is a new one not created by the
// EntityProvider yet.
const NewEntityID = "rip-new-entity-id"

// EntityCreator creates a resource that can be identified (an entity).
type EntityCreator[Ent Entity] interface {
	Create(ctx context.Context, ent Ent) (Ent, error)
}

// EntityGetter gets a entity with its id.
type EntityGetter[Ent Entity] interface {
	Get(ctx context.Context, id Entity) (Ent, error)
}

// EntityUpdater updates an entity.
type EntityUpdater[Ent Entity] interface {
	Update(ctx context.Context, ent Ent) error
}

// EntityDeleter deletes a entity with its id.
type EntityDeleter[Ent Entity] interface {
	Delete(ctx context.Context, id Entity) error
}

// EntityLister lists a group of entities.
type EntityLister[Ent any] interface {
	ListAll(ctx context.Context) ([]Ent, error)
}

// start EntityProvider OMIT

// EntityProvider provides identifiable resources.
type EntityProvider[Ent Entity] interface {
	EntityCreator[Ent]
	EntityGetter[Ent]
	EntityUpdater[Ent]
	EntityDeleter[Ent]
	EntityLister[Ent]
}

// end EntityProvider OMIT

// Entity is a resource that can be identified by a string.
// It comes from the concept of entity in Domain Driven Design.
type Entity interface {
	// IDString returns an ID in form of a string.
	IDString() string

	// IDFromString serialize an ID from s.
	IDFromString(s string) error
}

type stringID string

func (i *stringID) IDFromString(s string) error {
	*i = stringID(s)
	return nil
}

func (i stringID) IDString() string {
	return string(i)
}
