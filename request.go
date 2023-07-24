package rip

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dolanor/rip/encoding"
)

func preprocessRequest(reqMethod, handlerMethod string, header http.Header, urlPath string) (cleanedPath string, accept, contentType string, err error) {
	accept, err = bestHeaderValue(header, "Accept", encoding.AvailableEncodings)
	if err != nil {
		return "", "", "", Error{Status: http.StatusUnsupportedMediaType, Message: fmt.Sprintf("bad accept header format: %v", err)}
	}
	if reqMethod != handlerMethod {
		return "", "", "", Error{Status: http.StatusMethodNotAllowed, Message: "bad method"}
	}

	contentType, err = bestHeaderValue(header, "Content-Type", encoding.AvailableEncodings)
	if err != nil {
		return "", "", "", Error{Status: http.StatusUnsupportedMediaType, Message: fmt.Sprintf("bad content type header format: %v", err)}
	}

	// TODO check for the suffix, if .xml, .json, .html, etc
	// if it exists, it overwrites the "Content-Type" because it means the end-user used the URL bar to choose the format.
	splits := strings.Split(urlPath, ".")
	var hasExt bool
	ext := splits[len(splits)-1]
	switch ext {
	case "xml":
		accept = "text/xml"
		hasExt = true
	case "html":
		accept = "text/html"
		hasExt = true
	case "json":
		accept = "application/json"
		hasExt = true
	case "yaml":
		accept = "application/yaml"
		hasExt = true
	}

	cleanedPath = urlPath
	if hasExt {
		cleanedPath = strings.Join(splits[:len(splits)-1], ".")
	}

	// TODO add test for accept ""
	return cleanedPath, accept, contentType, nil
}
