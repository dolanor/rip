# RIP ⚰

REST In Peace

## Why?

Creating RESTful API in Go is in a way simple and fun in the first time, but also repetitive and error prone the more you create them.  
Copy pasting nearly the same code for each resource you want to GET or POST to except for the request and response types is not that cool, and `interface{}` neither.  
Let's get the best of both worlds with GENERICS 🎆 *everybody screams* 😱  

## How?

The idea would be to use the classic `net/http` package with handlers created from Go types.

```go
http.HandleFunc("/users", rip.HandleResource(User{}))
```

\* *Final code may differ from actual shown footage*
