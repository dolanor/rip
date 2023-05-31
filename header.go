package rip

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type headerQ struct {
	Value string
	Q     float32
}

func bestHeaderValue(header []string, serverPreferences []string) (string, error) {
	clientPreferences, err := headerValues(header)
	if err != nil {
		return "", err
	}

	best, ok := matchHeaderValue(clientPreferences, serverPreferences)
	if !ok {
		// FIXME : use a pkg error
		return "text/html", nil
		//return "", errors.New("no client preferences value found")
	}
	return best, nil
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

func headerValues(header []string) ([]headerQ, error) {
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
						err := fmt.Errorf("parsing q value of %v: %w", aQStrs, err)
						return hqs, Error{Code: ErrorCodeBadQArg, Message: err.Error()}
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
