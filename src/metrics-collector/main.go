package main

import (
	"fmt"
	"io"
	"bufio"
	"os/exec"
	"encoding/json"
	"log"
	"regexp"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	//TODO: add flag --push-label that adds name-value pairs onto each push thing i.e. http://localhost:9091/metrics/{job/node}/machine_type/sz/machine/%s
	inFileFlag = kingpin.Flag("json", "Read in a .json file.").PlaceHolder("file_name").File()
	deleteOldFlag = kingpin.Flag("delete-old", "Delete old, repeated scrapes in the event of a server cut").Bool()
	pushLabelCommand = kingpin.Command("push-label", "Add name-value pairs to push names in the form <name>=<value>")
	pushLabelArgs = pushLabelCommand.Arg("push-label-args", "push arguments").Strings()
	machineLabelFlag = kingpin.Flag("machine-label", "Specify the machine label").PlaceHolder("machine_label").String()
	pushURLFlag = kingpin.Flag("push-url", "Specify the url to read from").PlaceHolder("url").String()
	readPathFlags = kingpin.Flag("read-path", "specify the paths to read from (include a leading forward slash)").PlaceHolder("read_path").Strings()
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

	//set up push path
	var pushPathStr string
	for _, elem := range *pushLabelArgs {
		key, value, err := kvParse(elem)
		if err != nil {
			log.Fatal(err)
		}
		pushPathStr = fmt.Sprintf("%s/%s/%s", pushPathStr, key, value)
	}
	pushPathStr = pushPathStr + "/"

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
		var cmdstr string
		var finalPushPath string
		if *deleteOldFlag {
			finalPushPath = fmt.Sprintf("%s%smachine/%s", *pushURLFlag, pushPathStr, name)
			outBytes, err := exec.Command("curl", "-X", "DELETE", finalPushPath).Output()
			if err != nil {
				log.Fatal(err)
			}
			//fmt.Println("curl -X DELETE %s", finalPushPath)
			fmt.Println(string(outBytes))
		}

		//execute in command line
		cmdstr = fmt.Sprintf("curl -s http://%s:%s%s | ./relabeler --drop-default-metrics | curl --data-binary @- %s", host, port, *readPathFlags, finalPushPath)
		//fmt.Println(cmdstr)
		outBytes, err := exec.Command("bash", "-c", cmdstr).Output()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(outBytes))
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
