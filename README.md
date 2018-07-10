# metrics-collector
This is a program currently installed on our central component server (zentralkomponente, zk). It is running automatically every minute. It logs into each control unit through an HTTP tunnel, collects all of the metrics on the machine, relabels them, and pushes them to the Prometheus Pushgateway.

## How to Compile

To compile this program, cd into src/metrics-collector and run the following:
```
go build
```
If a binary is needed in another environment, lead the command with `env`, followed by `GOOS=<target_OS>` and `<GOARCH=target_architecture>`, and then finally with `go build`. For example, the following compiles for a Linux operating system with the AMD64 architecture:
```
env GOOS=linux GOARCH=amd64 go build
```

This will make an executable, ‘metrics-collector’. After, the binary can be copied to any desired place. The following copies the binary to the system path, which can be done by copying to /usr/local/bin:
```
cp metrics-collector /usr/local/bin
```


### Details
This program reads in a json file containing a list of control units and their information. For each control unit, the program logs into the control unit through an HTTP tunnel. Upon connecting, the program deletes all old metrics in the event of a server cut. Then, for each read path requested, it calls a separate program component, the relabeler, which add and drops metrics as requested. The relabeled metrics are then pushed to the Prometheus Pushgateway. Additionally, for each read path, a metric is added under the metric family metrics_collector_target_up that takes the value 1 if the get request succeeded and 0 if the get request failed.

The relabeler itself is capable of reading in from STDIN, files, and directories. The HTTP get requests are sent to the relabeler as a byte stream. The relabeler can add and drop labels, drop a set of 38 metrics, and write to a reqested file, or STDOUT if none is specified. For our metrics-collector, the relabeled text is written to STDOUT, captured as a byte array, and then pushed to the Prometheus Pushgateway through HTTP.

This program also has the option of printing logging information to the terminal for troubleshooting.

## Command Line

### Flags
`--json <file_name>`
Read in from a json file \<file_name\>

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


#### Example
This is an example call to the program from the command line. "sz.json.conf" is the file containing the information of all of the control units.
```
./metrics-collector --json sz.json.conf push-label job=node machine_type=sz --delete-old --push-url http://localhost:9091/metrics --read-path /static/metrics/node_exporter.prom --read-path /static/metrics/openhab.prom --machine-label machine --log
```

**Terminal Output:**
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

## Development/Build Setup
This program uses the language Golang. Go to the following website for installation instructions:
```
https://golang.org/doc/install
```

Additionally, dep is used for dependency handling. Go to the following website for installation instructions:
```
https://golang.github.io/dep/docs/installation.html
```

With dep installed, cd into metrics-collector and perform:
```
dep ensure
```
