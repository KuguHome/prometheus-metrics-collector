package main

import (
  "log"
)

//these functions do as the log.Print() functions, except they only execute if --log flag is present

func logPrint(a ...interface{}) {
  if *logFlag{
    log.Print(a...)
  }
}

func logPrintf(str string, v ...interface{}) {
  if *logFlag{
    log.Printf(str, v...)
  }
}

func logPrintln(a ...interface{}) {
  if *logFlag{
    log.Println(a...)
  }
}

func logFatal(a ...interface{}) {
  if *logFlag{
    log.Fatal(a...)
  }
}

func logFatalf(str string, v ...interface{}) {
  if *logFlag{
    log.Fatalf(str, v...)
  }
}
