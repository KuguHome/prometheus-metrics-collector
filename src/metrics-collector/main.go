package main

import (
	"fmt"
	"io"
	"bufio"
	"os/exec"
	"encoding/json"
	"log"
	"regexp"
	"net/http"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	inFileFlag = kingpin.Flag("json", "Read in a .json file.").Required().PlaceHolder("file_name").File()
	deleteOldFlag = kingpin.Flag("delete-old", "Delete old, repeated scrapes in the event of a server cut").Bool()
	pushLabelCommand = kingpin.Command("push-label", "Add name-value pairs to push names in the form <name>=<value>")
	pushLabelArgs = pushLabelCommand.Arg("push-label-args", "push arguments").Strings()
	machineLabelFlag = kingpin.Flag("machine-label", "Specify the machine label").Required().PlaceHolder("machine_label").String()
	pushURLFlag = kingpin.Flag("push-url", "Specify the url to read from").Required().PlaceHolder("url").String()
	readPathFlags = kingpin.Flag("read-path", "specify the paths to read from (include a leading forward slash)").Required().PlaceHolder("read_path").Strings()
)

func main() {
	kingpin.Parse()

	//essentially a parser
	dec := json.NewDecoder(bufio.NewReader(*inFileFlag))

	//set up structs for the parser
	type Tunnel struct {
		ProtoType string `json:"type"`
		User string `json:"user"`
		Port string `json:"port"`
	}

	type Master struct {
		Host string `json:"host"`
		Tunnels []Tunnel `json:"tunnels"`
		Description string `json:"description"`
		Name string `json:"name"`
		ID int `json:"id"`
	}

	type Machine struct {
		Master Master `json:"master"`
	}

	// ignore open bracket
	_, err := dec.Token()
	if err != nil {
		log.Fatal(err)
	}

	//set up push path without machine/name
	var pushPathStr string
	for _, elem := range *pushLabelArgs {
		key, value, err := kvParse(elem)
		if err != nil {
			log.Fatal(err)
		}
		pushPathStr = fmt.Sprintf("%s/%s/%s", pushPathStr, key, value)
	}

	//if there are more elements in the array, keep going
	for dec.More(){
		var machine Machine

		//parse with decoder
		if err := dec.Decode(&machine); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		//find relevant fields for command line args
		host := machine.Master.Host
		name := machine.Master.Name

		//fullPushPathStr including machine/name
		fullPushPathStr := fmt.Sprintf("%s%s/%s/%s", *pushURLFlag, pushPathStr, *machineLabelFlag, name)

		//find the http port
		var httpIdx int
		for idx, master := range machine.Master.Tunnels {
			if string(master.ProtoType) == "http"{
					httpIdx = idx
					break
			}
		}
		port := machine.Master.Tunnels[httpIdx].Port

		//remove all old metrics if server cuts so we don't a bunch of the same stuff
		if *deleteOldFlag {
			outBytes, err := exec.Command("curl", "-X", "DELETE", fullPushPathStr).Output()
			if err != nil {
				fmt.Printf("DELETE: %s does not already exist. Continuing.\n", fullPushPathStr)
			}
			//fmt.Println("curl -X DELETE %s", fullPushPathStr)
			fmt.Println(string(outBytes))
		}

		for _, path := range *readPathFlags {
			//TODO: no more command line, use http package
			hostStr := fmt.Sprintf("http://%s:%s%s", host, port, path)
			getResp, err := http.Get(hostStr)
    	if err != nil {
        	log.Fatalf("%s", err)
    	}
			_, err = http.Post(fullPushPathStr, "application/octet-stream", getResp.Body)
			if err != nil {
        	log.Fatalf("%s", err)
    	}

			//execute in command line
			//cmdstr = fmt.Sprintf("curl -s http://%s:%s%s | ./relabeler --drop-default-metrics | curl --data-binary @- %s", host, port, elem, fullPushPathStr)
			//fmt.Println(cmdstr)
			//outBytes, err := exec.Command("bash", "-c", cmdstr).Output()
			//if err != nil {
				//probably want it to print something here eventually
			//}
			//fmt.Println(string(outBytes))
		}
	}

	// ignore closing bracket
	_, err = dec.Token()
	if err != nil {
		log.Fatal(err)
	}
}

var strRegex string

func kvParse(str string) (string, string, error) {
	parts := regexp.MustCompile("=").Split(str, 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected KEY=VALUE got '%s'", str)
	}
	return parts[0], parts[1], nil
}
