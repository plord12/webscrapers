package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cgi"
	"os"
)

// simple logger for smtp2go

func CGIHandler(rw http.ResponseWriter, req *http.Request) {

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)

	body, err := io.ReadAll(req.Body)
	if err != nil {
		panic(fmt.Sprintf("ReadAll failed %v\n", err))
	}

	f, err := os.OpenFile("/var/log/smtp2go/smtp2go.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("File write %v\n", err))
	}
	defer f.Close()
	f.WriteString(string(body) + "\n")
	f.Chmod(0666)

}

func main() {
	err := cgi.Serve(http.HandlerFunc(CGIHandler))
	if err != nil {
		panic(fmt.Sprintf("could not server cgi: %v", err))
	}
}
