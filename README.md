# Prometheus Metrics Collector and Relabeler
This is a helper program for collecting metrics from field systems and crunching them, before passing on to the [Prometheus push gateway](https://github.com/prometheus/pushgateway) and ultimately scrape into [Prometheus](https://prometheus.io/). It has several input, output and processing modes.
It is currently installed on our gateway server, running every minute via cron. It collects two metrics on each field system through an HTTP tunnel, relabels them, and pushes them to the Pushgateway.

### Details
This program reads in a json file containing a list of field systems and their information. For each system, the program reaches the field system through a reverse SSH tunnel. Before connecting, the program deletes all old metrics, to make a system downtime transparent. Then, for each read path requested, it calls a separate program component, the relabeler, which adds and drops metrics as requested. The relabeled metrics are then pushed to the  Pushgateway. Additionally, for each read path, a metric is added under the metric family metrics_collector_target_up that takes the value 1 if the get request succeeded and 0 if the get request failed.

The included relabeler is capable of reading in from STDIN, files, and directories. The HTTP get requests are sent to the relabeler as a byte stream. The relabeler can add and drop labels, drop a set of 38 metrics (which are the standard metrics of go exporters), and write to a reqested file, or STDOUT if none is specified. For our metrics-collector, the relabeled text is written to STDOUT, captured as a byte array, and then pushed to the Prometheus Pushgateway through HTTP.

This program also has the option of printing logging information to the terminal for troubleshooting.

## Usage options
### Input
* STDIN (Unix pipe)
* single .prom file
* directory of .prom files
* set of field systems via HTTP, given by a .json file

### Output options
* STDOUT (Unix pipe)
* Single file
* Pushgateway via HTTP

### Processing options
* add custom labels
* drop certain metrics
* drop a set of 38 metrics (like `go_info` or `http_response_size_bytes`)

## Applications

The tool could be used to retouch scrapes from different distributed node_exporter instances, before pushing them to the Pushgateway. Normally, there would be a collision between the scrape's metrics and pushgateway's internal metrics, so before pushing, you can drop the colliding metrics via this tool.
You can import and push the data via `curl` and pipe, via cron jobs or HTTP/the json.

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
This is an example call to the program from the command line. "system.json" is the file containing the information of all of the field systems.
```
./metrics-collector --json system.json push-label job=node machine_type=sz --delete-old --push-url http://localhost:9091/metrics --read-path /static/metrics/node_exporter.prom --read-path /static/metrics/openhab.prom --machine-label machine --log
```

**Terminal Output:**
```
2018/07/10 07:45:45 Starting collection from system1...
2018/07/10 07:45:45 Deleting old metrics from http://localhost:9091/metrics/job/node/machine_type/sz/machine/system1
2018/07/10 07:45:45 Success
2018/07/10 07:45:45 Attempting GET from http://localhost:2000/static/metrics/node_exporter.prom
2018/07/10 07:45:45 Success
2018/07/10 07:45:45 Relabeling metrics...
2018/07/10 07:45:45 Relabeling complete
2018/07/10 07:45:45 Attempting POST to http://localhost:9091/metrics/job/node/machine_type/sz/machine/system1
2018/07/10 07:45:45 Success
2018/07/10 07:45:45 Attempting GET from http://localhost:2000/static/metrics/openhab.prom
2018/07/10 07:45:45 Failure: 404 Not Found. Continuing...
2018/07/10 07:45:45 Attempting POST to http://localhost:9091/metrics/job/node/machine_type/sz/machine/system1
2018/07/10 07:45:45 Success
2018/07/10 07:45:45 Collection from system1 complete

```

## Development/Build Setup

This program uses the language [Golang](https://golang.org/doc/install). Installation instructions are on the website.

Additionally, [dep](https://golang.github.io/dep/docs/installation.html) is used for dependency handling. Installation instructions are on the website.
With dep installed, to install the dependencies that this program requires, `cd` into `metrics-collector` and perform:
```
dep ensure
```

## How to Compile

To compile this program, `cd` into `metrics-collector/src/metrics-collector` and run:
```
go build
```
If a binary is needed in another environment, lead the command with `env`, followed by `GOOS=<target_OS>` and `GOARCH=<target_architecture>`, and then finally with `go build`. For example, the following compiles for a Linux operating system with the AMD64 architecture:
```
env GOOS=linux GOARCH=amd64 go build
```
This will make an executable, `metrics-collector`. After, the binary can be copied to any desired place. The following copies the binary to the system path, which can be done by copying to /usr/local/bin:
```
cp metrics-collector /usr/local/bin
```
