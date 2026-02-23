package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/juju/fslock"
)

type Email struct {
	Email_id string
	Event    string
	Rcpt     string
}

// simple logger for smtp2go

func CGIHandler(rw http.ResponseWriter, req *http.Request) {

	// ensure we handle these sequentially
	lock := fslock.New("/var/log/smtp2go/smtp2go.log")
	lockErr := lock.Lock()
	if lockErr != nil {
		panic(fmt.Sprintf("File lock failed %v\n", lockErr))
	}
	defer lock.Unlock()

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

	logf, logerr := os.OpenFile("/var/log/smtp2go/smtp2go.error", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logerr != nil {
		panic(fmt.Sprintf("File write %v\n", logerr))
	}
	defer logf.Close()

	var email Email
	err = json.Unmarshal([]byte(string(body)), &email)
	if err == nil {
		if email.Event != "processed" {
			key := email.Email_id + "-" + strings.ReplaceAll(strings.ReplaceAll(email.Rcpt, "@", "-"), ".", "-")

			cmd := exec.Command("/usr/bin/mosquitto_pub", "-r", "-t", "homeassistant/sensor/smtp/"+key+"/config", "-m", "{"+
				"\"name\": \"email-mqtt-"+key+"\","+
				"\"icon\": \"mdi:email\","+
				"\"expire_after\": 2628000,"+
				"\"state_topic\": \"homeassistant/sensor/smtp/"+key+"/state\","+
				"\"json_attributes_topic\": \"homeassistant/sensor/smtp/"+key+"/attributes\"}")
			out, cmderr := cmd.CombinedOutput()
			if cmderr != nil {
				fmt.Fprintf(logf, "Exec %v %s\n", cmderr, out)
			}

			cmd = exec.Command("/usr/bin/mosquitto_pub", "-r", "-t", "homeassistant/sensor/smtp/"+key+"/state", "-m", email.Event)
			out, cmderr = cmd.CombinedOutput()
			if cmderr != nil {
				fmt.Fprintf(logf, "Exec %v %s\n", cmderr, out)
			}

			var re = regexp.MustCompile("^{")
			s := re.ReplaceAllString(string(body), "{\"topic\":\""+"homeassistant/sensor/smtp/"+key+"\",")

			cmd = exec.Command("/usr/bin/mosquitto_pub", "-r", "-t", "homeassistant/sensor/smtp/"+key+"/attributes", "-m", s)
			out, cmderr = cmd.CombinedOutput()
			if cmderr != nil {
				fmt.Fprintf(logf, "Exec %v %s\n", cmderr, out)
			}
		}
	} else {
		panic(fmt.Sprintf("Unmarshal %v\n", err))
	}

}

func main() {
	err := cgi.Serve(http.HandlerFunc(CGIHandler))
	if err != nil {
		panic(fmt.Sprintf("could not server cgi: %v", err))
	}
}
