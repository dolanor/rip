package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/samonzeweb/godb"
	"github.com/samonzeweb/godb/adapters/sqlite"
	"github.com/samonzeweb/godb/tablenamer"

	"github.com/dolanor/rip"
	"github.com/dolanor/rip/encoding/html"
	"github.com/dolanor/rip/providers/godbprovider"
)

type Album struct {
	ID          string    `rip:"id" db:"id,key"`
	Name        string    `db:"name"`
	BandName    string    `db:"band_name"`
	ReleaseDate time.Time `db:"release_date"`
}

func main() {
	db, err := godb.Open(sqlite.Adapter, ":memory:")
	if err != nil {
		panic(err)
	}
	db.SetDefaultTableNamer(tablenamer.SnakePlural())

	_, err = db.CurrentDB().Exec(`
	CREATE TABLE IF NOT EXISTS albums(
		id text primary key,
		name text,
		band_name text,
		release_date date
	);
	`)
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	ap := godbprovider.New[Album](db, logger)
	ro := rip.NewRouteOptions().
		WithCodecs(
			html.NewEntityCodec("/albums/"),
			html.NewEntityFormCodec("/albums/"),
		)

	http.HandleFunc(rip.HandleEntities("/albums/", ap, ro))

	logger.Info("listening on http://localhost:55555/albums")
	http.ListenAndServe(":55555", nil)
}
