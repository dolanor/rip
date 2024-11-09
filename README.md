# RIP ⚰ [![Go Reference](https://pkg.go.dev/badge/github.com/dolanor/rip.svg)](https://pkg.go.dev/github.com/dolanor/rip) [![Go Report Card](https://goreportcard.com/badge/github.com/dolanor/rip)](https://goreportcard.com/report/github.com/dolanor/rip)

REST in peace

![gopher resting in peace](.meta/assets/rip.png)

## TL;DR

Look at the [examples](examples).

The smallest/simplest is the one with the [GORM provider](examples/gormprovider/main.go).

## Why?

Creating RESTful API in Go is in a way simple and fun in the first time, but also repetitive and error prone the more resources you handle.  
Copy pasting nearly the same code for each resource you want to GET or POST to except for the request and response types is not that cool, and `interface{}` neither.  
Let's get the best of both worlds with **GENERICS** 🎆 *everybody screams* 😱  

## How?

The idea would be to use the classic `net/http` package with handlers created from Go types.

```go
http.HandleFunc(rip.HandleEntities("/users", NewUserProvider(), nil))
```

and it would generate all the necessary boilerplate to have some sane (IMO) HTTP routes.
```go
// HandleEntities associates an urlPath with an entity provider, and handles all HTTP requests in a RESTful way:
//
//	POST   /entities/    : creates the entity
//	GET    /entities/:id : get the entity
//	PUT    /entities/:id : updates the entity (needs to pass the full entity data)
//	DELETE /entities/:id : deletes the entity
//	GET    /entities/    : lists the entities (accepts page and page_size query param)
//
// It also handles fields
//
//	GET    /entities/:id/name : get only the name field of the entity
//	PUT    /entities/:id/name : updates only the name entity field

```


given that `UserProvider` implements the `rip.EntityProvider` interface

```go
type EntityProvider[Ent any] interface {
	Create(ctx context.Context, ent Ent) (Ent, error)
	Get(ctx context.Context, id string) (Ent, error)
	Update(ctx context.Context, ent Ent) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]Ent, error)
}
```

⚠️: Disclaimer, the API is not stable yet, use or contribute at your own risks


\* *Final code may differ from actual shown footage*

## Play with it

```console
go run github.com/dolanor/rip/examples/srv-example@latest
// open your browser to http://localhost:8888/users/ and start editing users
```

## Features

- support for multiple encoding automatically selected with `Accept` and `Content-Type` headers, or entity extension `/entities/1.json`
  - JSON
  - protobuf
  - YAML
  - XML
  - msgpack
  - HTML (read version)
  - HTML forms (write version)
- middlewares
- automatic generation of HTML forms for live editing of entities

### Encoding

You can add your own encoding for your own mime type (I plan on adding some domain type encoding for specific entities, see #13).
It is quite easy to create if your encoding API follows generic standard library encoding packages like `encoding/json`. [Here is how `encoding/json` codec is implemented for RIP](encoding/json/json.go)

## Talks

I gave a [talk at GoLab 2023](https://www.youtube.com/watch?v=_OgqCKrONX8).
I presented it again at [FOSDEM 2024](https://www.youtube.com/watch?v=Z9DOhBCpQi4).

The slides are [in my talks repository](https://github.com/dolanor/talks/blob/main/rip/rip.slide)

(The FOSDEM talk present the more up-to-date API (per-route handler options), demo video (instead of live coding), + a live 3D demo, BUT, I couldn't display my note, so a lot of hesitation and parasite words, sorry about that)


## TODO

- [x] middleware support
- [ ] I'd like to have more composability in the entity provider (some are read-only, some can't list, some are write only…), haven't figured out the right way to design that, yet.
- [ ] it should work for nested entities
- [ ] improve the error API
- [ ] support for hypermedia discoverability
- [x] support for multiple data representation
- [ ] add automatic OpenAPI schema
- [ ] add automatic API client

## Thanks

- logo from Thierry Pfeiffer
