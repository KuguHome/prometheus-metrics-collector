package main

import (
  "os"
  "io"
  "bufio"
  "io/ioutil"
  "strings"
  "bytes"
  "fmt"

  dto "github.com/prometheus/client_model/go"

  "github.com/prometheus/common/expfmt"
  "github.com/golang/protobuf/proto"
  )

  //need this stuct to allow data to be passed outside of the scope of the function without explicitly having to create obnoxious parameters
  type Relabeler struct {
    OutBytes []byte
    extraMetricFamilies []*dto.MetricFamily
    GetSuccess bool
  }

  //struct to hold a label, value, and float64 so that they can all be grouped under one variadic parameter in addGaugeMetrics
  type LabelValueFloat struct {
  	Label string
  	Value string
  	Float float64
  }

  //set up the flags
  var (
    //so fields are immediately available for helper methods

    defaultFlags = []string{
  		"go_memstats_last_gc_time_seconds",
  		"go_goroutines",
  		"go_memstats_other_sys_bytes",
  		"go_gc_duration_seconds",
  		"process_virtual_memory_bytes",
  		"go_memstats_heap_inuse_bytes",
  		"process_open_fds",
  		"go_memstats_heap_alloc_bytes",
  		"go_threads",
  		"go_memstats_mcache_inuse_bytes",
  		"process_max_fds",
  		"go_memstats_alloc_bytes",
  		"http_response_size_bytes",
  		"process_start_time_seconds",
  		"go_memstats_heap_released_bytes",
  		"go_memstats_sys_bytes",
  		"go_memstats_heap_idle_bytes",
  		"process_resident_memory_bytes",
  		"go_memstats_mcache_sys_bytes",
  		"go_memstats_frees_total",
  		"go_memstats_heap_objects",
  		"go_memstats_next_gc_bytes",
  		"go_memstats_buck_hash_sys_bytes",
  		"go_memstats_stack_sys_bytes",
  		"go_memstats_heap_sys_bytes",
  		"go_memstats_mspan_inuse_bytes",
  		"go_memstats_gc_cpu_fraction",
  		"go_memstats_stack_inuse_bytes",
  		"http_request_duration_microseconds",
  		"go_memstats_mspan_sys_bytes",
  		"go_info",
  		"go_memstats_gc_sys_bytes",
  		"http_requests_total",
  		"go_memstats_lookups_total",
  		"process_cpu_seconds_total",
  		"go_memstats_mallocs_total",
  		"go_memstats_alloc_bytes_total",
  		"http_request_size_bytes"}
  )

func (r *Relabeler) relabel(inStream io.Reader) {
  //assign writer
  var writer io.Writer
  if *outFileFlagArg != "" {
    var err error
    writer, err = os.Create(*outFileFlagArg)
    if err != nil {
        logFatalf("Output file error: ", err)
    }
  } else {
    writer = os.Stdout
  }
  //goes through all cases of possible readers
  if (*inFileFlagArg == nil) && (*inDirFlagArg == "") {
    //case that this needs to take in a stream of bytes and then capture the bytes to use as input for something else
    //e.g. between when the metrics collector gets and posts and it isn't as simple as stdin and stdout
    if inStream != nil {
      //captures bytes that would otherwise just go to Stdout and into oblivion
      buf := new(bytes.Buffer)
      r.parseAndRebuild(inStream, buf, r.extraMetricFamilies)
      r.OutBytes = buf.Bytes()
    } else {
      r.parseAndRebuild(os.Stdin, writer, r.extraMetricFamilies)
    }
  } else {
    if *inFileFlagArg != nil && strings.HasSuffix((*inFileFlagArg).Name(), ".prom") {
      reader := bufio.NewReader(*inFileFlagArg)
      r.parseAndRebuild(reader, writer, r.extraMetricFamilies)
    } else {
      logPrintf("%s is not a .prom file\n", *inFileFlagArg)
    }
    //directory with .prom files
    if *inDirFlagArg != "" {
      filesInfo, err := ioutil.ReadDir(*inDirFlagArg)
      if err != nil {
          logPrintf("Error with input directory: %s\n", err)
      } else{
        for _, info := range filesInfo {
          openFile := fmt.Sprintf("%s/%s", *inDirFlagArg, info.Name())
          if strings.HasSuffix(info.Name(), ".prom") {
            reader, err := os.Open(openFile)
            if err != nil {
              logPrintf("Error opening %s: %s", openFile, err)
            }
            r.parseAndRebuild(reader, writer, r.extraMetricFamilies)
          } else {
            logPrintf("%s is not a .prom file\n", openFile)
          }
        }
      }
    }
  }
}

