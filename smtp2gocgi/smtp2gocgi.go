package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
)

type Email struct {
	Email_id string
	Event    string
}

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

	logf, logerr := os.OpenFile("/var/log/smtp2go/smtp2go.error", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logerr != nil {
		panic(fmt.Sprintf("File write %v\n", logerr))
	}
	defer logf.Close()

	var email Email
	err = json.Unmarshal([]byte(string(body)), &email)
	if err == nil {
		cmd := exec.Command("/usr/bin/mosquitto_pub", "-r", "-t", "homeassistant/sensor/smtp/"+email.Email_id+"/config", "-m", "{"+
			"\"name\": \"email-mqtt-"+email.Email_id+"\","+
			"\"icon\": \"mdi:email\","+
			"\"expire_after\": 2628000,"+
			"\"state_topic\": \"homeassistant/sensor/smtp/"+email.Email_id+"/state\","+
			"\"json_attributes_topic\": \"homeassistant/sensor/smtp/"+email.Email_id+"/attributes\"}")
		out, cmderr := cmd.CombinedOutput()
		if cmderr != nil {
			fmt.Fprintf(logf, "Exec %v %s\n", cmderr, out)
		}

		cmd = exec.Command("/usr/bin/mosquitto_pub", "-r", "-t", "homeassistant/sensor/smtp/"+email.Email_id+"/state", "-m", email.Event)
		out, cmderr = cmd.CombinedOutput()
		if cmderr != nil {
			fmt.Fprintf(logf, "Exec %v %s\n", cmderr, out)
		}

		cmd = exec.Command("/usr/bin/mosquitto_pub", "-r", "-t", "homeassistant/sensor/smtp/"+email.Email_id+"/attributes", "-m", string(body))
		out, cmderr = cmd.CombinedOutput()
		if cmderr != nil {
			fmt.Fprintf(logf, "Exec %v %s\n", cmderr, out)
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
