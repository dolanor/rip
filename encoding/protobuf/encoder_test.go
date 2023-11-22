package protobuf

import (
	"bytes"
	"os"
	"testing"

	testdata "github.com/dolanor/rip/encoding/protobuf/testdata"
)

func TestEncoder_Encode(t *testing.T) {
	w := testdata.User{
		Name: "Tanguy",
		Id:   1,
	}

	var b bytes.Buffer
	e := newEncoder(&b)

	err := e.Encode(&w)
	if err != nil {
		t.Fatal(err)
	}

	exp, err := os.ReadFile("testdata/user.pb")
	if err != nil {
		t.Fatal(err)
	}

	if string(b.Bytes()) != string(exp) {
		t.Fatal(err)
	}
}
