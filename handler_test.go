package rip_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dolanor/rip"
)

func TestHandleResourceWithPath(t *testing.T) {
	up := NewUserProvider()

	mux := http.NewServeMux()
	mux.HandleFunc(rip.HandleResource[*User]("/users/", up))
	s := httptest.NewServer(mux)

	u := User{Name: "Jane", BirthDate: time.Date(2009, time.November, 1, 23, 0, 0, 0, time.UTC)}

	c := s.Client()
	for name, codec := range rip.AvailableCodecs {
		var b bytes.Buffer
		err := codec.NewEncoder(&b).Encode(u)
		panicErr(t, err)
		t.Run(name, func(t *testing.T) {
			t.Run("create", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodPost, s.URL+"/users/", &b)
				panicErr(t, err)
				req.Header["Content-Type"] = []string{name}
				req.Header["Accept"] = []string{name}

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusCreated {
					b, err := ioutil.ReadAll(resp.Body)
					t.Fatalf("post status code is not 201: body: %v: %s", err, string(b))
				}

				var uCreated User
				err = codec.NewDecoder(resp.Body).Decode(&uCreated)
				panicErr(t, err)
				if uCreated != u {
					t.Fatal("user created != from original")
				}
			})

			t.Run("get", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodGet, s.URL+"/users/"+u.IDString(), nil)
				panicErr(t, err)
				req.Header["Accept"] = []string{name}

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					t.Fatalf("get status code is not 200: body: %v: %s", err, string(b.String()))
				}

				var uGet User
				err = codec.NewDecoder(resp.Body).Decode(&uGet)
				panicErr(t, err)
				if uGet != u {
					t.Fatal("user created != from original")
				}
			})

			uUpdated := u
			t.Run("update", func(t *testing.T) {
				uUpdated.BirthDate = uUpdated.BirthDate.Add(2 * time.Hour)
				err := codec.NewEncoder(&b).Encode(uUpdated)
				panicErr(t, err)

				req, err := http.NewRequest(http.MethodPut, s.URL+"/users/"+u.IDString(), &b)
				panicErr(t, err)
				req.Header["Content-Type"] = []string{name}
				req.Header["Accept"] = []string{name}

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Fatal("updated status code is not 200")
				}

				err = codec.NewDecoder(resp.Body).Decode(&uUpdated)
				panicErr(t, err)
				if uUpdated.BirthDate.Equal(u.BirthDate) {
					t.Fatal("updated birthdate not different")
				}
			})

			// Get after update
			t.Run("get after update", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodGet, s.URL+"/users/"+u.IDString(), nil)
				panicErr(t, err)
				req.Header["Accept"] = []string{name}

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Fatal("get status code is not 200")
				}

				var uGet User
				err = codec.NewDecoder(resp.Body).Decode(&uGet)
				panicErr(t, err)
				if uGet != uUpdated {
					t.Fatal("user updated != from original")
				}
			})

			t.Run("create other user", func(t *testing.T) {
				u := User{Name: "Joe", BirthDate: time.Date(2008, time.November, 1, 23, 0, 0, 0, time.UTC)}
				err := codec.NewEncoder(&b).Encode(u)
				panicErr(t, err)

				req, err := http.NewRequest(http.MethodPost, s.URL+"/users/", &b)
				panicErr(t, err)
				req.Header["Content-Type"] = []string{name}
				req.Header["Accept"] = []string{name}

				respCreate, err := c.Do(req)
				panicErr(t, err)
				defer respCreate.Body.Close()
				if respCreate.StatusCode != http.StatusCreated {
					t.Fatal("post status code is not 201")
				}

				var uCreated User
				err = codec.NewDecoder(respCreate.Body).Decode(&uCreated)
				panicErr(t, err)
				if uCreated != u {
					t.Fatal("user created != from original")
				}
			})

			t.Run("list", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodGet, s.URL+"/users/", nil)
				panicErr(t, err)
				req.Header["Accept"] = []string{name}

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Fatal("get status code is not 200")
				}

				var users []User
				dec := codec.NewDecoder(resp.Body)
				switch name {
				case "text/xml":
					// the curren XML impl doesn't create an array of users
					// it just streams more values.
					// FIXME make XML create a top array value
					for {
						var user User
						err = dec.Decode(&user)
						if err == io.EOF {
							break
						}
						panicErr(t, err)
						users = append(users, user)
					}
				case "text/json":
					err = dec.Decode(&users)
					panicErr(t, err)
				}
				if len(users) != 2 {
					t.Fatal("list does not contain 2 elements")
				}
			})

			t.Run("delete", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodDelete, s.URL+"/users/"+u.IDString(), nil)
				panicErr(t, err)

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusNoContent {
					t.Fatal("delete status is not 204")
				}
			})

			t.Run("get after delete", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodGet, s.URL+"/users/"+u.IDString(), nil)
				panicErr(t, err)
				req.Header["Accept"] = []string{name}

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusNotFound {
					t.Fatal("get status code after delete is not 404")
				}
			})

			t.Run("list after delete", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodGet, s.URL+"/users/", nil)
				panicErr(t, err)
				req.Header["Accept"] = []string{name}

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Fatal("get status code is not 200")
				}

				var users []User
				dec := codec.NewDecoder(resp.Body)
				switch name {
				case "text/xml":
					// the current XML impl doesn't create an array of users
					// it just streams more values.
					// FIXME make XML create a top array value
					for {
						var user User
						err = dec.Decode(&user)
						if err == io.EOF {
							break
						}
						panicErr(t, err)
						users = append(users, user)
					}
				case "text/json":
					err = dec.Decode(&users)
					panicErr(t, err)
				}
				if len(users) != 1 {
					t.Fatal("list does not contain 1 element contains:", len(users))
				}
			})

			t.Run("delete again (check idempotency)", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodDelete, s.URL+"/users/"+u.IDString(), nil)
				panicErr(t, err)
				req.Header["Accept"] = []string{name}

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusNoContent {
					t.Fatal("delete status is not 204")
				}
			})
		})
	}
}

