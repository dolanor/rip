package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

func browse(url string) {
	url = normalizeURL(url)
	openBrowser(url)
}

// openBrowser opens a browser in different OSes
// code from: https://gist.github.com/nanmu42/4fbaf26c771da58095fa7a9f14f23d27
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func normalizeURL(url string) string {
	if strings.HasPrefix(url, ":") {
		url = "localhost" + url
	}

	url = "http://" + url + "/users"

	return url
}
