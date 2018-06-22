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
	"io/ioutil"
	"os/exec"

	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/buger/jsonparser"
)

var (
	inFileFlagArg = kingpin.Flag("something", "Read in a .json file.").String()
)

func main() {
	kingpin.Parse()

	jsonArr, _ := ioutil.ReadFile(*inFileFlagArg)
	paths := [][]string {
		[]string{"master", "tunnels", "[0]", "port"},
		[]string{"master", "host"},
	}
	var port string
	//var host string

	//iterate through all the machines
	jsonparser.ArrayEach(jsonArr, func(value1 []byte, dataType jsonparser.ValueType, offset int, err error) {
		jsonparser.EachKey(value1, func(idx int, value2 []byte, vt jsonparser.ValueType, err error){
			switch idx {
				case 0:
								port = string(value2)
				case 1:
								outBytes, _ := exec.Command("ssh", "-P", port, string(value2)).CombinedOutput()
								fmt.Println(string(outBytes))
			}
	   }, paths...)
	 })
}
