package rip

import (
	"bytes"
	gjson "encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dolanor/rip/encoding"
	"github.com/dolanor/rip/encoding/json"
	"github.com/dolanor/rip/encoding/msgpack"
	"github.com/dolanor/rip/encoding/xml"
	"github.com/dolanor/rip/encoding/yaml"
)

func TestHandleResourceWithPath(t *testing.T) {
	encoding.RegisterCodec(json.Codec, json.MimeTypes...)
	encoding.RegisterCodec(yaml.Codec, yaml.MimeTypes...)
	encoding.RegisterCodec(msgpack.Codec, msgpack.MimeTypes...)
	encoding.RegisterCodec(xml.Codec, xml.MimeTypes...)

	up := newUserProvider()

	mux := http.NewServeMux()
	mux.HandleFunc(HandleEntities("/users/", up))
	s := httptest.NewServer(mux)

	u := user{Name: "Jane", BirthDate: time.Date(2009, time.November, 1, 23, 0, 0, 0, time.UTC)}

	c := s.Client()

	availableCodecs := encoding.AvailableCodecs().Codecs

	for name, codec := range availableCodecs {
		var b bytes.Buffer
		err := codec.NewEncoder(&b).Encode(u)
		panicErr(t, err)
		t.Run(name, func(t *testing.T) {
			t.Run("create", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodPost, s.URL+"/users/", &b)
				panicErr(t, err)
				req.Header["Content-Type"] = []string{name}
				req.Header["Accept"] = []string{name}

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusCreated {
					b, err := io.ReadAll(resp.Body)
					t.Fatalf("post status code is not 201: body: %v: %s", err, string(b))
				}
				var buf bytes.Buffer
				_, err = buf.ReadFrom(resp.Body)
				panicErr(t, err)

				var uCreated user
				err = codec.NewDecoder(&buf).Decode(&uCreated)

				panicErr(t, err)
				// somehow msgpack package changes the time reference (UTC to CET)
				// so we're more graceful with the checks
				if uCreated != u && !uCreated.BirthDate.Equal(u.BirthDate) {
					t.Fatal("user created != from original:", uCreated, u)
				}
			})

			t.Run("get", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodGet, s.URL+"/users/"+u.IDString(), nil)
				panicErr(t, err)
				req.Header["Accept"] = []string{name}

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					t.Fatalf("get status code is not 200: %d. body: %v: %s", resp.StatusCode, err, string(b.String()))
				}

				var uGet user
				err = codec.NewDecoder(resp.Body).Decode(&uGet)
				panicErr(t, err)
				if uGet != u && !uGet.BirthDate.Equal(u.BirthDate) {
					t.Fatal("user get != from original:", uGet, u)
				}
			})

			uUpdated := u
			t.Run("update", func(t *testing.T) {
				uUpdated.BirthDate = uUpdated.BirthDate.Add(2 * time.Hour)
				err := codec.NewEncoder(&b).Encode(uUpdated)
				panicErr(t, err)

				req, err := http.NewRequest(http.MethodPut, s.URL+"/users/"+u.IDString(), &b)
				panicErr(t, err)
				req.Header["Content-Type"] = []string{name}
				req.Header["Accept"] = []string{name}

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Fatal("updated status code is not 200:", resp.StatusCode)
				}

				err = codec.NewDecoder(resp.Body).Decode(&uUpdated)
				panicErr(t, err)
				if uUpdated.BirthDate.Equal(u.BirthDate) {
					t.Fatal("updated birthdate not different")
				}
			})

			// Get after update
			t.Run("get after update", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodGet, s.URL+"/users/"+u.IDString(), nil)
				panicErr(t, err)
				req.Header["Accept"] = []string{name}

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Fatal("get status code is not 200:", resp.StatusCode)
				}

				var uGet user
				err = codec.NewDecoder(resp.Body).Decode(&uGet)
				panicErr(t, err)
				if uGet != uUpdated && !uGet.BirthDate.Equal(uUpdated.BirthDate) {
					t.Fatal("user get != from original:", uGet, uUpdated)
				}
			})

			t.Run("create other user", func(t *testing.T) {
				u := user{Name: "Joe", BirthDate: time.Date(2008, time.November, 1, 23, 0, 0, 0, time.UTC)}
				err := codec.NewEncoder(&b).Encode(u)
				panicErr(t, err)

				req, err := http.NewRequest(http.MethodPost, s.URL+"/users/", &b)
				panicErr(t, err)
				req.Header["Content-Type"] = []string{name}
				req.Header["Accept"] = []string{name}

				respCreate, err := c.Do(req)
				panicErr(t, err)
				defer respCreate.Body.Close()
				if respCreate.StatusCode != http.StatusCreated {
					t.Fatal("post status code is not 201:", respCreate.StatusCode)
				}

				var uCreated user
				err = codec.NewDecoder(respCreate.Body).Decode(&uCreated)
				panicErr(t, err)
				if uCreated != u && !uCreated.BirthDate.Equal(u.BirthDate) {
					t.Fatal("user created != from original:", uCreated, u)
				}
			})

			t.Run("list", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodGet, s.URL+"/users/", nil)
				panicErr(t, err)
				req.Header["Accept"] = []string{name}

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Fatal("get status code is not 200:", resp.StatusCode)
				}

				var users []user
				dec := codec.NewDecoder(resp.Body)
				switch name {
				case "application/xml", "text/xml":
					// the curren XML impl doesn't create an array of users
					// it just streams more values.
					// FIXME make XML create a top array value
					for {
						var user user
						err = dec.Decode(&user)
						if err == io.EOF {
							break
						}
						panicErr(t, err)
						users = append(users, user)
					}
				default:
					err = dec.Decode(&users)
					panicErr(t, err)

				}
				if len(users) != 2 {
					t.Fatal("list does not contain 2 elements, contains:", len(users))
				}
			})

			t.Run("delete", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodDelete, s.URL+"/users/"+u.IDString(), nil)
				panicErr(t, err)

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Fatal("delete status is not 200:", resp.StatusCode)
				}
			})

			t.Run("get after delete", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodGet, s.URL+"/users/"+u.IDString(), nil)
				panicErr(t, err)
				req.Header["Accept"] = []string{name}

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusNotFound {
					t.Fatal("get status code after delete is not 404:", resp.StatusCode)
				}
			})

			t.Run("list after delete", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodGet, s.URL+"/users/", nil)
				panicErr(t, err)
				req.Header["Accept"] = []string{name}

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Fatal("get status code is not 200:", resp.StatusCode)
				}

				var users []user
				dec := codec.NewDecoder(resp.Body)
				switch name {
				case "text/xml":
					// the current XML impl doesn't create an array of users
					// it just streams more values.
					// FIXME make XML create a top array value
					for {
						var user user
						err = dec.Decode(&user)
						if err == io.EOF {
							break
						}
						panicErr(t, err)
						users = append(users, user)
					}
				default:
					err = dec.Decode(&users)
					panicErr(t, err)
				}
				if len(users) != 1 {
					t.Fatal("list does not contain 1 element contains:", len(users))
				}
			})

			t.Run("delete again (check idempotency)", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodDelete, s.URL+"/users/"+u.IDString(), nil)
				panicErr(t, err)
				req.Header["Accept"] = []string{name}

				resp, err := c.Do(req)
				panicErr(t, err)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Fatal("delete status is not 200:", resp.StatusCode)
				}
			})
		})
	}
}

