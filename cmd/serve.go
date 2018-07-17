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

package cmd

import (
	"fmt"
	"time"
	"os"
	"io/ioutil"
	"bufio"

	"github.com/dgeorgievski/uwsgi-monitor/metrics"
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

var (
	serviceName string
	reportAddress string
	reportPort int
	collectFreq int
	metricsPath string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Collect and serve uwsgi metrics over http",
	Long: `Collect metrics from uwsgi in regular intervals,
and serve them over /metrics API endpoint.`,
	Run: func(cmd *cobra.Command, args []string) {
		go collectMetrics(collectFreq)
		serveMetrics()
	},
}

/*
* Collect uwsgi metrics in regular intervals
*/
func collectMetrics(freq int) {
	log.Info("Collecting Metrics")

	// collect immediately before entering the loop
	getUwsgiMetrics(uwsgiAddress, uwsgiPort)

	ticker := time.NewTicker(time.Duration(freq) * time.Second)

	for {
		select {
		case  <- ticker.C:
			getUwsgiMetrics(uwsgiAddress, uwsgiPort)
		}
	}
}


func getUwsgiMetrics(uwsgiAddress string, uwsgiPort string) {

		ret := metrics.GetMetrics(uwsgiAddress, uwsgiPort)

		if ret.HttpCode != 200 || ret.Message != "OK" {
				log.Error("Code: ", ret.HttpCode, " Message: ", ret.Message)
				return
		}

		tmpfile, err := ioutil.TempFile("data", "metrics")
		if err != nil {
	  	log.Error(err)
			return
	  }
		wio := bufio.NewWriter(tmpfile)

		// iterateMetrics("uwsgi", ret.Metrics)
		labels := make(map[string]string)
		labels["service"] = serviceName

	  // tstamp := fmt.Sprintf("%d", time.Now().Unix())
		tstamp := ""
		ret.Uwsgi.PrintMetrics("uwsgi", wio, labels, tstamp)

		// move tmp file to metrics.txt
		tmpFilePath := tmpfile.Name()
		log.Debug("Closing temp file: ", tmpFilePath)
		wio.Flush()
		if err := tmpfile.Close(); err != nil {
			log.Error(err)
			return
		}

		log.Debug("Renaming temp file: ", tmpFilePath, " to data/metrics.txt")
		if err := os.Rename(tmpFilePath, "data/metrics.txt"); err != nil {
	    log.Error(err)
	  }
}


/*
* Serve metrics on /metrics endpoint
*/
func serveMetrics() {

	r := gin.Default()

	serverAddress := fmt.Sprintf("%s:%d", reportAddress, reportPort)

	r.StaticFile("/metrics", metricsPath)

	// ping
	r.GET("/ping", func(c *gin.Context) {
 		c.JSON(200, gin.H{
	 		"message": "pong",
 		})
	})

	log.Info("Starting server on ", serverAddress)
	log.Info("Collection frequency (s): ", collectFreq)
	r.Run(serverAddress)
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVar(&serviceName, "service-name", "", "Service name (Required).")
 	serveCmd.MarkFlagRequired("service-name")
	serveCmd.Flags().StringVar(&reportAddress, "report-address", "0.0.0.0", "host address for reporting metrics.")
	serveCmd.Flags().IntVar(&reportPort, "report-port", 6080, "port for reporting metrics.")
	serveCmd.Flags().IntVar(&collectFreq, "collect-freq", 300, "Frequency of metrics collection.")

	// log.SetFormatter(&log.JSONFormatter{})
  log.SetOutput(os.Stdout)

	metricsPath = "data/metrics.txt"
}
