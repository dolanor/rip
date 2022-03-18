package rip

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type HeaderQ struct {
	Value string
	Q     float32
}

func BestHeaderValue(header []string, serverPreferences []string) (string, error) {
	clientPreferences, err := HeaderValues(header)
	if err != nil {
		return "", err
	}

	best := MatchHeaderValue(clientPreferences, serverPreferences)
	return best, nil
}

func MatchHeaderValue(clientPreferences []HeaderQ, serverPreferences []string) string {
	for _, c := range clientPreferences {
		for _, s := range serverPreferences {
			// we found a match
			if c.Value == s {
				return s
			}
		}
	}
	return ""
}

func HeaderValues(header []string) ([]HeaderQ, error) {
	var hqs []HeaderQ
	for _, h := range header {
		for _, aQStrs := range strings.Split(h, ",") {
			aQStrs = strings.TrimSpace(aQStrs)
			aQ := strings.Split(aQStrs, ";")
			if len(aQ) == 1 {
				hq := HeaderQ{Value: aQ[0], Q: 1.0}
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
						return hqs, Error{Code: ErrorCodeBadQArg, Err: err, Message: err.Error()}
					}
					hq := HeaderQ{Value: aQ[0], Q: float32(q)}
					hqs = append(hqs, hq)
				}
				if !hasQ {
					hq := HeaderQ{Value: aQ[0], Q: 1.0}
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
