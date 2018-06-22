package main

// import (
//   "github.com/buger/jsonparser"
//   "fmt"
// )
//
// func main() {
//   data := []byte(`{
//     "person": {
//       "name": {
//         "first": "Leonid",
//         "last": "Bugaev",
//         "fullName": "Leonid Bugaev"
//       },
//       "github": {
//         "handle": "buger",
//         "followers": 109
//       },
//       "avatars": [
//         { "url": "penis", "type": "thumbnail" },
// 				{ "url": "morepenis", "type": "thumbnail" }
//       ]
//     },
//     "company": {
//       "name": "Acme"
//     }
//   }`)
//
//   jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
//   	fmt.Println(jsonparser.GetString(value, "url"))
//   }, "person", "avatars")
//
// }

import (
	//"log/syslog"
	//"os"
	"fmt"
	"io"
	"bufio"
	"os/exec"
	"encoding/json"
	"log"

	"gopkg.in/alecthomas/kingpin.v2"
	//"github.com/buger/jsonparser"
)

var (
	inFileFlagArg = kingpin.Flag("json", "Read in a .json file.").File()
	outNameArg = kingpin.Flag("name", "Append a name to the nodes").String()
)

func main() {
	kingpin.Parse()

	dec := json.NewDecoder(bufio.NewReader(*inFileFlagArg))

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

		//this is essentially the parsing part
		if err := dec.Decode(&machine); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		host := machine.Master.Host
		var httpIdx int

		//find the http port
		for idx, master := range machine.Master.Tunnels {
			if string(master.ProtoType) == "http"{
					httpIdx = idx
					break
			}
		}

		port := machine.Master.Tunnels[httpIdx].Port
		cmdstr := fmt.Sprintf("curl -s http://%v:%v/static/metrics/node.txt | relabeler --drop-default-metrics | curl --data-binary @- http://localhost:9091/metrics/job/node/instance/kugu-sz-%s", port, host, *outNameArg)
		outBytes, err := exec.Command(cmdstr).Output()
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
