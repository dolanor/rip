package rip

import "testing"

func TestGetEntityField(t *testing.T) {
	trimmedURLPath := "/res/1/name"
	wantEnt := "/res/1"
	wantField := "name"
	gotEnt, gotField := getEntityField(trimmedURLPath)
	if gotEnt != wantEnt {
		t.Fatalf("wrong entity path:\nwant=%s, got=%s", wantEnt, gotEnt)
	}

	if gotField != wantField {
		t.Fatalf("wrong field:\nwant=%s, got=%s", wantField, gotField)
	}
}
