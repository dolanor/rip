package rip

import (
	"fmt"
	"net/http"
)

func preprocessRequest(reqMethod, handlerMethod string, header http.Header) (accept, contentType string, err error) {
	accept, err = bestHeaderValue(header["Accept"], AvailableEncodings)
	if err != nil {
		return "", "", Error{Status: http.StatusUnsupportedMediaType, Message: fmt.Sprintf("bad accept header format: %v", err)}
	}
	if reqMethod != handlerMethod {
		return "", "", Error{Status: http.StatusMethodNotAllowed, Message: "bad method"}
	}

	contentType, err = bestHeaderValue(header["Content-Type"], AvailableEncodings)
	if err != nil {
		return "", "", Error{Status: http.StatusUnsupportedMediaType, Message: fmt.Sprintf("bad content type header format: %v", err)}
	}

	// TODO check for the suffix, if .xml, .json, .html, etc
	// if it exists, it overwrites the "Content-Type" because it means the end-user used the URL bar to choose the format.

	//TODO add test for accept ""
	return accept, contentType, nil
}
