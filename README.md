# dayreport-send
This is a program currently installed on our central component server (zentralkomponente, zk). It is running automatically once a day. It logs into the control units (steuerzentrale, SZ, RaspberryPI), downloads the tagesbericht (daily status reports) pdf files (see https://gitlab.kugu-home.com/projekte/metrics-collector/blob/master/examples/tagesbericht_2018-06-18.pdf) and sends them out to e.g. Leo & Christopher.

## how to compile

1. ```GOPATH="$GOPATH:`pwd`" go get -d ...```
2. ```GOPATH="$GOPATH:`pwd`" go install metrics-collector```

### Details
This program reads in a .json file containing a list of control units and their information. It parses the file, logs into each machine through an HTTP tunnel, and does as described above.

# Command Line

### Flags
`--json <file_name>`
Read in from a .json file \<file_name\>

`--delete-old`
Delete old, repeated scrapes in the event of a server cut

`--machine-label <machine_label>`
Specify the machine label

`--push-url <url>`
specify the paths to read from (include a leading forward slash)

`-a, --add-label <label>=<value>`
The label-value pair \<label\>=\<value\> is added to the incoming text in the correct format. Can be called an arbitrary number of times.

`-d, --drop-metric <some_metric>`
The metric given by some_metric is dropped. Can be called an arbitrary number of times.

`--drop-default`
Drop default metrics

`--in <file_name>`
Read in a .prom file \<file_name\>

`--out <file_name>`
Write out to a file \<file_name\>

`--in-dir <dir_name>`
Read in a directory \<dir_name\> and run the program on each .prom file in the directory. Does not go into sub-directories.

`--log`
Write logs to STDERR

### Commands
`help [command...]`
Show help

`push-label [<push-label-args>...]`
Add name-value pairs to push names in the form <name>=<value>


### Example
This is an example call to the program from the command line. "sz.json" is the file containing the information of all of the control units.
```
./metrics-collector --json sz.json.conf push-label job=node machine_type=sz --delete-old --push-url http://localhost:9091/metrics --read-path /static/metrics/node_exporter.prom --read-path /static/metrics/openhab.prom --machine-label machine --log
```

Terminal Output:
```
2018/07/10 07:45:45 Starting collection from lkuttner...
2018/07/10 07:45:45 Deleting old metrics from http://localhost:9091/metrics/job/node/machine_type/sz/machine/lkuttner
2018/07/10 07:45:45 Success
2018/07/10 07:45:45 Attempting GET from http://localhost:2020/static/metrics/node_exporter.prom
2018/07/10 07:45:45 Success
2018/07/10 07:45:45 Relabeling metrics...
2018/07/10 07:45:45 Relabeling complete
2018/07/10 07:45:45 Attempting POST to http://localhost:9091/metrics/job/node/machine_type/sz/machine/lkuttner
2018/07/10 07:45:45 Success
2018/07/10 07:45:45 Attempting GET from http://localhost:2020/static/metrics/openhab.prom
2018/07/10 07:45:45 Failure: 404 Not Found. Continuing...
2018/07/10 07:45:45 Attempting POST to http://localhost:9091/metrics/job/node/machine_type/sz/machine/lkuttner
2018/07/10 07:45:45 Success
2018/07/10 07:45:45 Collection from lkuttner complete

```

# Development/Build Setup
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
