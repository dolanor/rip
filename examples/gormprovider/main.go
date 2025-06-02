package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/dolanor/rip"
	"github.com/dolanor/rip/encoding/html"
	"github.com/dolanor/rip/encoding/json"
	"github.com/dolanor/rip/encoding/yaml"
	"github.com/dolanor/rip/providers/gormprovider"
)

// start custom type OMIT
type Album struct {
	ID          string
	Name        string
	Artist      string
	ReleaseDate time.Time
}

// end custom type OMIT

func main() {
	// start full OMIT

	// start db init OMIT
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Album{})
	// end db init OMIT

	// start log init OMIT
	logger := slog.Default()
	// end log init OMIT

	// start gormprovider init OMIT
	ap := gormprovider.New[Album](db, logger)
	// end gormprovider init OMIT

	// start rip init OMIT
	codecOpt := rip.WithCodecs(
		json.Codec,
		yaml.Codec,
		html.NewEntityCodec("/albums/"),
		html.NewEntityFormCodec("/albums/"),
	)
	// end rip init OMIT

	// start http init OMIT
	http.HandleFunc(rip.HandleEntities("/albums/", ap, codecOpt)) //HLinterestingCall

	logger.Info("listening on http://localhost:55555/albums") // HLinterestingCall
	http.ListenAndServe(":55555", nil)
	// end http init OMIT

	// end full OMIT
}
