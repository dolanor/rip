package rip

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type headerQ struct {
	Value string
	Q     float32
}

func bestHeaderValue(header http.Header, headerName string, serverPreferences []string) (string, error) {
	clientPreferences, err := headerValues(headerName, header[headerName])
	if err != nil {
		return "", err
	}

	if len(clientPreferences) > 0 &&
		clientPreferences[0].Value == "*/*" {
		// If the first type is a catch all and there is a second, we'll select the second
		if len(clientPreferences) > 1 {
			clientPreferences = clientPreferences[1:]
		} else {
			// Otherwise, we just declare we have no preferences
			clientPreferences = []headerQ{}
		}
	}

	if len(clientPreferences) == 0 {
		// check in request Content-Type
		headerName := "Content-Type"
		clientPreferences, err = headerValues(headerName, header[headerName])
		if err != nil {
			return "", err
		}
	}

	best, ok := matchHeaderValue(clientPreferences, serverPreferences)
	if ok {
		return best, nil
	}

	return "", nil
}

func matchHeaderValue(clientPreferences []headerQ, serverPreferences []string) (string, bool) {
	for _, c := range clientPreferences {
		for _, s := range serverPreferences {
			// we found a match
			if c.Value == s {
				return s, true
			}
		}
	}
	return "", false
}

func headerValues(headerName string, header []string) ([]headerQ, error) {
	var hqs []headerQ
	for _, h := range header {
		for _, aQStrs := range strings.Split(h, ",") {
			aQStrs = strings.TrimSpace(aQStrs)
			aQ := strings.Split(aQStrs, ";")
			if len(aQ) == 1 {
				hq := headerQ{Value: aQ[0], Q: 1.0}
				hqs = append(hqs, hq)
			} else {
				var hasQ bool
				for _, v := range aQ[1:] {
					qp := strings.Split(v, "=")
					if len(qp) < 2 {
						continue
					}
					if strings.TrimSpace(qp[0]) != "q" {
						continue
					}
					hasQ = true
					q, err := strconv.ParseFloat(qp[1], 32)
					if err != nil {
						err := fmt.Errorf("parsing q value in header %v of %v: %w", headerName, aQStrs, err)
						return hqs, Error{Code: errorCodeBadQArg, Detail: err.Error(), Source: ErrorSource{Header: headerName}}
					}
					hq := headerQ{Value: aQ[0], Q: float32(q)}
					hqs = append(hqs, hq)
				}
				if !hasQ {
					hq := headerQ{Value: aQ[0], Q: 1.0}
					hqs = append(hqs, hq)
				}
			}
		}
	}

	if len(hqs) == 0 {
		return hqs, nil
	}

	// TODO the: sort during slice creation
	sort.Slice(hqs, func(a, b int) bool {
		return hqs[a].Q > hqs[b].Q
	})

	return hqs, nil
}
