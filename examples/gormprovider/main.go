package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/dolanor/rip"
	"github.com/dolanor/rip/encoding/html"
	"github.com/dolanor/rip/providers/gormprovider"
)

type Album struct {
	ID          string `rip:"id"`
	Name        string
	BandName    string
	ReleaseDate time.Time
}

func main() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Album{})

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	ap := gormprovider.New[Album](db, logger)

	ro := rip.NewRouteOptions().
		WithCodecs(
			html.NewEntityCodec("/albums/"),
			html.NewEntityFormCodec("/albums/"),
		)

	http.HandleFunc(rip.HandleEntities("/albums/", ap, ro))

	logger.Info("listening on http://localhost:55555/albums")
	http.ListenAndServe(":55555", nil)
}
