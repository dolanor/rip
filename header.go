package rip

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type headerChoice struct {
	Value         string
	QualityFactor float32
}

func contentNegociateBestHeaderValue(header http.Header, headerName string, serverPreferences []string) (string, error) {
	clientPreferences, err := headerChoices(headerName, header[headerName])
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
			clientPreferences = []headerChoice{}
		}
	}

	if len(clientPreferences) == 0 {
		// check in request Content-Type
		headerName := "Content-Type"
		clientPreferences, err = headerChoices(headerName, header[headerName])
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

func matchHeaderValue(clientPreferences []headerChoice, serverPreferences []string) (string, bool) {
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

// headerChoices parse headerName values and sort them by their quality q preferences.
//
// eg. header of text/html; q=0.2, text/xml; q=0.6, application/json; q=0.4
// will return: [ {text/xml, 0.6}, {application/json, 0.4}, {text/html, 0.2} ]
func headerChoices(headerName string, header []string) ([]headerChoice, error) {
	var headerChoices []headerChoice
	for _, h := range header {
		for _, choiceStrs := range strings.Split(h, ",") {
			choiceStrs = strings.TrimSpace(choiceStrs)
			choiceAndQuality := strings.Split(choiceStrs, ";")
			choiceValue := choiceAndQuality[0]

			if len(choiceAndQuality) == 1 {
				choice := headerChoice{Value: choiceValue, QualityFactor: 1.0}
				headerChoices = append(headerChoices, choice)
			} else {
				choiceOptions := choiceAndQuality[1:]

				var hasQuality bool

				for _, opt := range choiceOptions {
					qualityAndPercent := strings.Split(opt, "=")
					if len(qualityAndPercent) < 2 {
						continue
					}

					if strings.TrimSpace(qualityAndPercent[0]) != "q" {
						continue
					}

					hasQuality = true
					percentStr := qualityAndPercent[1]
					quality, err := strconv.ParseFloat(percentStr, 32)
					if err != nil {
						err := fmt.Errorf("parsing q value in header %v of %v: %w", headerName, choiceStrs, err)
						return headerChoices, Error{Code: errorCodeBadQArg, Detail: err.Error(), Source: ErrorSource{Header: headerName}}
					}
					choice := headerChoice{Value: choiceValue, QualityFactor: float32(quality)}
					headerChoices = append(headerChoices, choice)
				}

				if !hasQuality {
					choice := headerChoice{Value: choiceValue, QualityFactor: 1.0}
					headerChoices = append(headerChoices, choice)
				}
			}
		}
	}

	if len(headerChoices) == 0 {
		return headerChoices, nil
	}

	// TODO the: sort during slice creation
	sort.Slice(headerChoices, func(a, b int) bool {
		return headerChoices[a].QualityFactor > headerChoices[b].QualityFactor
	})

	return headerChoices, nil
}
