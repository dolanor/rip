package rip

import "testing"

func TestGetEntityField(t *testing.T) {
	cases := []struct {
		path         string
		entityPath   string
		wantEntityID string
		wantField    string
	}{
		{"/ent/1", "/ent", "1", ""},
		{"/ent/1/", "/ent", "1", ""},
		{"/ent/1/name", "/ent", "1", "name"},
		{"/users/123/addresses/2", "/users/123/addresses", "2", ""},
		{"/users/123/addresses/2/", "/users/123/addresses", "2", ""},
		{"/users/123/addresses/2/city", "/users/123/addresses", "2", "city"},
		{"/ent///1", "/ent", "1", ""},
		{"/ent///1", "/ent//", "1", ""},
		{"/ent/1///", "/ent", "1", ""},
		{"/ent/1////name", "/ent", "1", "name"},
		{"/ent///1////name//", "/ent", "1", "name"},
	}

	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			gotEnt, gotField := getEntityField(c.entityPath, c.path)
			if gotEnt != c.wantEntityID {
				t.Errorf("wrong entity path:\nwant=%s, got=%s", c.wantEntityID, gotEnt)
			}

			if gotField != c.wantField {
				t.Errorf("wrong field:\nwant=%s, got=%s", c.wantField, gotField)
			}
		})
	}
}
