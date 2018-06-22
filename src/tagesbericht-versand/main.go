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
	//"os/exec"
	"encoding/json"
	"log"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
	//"github.com/buger/jsonparser"
)

var (
	inFileFlagArg = kingpin.Flag("something", "Read in a .json file.").String()
)

func main() {
	kingpin.Parse()

	dec := json.NewDecoder(strings.NewReader(*inFileFlagArg))
	fmt.Println(dec);

	type Tunnel struct {
		ProtoType string
		User string
		Port string
	}

	type Master struct {
		Host string
		Tunnels []Tunnel
		Description string
		Name string
		ID string
	}

	type Machine struct {
		Masters []Master
	}

	// read open bracket
	t, err := dec.Token()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%T: %v\n", t, t)

	for dec.More(){
		var master Master
		if err := dec.Decode(&master); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		host := master.Host
		port := master.Tunnels[1].Port
		fmt.Println("Host: %s", host)
		fmt.Println("Port: %s", port)
		fmt.Println("");
	}

}
