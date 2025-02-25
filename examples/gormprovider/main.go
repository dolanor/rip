package main

import (
	"log/slog"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/dolanor/rip"
	"github.com/dolanor/rip/encoding/html"
	"github.com/dolanor/rip/encoding/json"
	"github.com/dolanor/rip/providers/gormprovider"
)

type Album struct {
	ID          string
	Name        string
	Artist      string
	ReleaseDate time.Time
}

func main() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Album{})

	ap := gormprovider.New[Album](db, slog.Default())

	ro := rip.NewRouteOptions().
		WithCodecs(
			json.Codec,
			html.NewEntityCodec("/albums/"),
			html.NewEntityFormCodec("/albums/"),
		)

	http.HandleFunc(rip.HandleEntities("/albums/", ap, ro))

	slog.Info("listening on http://localhost:55555/albums")
	http.ListenAndServe(":55555", nil)
}