func TestMiddleware(t *testing.T) {
	up := newUserProvider()
	var callNum int
	middleware := func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			callNum++
			f(w, r)
		}
	}
	encoding.RegisterCodec(json.Codec, json.MimeTypes...)

	mux := http.NewServeMux()
	mux.HandleFunc(HandleEntities[*user]("/users/", up, middleware))
	s := httptest.NewServer(mux)

	u := user{Name: "Jane", BirthDate: time.Date(2009, time.November, 1, 23, 0, 0, 0, time.UTC)}
	b, err := gjson.Marshal(u)
	panicErr(t, err)

	c := s.Client()
	t.Run("create", func(t *testing.T) {
		respCreate, err := c.Post(s.URL+"/users/", "application/json", bytes.NewReader(b))
		panicErr(t, err)
		defer respCreate.Body.Close()

		if respCreate.StatusCode != http.StatusCreated {
			t.Fatal("post status code is not 201:", respCreate.StatusCode)
		}

		var uCreated user
		err = gjson.NewDecoder(respCreate.Body).Decode(&uCreated)
		panicErr(t, err)
		if uCreated != u {
			t.Fatal("user created != from original")
		}
	})
	if callNum != 1 {
		t.Fatalf("middleware registered %d calls", callNum)
	}
}

func panicErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
