package sqluser

import (
	"time"
)

// start User Entity OMIT
type User struct {
	ID           int       `json:"id" xml:"id"`
	BirthDate    time.Time `json:"birth_date" xml:"birth_date"`
	Name         string    `json:"name" xml:"name"`
	EmailAddress string    `json:"email_address" xml:"email_address"`
}

// end User Entity OMIT
