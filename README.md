# dayreport-send
This is a program currently installed on our central component server (zentralkomponente, zk). It is running automatically once a day. It logs into the control units (steuerzentrale, SZ, RaspberryPI), downloads the tagesbericht (daily status reports) pdf files (see https://gitlab.kugu-home.com/projekte/metrics-collector/blob/master/examples/tagesbericht_2018-06-18.pdf) and sends them out to e.g. Leo & Christopher.

## how to compile

1. ```GOPATH="$GOPATH:`pwd`" go get -d ...```
2. ```GOPATH="$GOPATH:`pwd`" go install metrics-collector```

### Details
This program reads in a .json file containing a list of control units and their information. It parses the file, logs into each machine through an HTTP tunnel, and does as described above.

### Command Line
`--json`
Read in from a .json file "file_name"

`--delete-old`
Delete old, repeated scrapes in the event of a server cut

### Example
This is an example call to the program from the command line. "sz.json" is the file containing the information of all of the control units.
```
./metrics-collector --delete-old --json sz.json
```

### Development/Build Setup
This program uses the language Golang. Go to the following website for installation instructions:
```
https://golang.org/doc/install
```

Additionally, dep is used for dependency handling. Go to the following website for installation instructions:
```
https://golang.github.io/dep/docs/installation.html
```

### Making it Runnable From the Command Line
Compile the program with the following
```
go build -o metrics-collector main.go
```

This will make an executable, ‘metrics-collector’. After, the program can be copied to the system path, which can be done by copying to /usr/local/bin:
```
cp metrics-collector /usr/local/bin
```
