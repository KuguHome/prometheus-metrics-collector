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
	//"os/exec"
	"encoding/json"
	"log"

	"gopkg.in/alecthomas/kingpin.v2"
	//"github.com/buger/jsonparser"
)

var (
	inFileFlagArg = kingpin.Flag("something", "Read in a .json file.").File()
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

	// read open bracket
	t, err := dec.Token()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%T: %v\n", t, t)

	for dec.More(){
		var machine Machine
		if err := dec.Decode(&machine); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Host: %v\n", machine.Master.Host)
		fmt.Printf("Port: %v\n", machine.Master.Tunnels[1].Port)
		fmt.Println("");
	}

	// read closing bracket
	t, err = dec.Token()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%T: %v\n", t, t)

}
