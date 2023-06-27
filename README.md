# RIP ⚰

REST in peace

## Why?

Creating RESTful API in Go is in a way simple and fun in the first time, but also repetitive and error prone the more resource you handle.  
Copy pasting nearly the same code for each resource you want to GET or POST to except for the request and response types is not that cool, and `interface{}` neither.  
Let's get the best of both worlds with **GENERICS** 🎆 *everybody screams* 😱  

## How?

The idea would be to use the classic `net/http` package with handlers created from Go types.

```go
http.HandleFunc(rip.HandleResource("/users", NewUserProvider())
```

given that `UserProvider` implements the `rip.ResourceProvider` interface

```go
// simplified version
type ResourceProvider[Rsc IdentifiableResource] interface {
	Create(ctx context.Context, res Rsc) (Rsc, error)
	Get(ctx context.Context, id IdentifiableResource) (Rsc, error)
	Update(ctx context.Context, res Rsc) error
	Delete(ctx context.Context, id IdentifiableResource) error
	ListAll(ctx context.Context) ([]Rsc, error)
}
```

and your resource implements the `IdentifiableResource` interface

```go
type IdentifiableResource interface {
	IDString() string
	IDFromString(s string) error
}
```

⚠️: Disclaimer, the API is not stable yet, use or contribute at your own risks


\* *Final code may differ from actual shown footage*

## TODO

- [x] middleware support
- [ ] I'd like to have more composability in the resource provider (some are read-only, some can't list, some are write only…), haven't figured out the right way to design that, yet.
- [ ] it should work for nested resources
- [ ] improve the error API
- [ ] support for hypermedia discoverability
- [x] support for multiple data representation

