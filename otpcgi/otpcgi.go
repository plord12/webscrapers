package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cgi"
	"os"
	"strings"
)

// takes the output of sms forwaded and writes to a file to be picked up by webscraper

func CGIHandler(rw http.ResponseWriter, req *http.Request) {

	type Otp struct {
		From string
		Text string
	}
	var otp Otp

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)

	body, err := io.ReadAll(req.Body)
	if err != nil {
		panic(fmt.Sprintf("ReadAll failed %v\n", err))
	}
	err = json.Unmarshal(body, &otp)
	if err != nil {
		panic(fmt.Sprintf("Unmarshal failed %v\n", err))
	}
	f, err := os.Create("/home/plord/src/webscrapers/otp/" + strings.ToLower(otp.From))
	if err != nil {
		panic(fmt.Sprintf("File write %v\n", err))
	}
	defer f.Close()
	f.WriteString(otp.Text + "\n")
	f.Chmod(0666)

}

func main() {
	err := cgi.Serve(http.HandlerFunc(CGIHandler))
	if err != nil {
		panic(fmt.Sprintf("could not server cgi: %v", err))
	}
}
