package main

import (
	"fmt"
	"io"
	"bufio"
	"os/exec"
	"encoding/json"
	"log"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	inFileFlagArg = kingpin.Flag("json", "Read in a .json file.").File()
)

func main() {
	kingpin.Parse()

	//essentially a parser
	dec := json.NewDecoder(bufio.NewReader(*inFileFlagArg))

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

	//if there are more elements in the array, keep going
	for dec.More(){
		var machine Machine

		//parse with decorder
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

		//execute in command line
		cmdstr := fmt.Sprintf("curl -s http://%v:%v/static/metrics/node_exporter.prom | ./relabeler --drop-default-metrics | curl --data-binary @- http://localhost:9091/metrics/job/node/machine_type/sz/machine/%s", host, port, name)
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
