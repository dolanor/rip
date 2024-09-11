package encoding

import (
	"net/http"
)

type RequestResponseWriter struct {
	http.ResponseWriter
	Request *http.Request
}
