package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
	"flag"
	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	"google.golang.org/protobuf/proto"
)

var logLevel slog.Level = slog.LevelError

func main() {
	var logLvlFlag int = 0
	flag.IntVar(&logLvlFlag, "v", 0, "0: LevelError\n1: LevelWarn\n2: LevelInfo\n3: LevelDebug")
	switch logLvlFlag {
	case 0:
		logLevel = slog.LevelError
	case 1:
		logLevel = slog.LevelWarn
	case 2:
		logLevel = slog.LevelInfo
	case 3:
		logLevel = slog.LevelDebug
	}

	startTime := time.Now()
	os.MkdirAll("out", 0755)

	var err error
	logFile, err := os.Create("out/log.log")
	if err != nil {
		logFile = os.Stdout
	}
	// default buffer size is 4096
	// 32768 seems optimal for my machine when log level is INFO
	bufSize := 4096
	if len(os.Args) == 2 {
		bufSize, err = strconv.Atoi(os.Args[1])
		if err != nil {
			bufSize = 4096
		}
	}
	logWriter := bufio.NewWriterSize(logFile, bufSize)
	defer func() {
		logWriter.Flush()
		logFile.Close()
	}()

	logger := slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "time" {
				return slog.Attr{}
			}
			return slog.Attr{Key: a.Key, Value: a.Value}
		},
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	// The file path to an unzipped demo file.
	file, err := os.Open("in/2025-04-10_Mirage_16-14.dem")
	if err != nil {
		slog.Error("Failed to open demo file", "error", err)
		os.Exit(1)
	}
	defer file.Close()

	parser := dem.NewParser(file)
	defer parser.Close()


	serializedBytes := []byte{}

	header, err := parser.ParseHeader()

	parser.RegisterNetMessageHandler(func(m proto.Message) {
		bytes, err := proto.Marshal(m)
		if (err != nil) {
			slog.Error("Unable to serialize message", "Message", m)
			return
		}
		serializedBytes = append(serializedBytes, bytes...)
	})

	// Parse the full demo file.
	err = parser.ParseToEnd()
	// var moreFrames bool = true
	// for moreFrames {
	// 	moreFrames, err = parser.ParseNextFrame()
	// }
	elapsed := time.Since(startTime)
	slog.Warn(fmt.Sprintf("Parsing took: %v", elapsed))

	// write out serialized demo file
	rewrittenDemoFile, err := os.Create("out/2025-04-10_Mirage_16-14_recomp.dem")
	if (err != nil) {
		slog.Error(err.Error())
	} else {
		defer rewrittenDemoFile.Close()
		_, err = rewrittenDemoFile.Write(serializedBytes)
		if (err != nil) {
			slog.Error(err.Error())
		}
	}	
}