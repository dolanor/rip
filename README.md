# RIP ⚰ [![Go Reference](https://pkg.go.dev/badge/github.com/dolanor/rip.svg)](https://pkg.go.dev/github.com/dolanor/rip) [![Go Report Card](https://goreportcard.com/badge/github.com/dolanor/rip)](https://goreportcard.com/report/github.com/dolanor/rip)

REST in peace

![gopher resting in peace](.meta/assets/rip.png)

## Why?

Creating RESTful API in Go is in a way simple and fun in the first time, but also repetitive and error prone the more resources you handle.  
Copy pasting nearly the same code for each resource you want to GET or POST to except for the request and response types is not that cool, and `interface{}` neither.  
Let's get the best of both worlds with **GENERICS** 🎆 *everybody screams* 😱  

## How?

The idea would be to use the classic `net/http` package with handlers created from Go types.

```go
http.HandleFunc(rip.HandleEntities("/users", NewUserProvider(), nil)
```

and it would generate all the necessary boilerplate to have some sane (IMO) HTTP routes.
```go
// HandleEntities associates an urlPath with an entity provider, and handles all HTTP requests in a RESTful way:
//
//	POST   /entities/    : creates the entity
//	GET    /entities/:id : get the entity
//	PUT    /entities/:id : updates the entity (needs to pass the full entity data)
//	DELETE /entities/:id : deletes the entity
//	GET    /entities/    : lists the entities

```


given that `UserProvider` implements the `rip.EntityProvider` interface

```go
// simplified version
type EntityProvider[Ent Entity] interface {
	Create(ctx context.Context, ent Ent) (Ent, error)
	Get(ctx context.Context, id Entity) (Ent, error)
	Update(ctx context.Context, ent Ent) error
	Delete(ctx context.Context, id Entity) error
	ListAll(ctx context.Context) ([]Ent, error)
}
```

and your resource implements the `Entity` interface

```go
type Entity interface {
	IDString() string
	IDFromString(s string) error
}
```

Right now, it can talk several encoding in reading and writing: JSON, protobuf, XML, YAML, msgpack, HTML and HTML form.
Based on `Accept` and `Content-Type` headers, you can be asymmetrical in encoding: send JSON and read XML.

HTML/HTML Forms allows you to edit your resources directly from your web browser. It's very basic for now.

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

## Talk

I gave a talk at GoLab 2023.
The slides are [in my talks repository](https://github.com/dolanor/talks/blob/main/rip/rip.slide)

## TODO

- [x] middleware support
- [ ] I'd like to have more composability in the entity provider (some are read-only, some can't list, some are write only…), haven't figured out the right way to design that, yet.
- [ ] it should work for nested entities
- [ ] improve the error API
- [ ] support for hypermedia discoverability
- [x] support for multiple data representation

## Thanks

- logo from Thierry Pfeiffer
