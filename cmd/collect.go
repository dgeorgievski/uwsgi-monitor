package cmd

import (
	"fmt"
	"strings"
	"log"
	"os"
	// "time"
	"reflect"

	"github.com/dgeorgievski/uwsgi-monitor/metrics"
	"github.com/spf13/cobra"
)

func iterateMetrics(root string, metrics interface{}) {
		mets := reflect.ValueOf(metrics)

		typeOfM := mets.Type()

		var new_root string
		switch kind := mets.Kind(); kind {
		case reflect.Struct:
			for i := 0; i < mets.NumField(); i++ {
				fld := mets.Field(i)
				new_root = fmt.Sprintf("%s_%s", root, strings.ToLower(typeOfM.Field(i).Name))
				iterateMetrics(new_root, fld.Interface())

			}

		case reflect.Slice:
			for i:=0; i<mets.Len(); i++ {
				iterateMetrics(root, mets.Index(i).Interface())
			}

		case reflect.Map:
			keys := mets.MapKeys()
			for i:=0; i<len(keys); i++ {
				iterateMetrics(root, mets.MapIndex(keys[i]).Interface())
			}

		default:
			fmt.Printf("%s %v \n", root, mets.Interface())
		}
}

/****************************************************
*
****************************************************/
func getMetrics(uwsgiAddress string, uwsgiPort string) {
		ret := metrics.GetMetrics(uwsgiAddress, uwsgiPort)

		if ret.HttpCode != 200 || ret.Message != "OK" {
				log.Fatal("Code: ", ret.HttpCode, " Message: ", ret.Message)
				os.Exit(1)
		}

		// iterateMetrics("uwsgi", ret.Metrics)
		labels := make(map[string]string)
		labels["service"] = "els"

	  // tstamp := fmt.Sprintf("%d", time.Now().Unix())
		tstamp := ""
		ret.Uwsgi.PrintMetrics("uwsgi", os.Stdout, labels, tstamp)

}

// collectCmd represents the collect command
var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect metrics from uwsgi server.",
	Long: `Collect uwsgi metrics, print them to stdout, and exit.`,
	Run: func(cmd *cobra.Command, args []string) {

		getMetrics(uwsgiAddress, uwsgiPort)
	},
}

func init() {
	rootCmd.AddCommand(collectCmd)
}
