package rip

import (
	"fmt"
	"strconv"
	"strings"
)

type HeaderQ struct {
	Value string
	Q     float32
}

func BestHeaderValue(header []string) (string, error) {
	var hqs []HeaderQ
	for _, h := range header {
		for _, aQStrs := range strings.Split(h, ",") {
			aQStrs = strings.TrimSpace(aQStrs)
			aQ := strings.Split(aQStrs, ";")
			if len(aQ) == 1 {
				hq := HeaderQ{Value: aQ[0], Q: 1.0}
				hqs = append(hqs, hq)
			} else {
				qp := strings.Split(aQ[1], "=")
				if len(qp) < 2 {
					continue
				}
				q, err := strconv.ParseFloat(qp[1], 32)
				if err != nil {
					return "", fmt.Errorf("parsing q value: %w", err)
				}
				hq := HeaderQ{Value: aQ[0], Q: float32(q)}
				hqs = append(hqs, hq)
			}
		}
	}

	if len(hqs) == 0 {
		return "", nil
	}
	var last float32
	var best int
	for i := range hqs {
		q := hqs[i].Q
		if q > last {
			last = q
			best = i
		}
	}

	return hqs[best].Value, nil
}
