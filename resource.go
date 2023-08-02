package rip

import "context"

// ResourceCreater creates a resource that can be identified.
type ResourceCreater[Rsc IdentifiableResource] interface {
	Create(ctx context.Context, res Rsc) (Rsc, error)
}

// ResourceGetter gets a resource with its id.
type ResourceGetter[Rsc IdentifiableResource] interface {
	Get(ctx context.Context, id IdentifiableResource) (Rsc, error)
}

// ResourceUpdater updates an identifiable resource.
type ResourceUpdater[Rsc IdentifiableResource] interface {
	Update(ctx context.Context, res Rsc) error
}

// ResourceDeleter deletes a resource with its id.
type ResourceDeleter[Rsc IdentifiableResource] interface {
	Delete(ctx context.Context, id IdentifiableResource) error
}

// ResourceLister lists a group of resources.
type ResourceLister[Rsc any] interface {
	ListAll(ctx context.Context) ([]Rsc, error)
}

// ResourceProvider provides identifiable resources.
type ResourceProvider[Rsc IdentifiableResource] interface {
	ResourceCreater[Rsc]
	ResourceGetter[Rsc]
	ResourceUpdater[Rsc]
	ResourceDeleter[Rsc]
	ResourceLister[Rsc]
}

// IdentifiableResource is a resource that can be identified by an string.
type IdentifiableResource interface {
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
