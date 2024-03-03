package memuser

import (
	"strconv"
	"time"
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
