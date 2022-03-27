# RIP ⚰

REST in peace

## Why?

Creating RESTful API in Go is in a way simple and fun in the first time, but also repetitive and error prone the more you create them.  
Copy pasting nearly the same code for each resource you want to GET or POST to except for the request and response types is not that cool, and `interface{}` neither.  
Let's get the best of both worlds with **GENERICS** 🎆 *everybody screams* 😱  

## How?

⚠️: Disclaimer, the API is not stable yet, use or contribute at your own risks

The idea would be to use the classic `net/http` package with handlers created from Go types.

```go
http.HandleFunc(rip.HandleResource[User, UserProvider]("/users", NewUserProvider())
```

given that `UserProvider` implements the `rip.ResourceProvider` interface

```go
// ResourceProvider provides identifiable resources.
type ResourceProvider[Rsc IdentifiableResource] interface {
	Creater[Rsc]
	Getter[Rsc]
	Updater[Rsc]
	Deleter[Rsc]
	Lister[Rsc]
}

// Creater creates a resource that can be identified.
type Creater[Rsc IdentifiableResource] interface {
	Create(ctx context.Context, res Rsc) (Rsc, error)
}

// Getter gets a resource with its id.
type Getter[Rsc IdentifiableResource] interface {
	Get(ctx context.Context, id IdentifiableResource) (Rsc, error)
}

// Updater updates an identifiable resource.
type Updater[Rsc IdentifiableResource] interface {
	Update(ctx context.Context, res Rsc) error
}

// Deleter deletes a resource with its id.
type Deleter[Rsc IdentifiableResource] interface {
	Delete(ctx context.Context, id IdentifiableResource) error
}

// Lister lists a group of resources.
type Lister[Rsc any] interface {
	ListAll(ctx context.Context) ([]Rsc, error)
}


```

\* *Final code may differ from actual shown footage*

## TODO

- [x] middleware support
- [ ] I'd like to have more composability in the resource provider (some are read-only, some can't list, some are write only…), haven't figured out the right way to design that, yet.
- [ ] it should work for nested resources
- [ ] improve the error API
- [ ] support for hypermedia discoverability
- [ ] support for multiple data representation