func TestMiddleware(t *testing.T) {
	up := NewUserProvider()
	var callNum int
	middleware := func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			callNum++
			f(w, r)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc(rip.HandleResource[*User]("/users/", up, middleware))
	s := httptest.NewServer(mux)

	u := User{Name: "Jane", BirthDate: time.Date(2009, time.November, 1, 23, 0, 0, 0, time.UTC)}
	b, err := json.Marshal(u)
	panicErr(t, err)

	c := s.Client()
	t.Run("create", func(t *testing.T) {
		respCreate, err := c.Post(s.URL+"/users/", "text/json", bytes.NewReader(b))
		panicErr(t, err)
		defer respCreate.Body.Close()
		if respCreate.StatusCode != http.StatusCreated {
			t.Fatal("post status code is not 201")
		}

		var uCreated User
		err = json.NewDecoder(respCreate.Body).Decode(&uCreated)
		panicErr(t, err)
		if uCreated != u {
			t.Fatal("user created != from original")
		}
	})
	if callNum != 1 {
		t.Fatalf("middleware registered %d calls", callNum)
	}
}

type User struct {
	Name      string    `json:"name" xml:"name"`
	BirthDate time.Time `json:"birth_date" xml:"birth_date"`
}

func (u User) IDString() string {
	return u.Name
}

func (u *User) IDFromString(s string) error {
	u.Name = s

	return nil
}

type UserProvider struct {
	mem map[string]User
}

func NewUserProvider() *UserProvider {
	return &UserProvider{
		mem: map[string]User{},
	}
}

func (up *UserProvider) Create(ctx context.Context, u *User) (*User, error) {
	up.mem[u.Name] = *u
	return u, nil
}

func (up UserProvider) Get(ctx context.Context, ider rip.IdentifiableResource) (*User, error) {
	u, ok := up.mem[ider.IDString()]
	if !ok {
		return &User{}, rip.Error{Code: rip.ErrorCodeNotFound, Message: "user not found"}
	}
	return &u, nil
}

func (up *UserProvider) Delete(ctx context.Context, ider rip.IdentifiableResource) error {
	_, ok := up.mem[ider.IDString()]
	if !ok {
		return rip.Error{Code: rip.ErrorCodeNotFound, Message: "user not found"}
	}

	delete(up.mem, ider.IDString())
	return nil
}

func (up *UserProvider) Update(ctx context.Context, u *User) error {
	_, ok := up.mem[u.Name]
	if !ok {
		return rip.Error{Code: rip.ErrorCodeNotFound, Message: "user not found"}
	}
	up.mem[u.Name] = *u

	return nil
}

func (up *UserProvider) ListAll(ctx context.Context) ([]*User, error) {
	var users []*User
	for _, u := range up.mem {
		// we copy to avoid referring the same pointer that would get updated
		u := u
		users = append(users, &u)
	}

	return users, nil
}

func panicErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