//rebuild the text with the new labels and write to writeTo
func writeOut(families map[string]*dto.MetricFamily, labelPairs []*dto.LabelPair, writeTo io.Writer) {
  for _, metricFamily := range families {
    for _, metric := range metricFamily.Metric {
      metric.Label = append(metric.Label, labelPairs...)
    }
    expfmt.MetricFamilyToText(writeTo, metricFamily)
  }
}

//converts key-value map into LabelPair slice
func pairArgsToSlice() []*dto.LabelPair {
  var pairs []*dto.LabelPair
  for key, value := range *labelFlagArgs {
        pairs = append(pairs, &dto.LabelPair{
          Name:  proto.String(key),
				  Value: proto.String(value),
        })
      }
      return pairs
}

//parses a stream coming in from readFrom, adds and drops metrics, rebuilds it, and writes it to writeTo
func (r *Relabeler) parseAndRebuild(readFrom io.Reader, writeTo io.Writer, extraMetricFamilies []*dto.MetricFamily) {
  //creates TextParser and parses text into metrics
  var parser expfmt.TextParser

  parsedFamilies, err := parser.TextToMetricFamilies(readFrom)

  if r.GetSuccess {
    //only print this message if the get was successful to begin with, otherwise the message is meaningless and will be confusing
    if err != nil {
  			logPrintf("function parseAndRebuild() error: %s\n", err)
  		}
  }

  //for each device, add extra metrics
  parsedFamilies = addFamilies(parsedFamilies, extraMetricFamilies)

  validPairs := pairArgsToSlice()

  //add the default drop metrics to the list of metrics to be dropped
  if *defaultDropFlag {
    *dropFlagArgs = append(*dropFlagArgs, defaultFlags...)
  }

  //delete metrics requested to be dropped
  for _, name := range *dropFlagArgs {
    delete(parsedFamilies, name)
  }

  writeOut(parsedFamilies, validPairs, writeTo)
}

//adds the metric family in with the string for parsing. i still don't know why it needs a string-metricfamily map
func addFamilies (a map[string]*dto.MetricFamily, b []*dto.MetricFamily) map[string]*dto.MetricFamily {
  num := 1
  //change the key with every additional metric family so that it doesn't just overwrite
  for _, family := range b {
    numstr := fmt.Sprintf("new%d", num)
    a[numstr] = family
    num++
  }
  return a
}

//creates a new gauge metric family, assigned to the respective field in the target struct
func (r *Relabeler) newGaugeMetricFamily(name string, help string) *dto.MetricFamily {
  metricFamily := &dto.MetricFamily{
    Name: &name,
    Help: &help,
    Type: dto.MetricType_GAUGE.Enum(),
    Metric: []*dto.Metric{},
  }
  r.extraMetricFamilies = append(r.extraMetricFamilies, metricFamily)
  return metricFamily
}

//add a new metric to the metric family pointed to by family, with the label, value, and floats desired
func addGaugeMetrics(family *dto.MetricFamily, labelvaluefloat ...LabelValueFloat) {
  for _, lvf := range labelvaluefloat{
    metric := &dto.Metric{
      Label: []*dto.LabelPair{
        &dto.LabelPair{
        Name:  proto.String(lvf.Label),
        Value: proto.String(lvf.Value),
      },
    },
      Gauge: &dto.Gauge{
        Value: proto.Float64(lvf.Float),
      },
    }
    family.Metric = append(family.Metric, metric)
  }
}

func (r *Relabeler) setGetSuccess(b bool) {
  r.GetSuccess = b
}
