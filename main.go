// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
	"path/filepath"
	"raditzlawliet/solr-copy-to/cmd"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	colorable "github.com/mattn/go-colorable"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

func init() {
	InitLogger()
}

func main() {
	cmd.Execute()
}

func InitLogger() {

	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: "2006/01/02 15:04:05"})
	log.SetOutput(colorable.NewColorableStdout())

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	// log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

	logName := "solr-to-mgo"
	pathFolder := "log"
	logFileName := logName + ".%Y%m%d-%H%M.log"

	// creating log folder
	_ = os.MkdirAll(pathFolder, 0766)

	writer, err := rotatelogs.New(
		filepath.Join(pathFolder, logFileName),
		rotatelogs.WithLinkName(filepath.Join(pathFolder, logFileName)),
		rotatelogs.WithMaxAge(time.Duration(86400)*time.Second),        // default max 7 days
		rotatelogs.WithRotationTime(time.Duration(604800)*time.Second), // default 24 hour
	)

	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	log.AddHook(lfshook.NewHook(
		lfshook.WriterMap{
			log.DebugLevel: writer,
			log.InfoLevel:  writer,
			log.WarnLevel:  writer,
			log.ErrorLevel: writer,
			log.FatalLevel: writer,
			log.PanicLevel: writer,
		},

		&log.JSONFormatter{},
		// &log.TextFormatter{FullTimestamp: true, TimestampFormat: "2006/01/02 15:04:05"},
	))
}
