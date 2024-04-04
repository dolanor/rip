package rip

import (
	"fmt"
	"net/http"
)

func preprocessRequest(reqMethod, handlerMethod string, header http.Header, urlPath string, options *RouteOptions) (cleanedPath string, accept, contentType string, err error) {
	accept, err = bestHeaderValue(header, "Accept", options.codecs.OrderedMimeTypes)
	if err != nil {
		return "", "", "", Error{
			Status: http.StatusBadRequest,
			Detail: fmt.Sprintf("bad accept header format: %v, codecs available: %v", header["Accept"], options.codecs.OrderedMimeTypes),
		}
	}

	if accept == "" &&
		(reqMethod == http.MethodPost || reqMethod == http.MethodGet) {
		return "", "", "", Error{
			Status: http.StatusNotAcceptable,
			Detail: fmt.Sprintf("bad accept type: %v, codecs available: %v", header["Accept"], options.codecs.OrderedMimeTypes),
		}
	}

	if reqMethod != handlerMethod {
		return "", "", "", Error{Status: http.StatusMethodNotAllowed, Detail: "bad method"}
	}

	if reqMethod == http.MethodGet {
		// We can ignore content type as there should be no body for a GET
		return urlPath, accept, contentType, nil
	}

	contentType, err = bestHeaderValue(header, "Content-Type", options.codecs.OrderedMimeTypes)
	if err != nil {
		return "", "", "", Error{Status: http.StatusUnsupportedMediaType, Detail: fmt.Sprintf("bad content type header format: %v", err)}
	}

	return urlPath, accept, contentType, nil
}
