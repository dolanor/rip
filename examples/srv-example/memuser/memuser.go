package memuser

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/dolanor/rip"
)

// start User Entity OMIT
type User struct {
	ID           int       `json:"id" xml:"id"`
	BirthDate    time.Time `json:"birth_date" xml:"birth_date"`
	Name         string    `json:"name" xml:"name"`
	EmailAddress string    `json:"email_address" xml:"email_address"`
}

// start User Entity interface OMIT
func (u User) IDString() string {
	return strconv.Itoa(u.ID)
}

func (u *User) IDFromString(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	u.ID = n
	return nil
}

// end User Entity OMIT
// end User Entity interface OMIT

// start User Provider OMIT
type UserProvider struct {
	mem map[int]*User
}

// end User Provider OMIT

func NewUserProvider() *UserProvider {
	u := User{ID: 1, Name: "Jean", EmailAddress: "jean@example.com", BirthDate: time.Date(1900, time.November, 15, 0, 0, 0, 0, time.UTC)}
	return &UserProvider{
		mem: map[int]*User{
			u.ID: &u,
		},
	}
}

func (up *UserProvider) Create(ctx context.Context, u *User) (*User, error) {
	log.Printf("SaveUser: %+v", *u)
	id := rand.Intn(1000)
	u.ID = id

	up.mem[u.ID] = u
	return u, nil
}

func (up UserProvider) Get(ctx context.Context, ent rip.Entity) (*User, error) {
	log.Printf("GetUser: %+v", ent.IDString())

	if ent.IDString() == rip.NewEntityID {
		return &User{}, nil
	}

	id, err := strconv.Atoi(ent.IDString())
	if err != nil {
		return nil, err
	}

	u, ok := up.mem[id]
	if !ok {
		return &User{}, rip.ErrNotFound
	}
	return u, nil
}

func (up *UserProvider) Delete(ctx context.Context, ent rip.Entity) error {
	log.Printf("DeleteUser: %+v", ent.IDString())
	id, err := strconv.Atoi(ent.IDString())
	if err != nil {
		return err
	}

	_, ok := up.mem[id]
	if !ok {
		return rip.ErrNotFound
	}

	delete(up.mem, id)
	return nil
}

// start User Provider Update OMIT
func (up *UserProvider) Update(ctx context.Context, u *User) error {
	log.Printf("UpdateUser: %+v", u.IDString())
	_, ok := up.mem[u.ID]
	if !ok {
		return rip.ErrNotFound
	}
	up.mem[u.ID] = u

	return nil
}

// end User Provider Update OMIT

func (up UserProvider) ListAll(ctx context.Context) ([]*User, error) {
	log.Printf("ListAllUser")
	var users []*User
	for _, u := range up.mem {
		u := u
		users = append(users, u)
	}
	return users, nil
}
