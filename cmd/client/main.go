// cmd/client/main.go

package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

func main() {
	url := "http://localhost:8080/static/index.html"

	// Open default browser
	err := openBrowser(url)
	if err != nil {
		fmt.Println("Please open the browser and go to:", url)
	}
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // "linux", "freebsd", etc.
		cmd = "xdg-open"
		args = []string{url}
	}
	return exec.Command(cmd, args...).Start()
}