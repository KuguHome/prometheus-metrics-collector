package main

import (
  "os"
  "io"
  "bufio"
  "io/ioutil"
  "log"
  "strings"
  "bytes"

  dto "github.com/prometheus/client_model/go"

  "github.com/prometheus/common/expfmt"
  "github.com/golang/protobuf/proto"
  )

  //need this stuct to allow data to be passed outside of the scope of the function without explicitly having to create onoxious parameters
  type Relabeler struct {
    OutBytes []byte
  }

  //set up the flags
  var (
    //so fields are immediately available for helper methods
    labelArgs *map[string]string
    dropArgs *[]string
    inFileArg **os.File
    outFileArg *string
    defaultDrop *bool
    inDirArg *string

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

func (r *Relabeler) relabel(labelFlagArgs *map[string]string, dropFlagArgs *[]string, inFileFlagArg **os.File, outFileFlagArg *string, defaultDropFlag *bool, inDirFlagArg *string, inStream io.Reader) {
  //parses command line flags into a key=value map

  labelArgs = labelFlagArgs
  dropArgs = dropFlagArgs
  inFileArg = inFileFlagArg
  outFileArg = outFileFlagArg
  defaultDrop = defaultDropFlag
  inDirArg = inDirFlagArg

  var writer io.Writer
  if *outFileArg != "" {
    var err error
    writer, err = os.Create(*outFileArg)
    if err != nil {
        log.Fatal(err)
    }
  } else {
    writer = os.Stdout
  }
  //goes through all cases of possible readers
  if (*inFileArg == nil) && (*inDirArg == "") {
    //case that this needs to take in a stream of bytes and then capture the bytes to use as input for something else
    //e.g. between when the metrics collector gets and posts and it isn't as simple as stdin and stdout
    if inStream != nil {
      buf := new(bytes.Buffer)
      parseAndRebuild(inStream, buf)
      r.OutBytes = buf.Bytes()
    } else {
      parseAndRebuild(os.Stdin, writer)
    }
  } else {
    if *inFileArg != nil && strings.HasSuffix((*inFileArg).Name(), ".prom") {
      reader := bufio.NewReader(*inFileArg)
      parseAndRebuild(reader, writer)
    }
    //directory with .prom files
    if *inDirArg != "" {
      filesInfo, err := ioutil.ReadDir(*inDirArg)
      if err != nil {
          log.Fatal(err)
      }
      for _, info := range filesInfo {
        if strings.HasSuffix(info.Name(), ".prom") {
          reader, err := os.Open("" + *inDirArg + "/" + info.Name())
          if err != nil {
              log.Fatal(err)
          }
          parseAndRebuild(reader, writer)
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
  for key, value := range *labelArgs {
        pairs = append(pairs, &dto.LabelPair{
          Name:  proto.String(key),
				  Value: proto.String(value),
        })
      }
      return pairs
}

//parses a stream coming in from readFrom, adds and drops metrics, rebuilds it, and writes it to writeTo
func parseAndRebuild(readFrom io.Reader, writeTo io.Writer) {
  //creates TextParser and parses text into metrics
  var parser expfmt.TextParser

  parsedFamilies, _ := parser.TextToMetricFamilies(readFrom)
  validPairs := pairArgsToSlice()

  //add the default drop metrics to the list of metrics to be dropped
  if *defaultDrop {
    *dropArgs = append(*dropArgs, defaultFlags...)
  }

  //delete metrics requested to be dropped
  for _, name := range *dropArgs {
    delete(parsedFamilies, name)
  }

  writeOut(parsedFamilies, validPairs, writeTo)
}
