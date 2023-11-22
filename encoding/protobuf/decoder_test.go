package protobuf

import (
	"os"
	"testing"

	testdata "github.com/dolanor/rip/encoding/protobuf/testdata"
)

func TestDecoder_Decode(t *testing.T) {
	f, err := os.Open("testdata/user.pb")
	if err != nil {
		t.Fatal(err)
	}

	d := newDecoder(f)

	var user testdata.User
	err = d.Decode(&user)
	if err != nil {
		t.Fatal(err)
	}

	if user.Name != "Tanguy" {
		t.Fatal("could not decode name")
	}

	if user.Id != 1 {
		t.Fatal("could not decode ID")
	}
}
