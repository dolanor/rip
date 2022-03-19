package rip_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dolanor/rip"
)

func TestHandleResourceWithPath(t *testing.T) {
	up := NewUserProvider()
	http.HandleFunc(rip.HandleResourceWithPath[*User, *UserProvider]("/users/", up))

	s := httptest.NewServer(http.DefaultServeMux)

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

	t.Run("get", func(t *testing.T) {
		resp, err := c.Get(s.URL + "/users/" + u.IDString())
		panicErr(t, err)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatal("get status code is not 200")
		}

		var uGet User
		err = json.NewDecoder(resp.Body).Decode(&uGet)
		panicErr(t, err)
		if uGet != u {
			t.Fatal("user created != from original")
		}
	})

	uUpdated := u
	t.Run("update", func(t *testing.T) {
		uUpdated.BirthDate = uUpdated.BirthDate.Add(2 * time.Hour)
		b, err = json.Marshal(uUpdated)
		panicErr(t, err)

		req, err := http.NewRequest(http.MethodPut, s.URL+"/users/"+u.IDString(), bytes.NewReader(b))
		panicErr(t, err)

		resp, err := c.Do(req)
		panicErr(t, err)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatal("updated status code is not 200")
		}

		err = json.NewDecoder(resp.Body).Decode(&uUpdated)
		panicErr(t, err)
		if uUpdated.BirthDate.Equal(u.BirthDate) {
			t.Fatal("updated birthdate not different")
		}
	})

	// Get after update
	t.Run("get after update", func(t *testing.T) {
		resp, err := c.Get(s.URL + "/users/" + u.IDString())
		panicErr(t, err)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatal("get status code is not 200")
		}

		var uGet User
		err = json.NewDecoder(resp.Body).Decode(&uGet)
		panicErr(t, err)
		if uGet != uUpdated {
			t.Fatal("user updated != from original")
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
		resp, err := c.Get(s.URL + "/users/" + u.IDString())
		panicErr(t, err)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatal("get status code after delete is not 404")
		}
	})

	t.Run("delete again (check idempotency)", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, s.URL+"/users/"+u.IDString(), nil)
		panicErr(t, err)

		resp, err := c.Do(req)
		panicErr(t, err)
		defer resp.Body.Close()
		t.Log(resp.StatusCode)
		if resp.StatusCode != http.StatusNoContent {
			t.Fatal("delete status is not 204")
		}
	})
}

func Greet(ctx context.Context, name string) (string, error) {
	return "Hello " + name, nil
}

type User struct {
	Name      string    `json:"name" xml:"name"`
	BirthDate time.Time `json:"birth_date" xml:"birth_date"`
}

func (u User) IDString() string {
	return u.Name
}

func (u *User) FromString(s string) {
	u.Name = s
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
	log.Printf("SaveUser: %+v", *u)
	up.mem[u.Name] = *u
	return u, nil
}

func (up UserProvider) Get(ctx context.Context, ider rip.IDer) (*User, error) {
	log.Printf("GetUser: %+v", ider.IDString())
	u, ok := up.mem[ider.IDString()]
	if !ok {
		return &User{}, rip.Error{Code: rip.ErrorCodeNotFound, Message: "user not found"}
	}
	return &u, nil
}

func (up *UserProvider) Delete(ctx context.Context, ider rip.IDer) error {
	log.Printf("DeleteUser: %+v", ider.IDString())
	_, ok := up.mem[ider.IDString()]
	if !ok {
		return rip.Error{Code: rip.ErrorCodeNotFound, Message: "user not found"}
	}

	delete(up.mem, ider.IDString())
	return nil
}

func (up *UserProvider) Update(ctx context.Context, u *User) error {
	log.Printf("UpdateUser: %+v", u.IDString())
	_, ok := up.mem[u.Name]
	if !ok {
		return rip.Error{Code: rip.ErrorCodeNotFound, Message: "user not found"}
	}
	up.mem[u.Name] = *u

	return nil
}

func panicErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
