package utils

import (
	"context"
	log "github.com/sirupsen/logrus"
	"fmt"
	"strings"
	"dataframe-service/src/app/config"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type queryData struct{
	bucket string `bucket name`
	measurement []string `measurement`
	metricreport string `metrics report`
	host_name string `host name`
	duration string `query duration`
	start string `query start time`
	stop string `query stop time`
}

func NewInfluxDB(bucket string,measurement []string,metrics_report string,host_name string,duration string,start string,stop string) queryData{
	e := queryData{
			bucket: bucket,
			measurement: measurement,
			metricreport: metrics_report,
			host_name: host_name,
			duration: duration,
			start: start,
			stop: stop,
		}
	return e
}

//Execute Query
func(qD queryData) ExecuteQuery() *api.QueryTableResult{

	url := config.GetConfig().DataframeConfig.Data.URL
	token := config.GetConfig().DataframeConfig.Data.Token
	organization := config.GetConfig().DataframeConfig.Data.Organization
	
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient(url, token)

	// Ensures background processes finishes
    defer client.Close()

	// Get query client
	queryAPI := client.QueryAPI(organization)

	// Generate Flux Query
	fluxQuery := qD.GenerateQuery()

	// Use default dialect
	result, err := queryAPI.Query(context.Background(), fluxQuery)
	checkError("Cannot get the data", err)

	log.Infof("result = %v",result)
	return result
}

func(qD queryData) GenerateQuery() string{
	var query_duration string
	if qD.stop == ""{
		query_duration = fmt.Sprintf("start: -%s",qD.duration)
	}else{
		query_duration = fmt.Sprintf("start: %s, stop: %s",qD.start,qD.stop)
	}

	query_measurement := fmt.Sprintf("%s%s%s", "r[\"_measurement\"] == \"", strings.Join(qD.measurement, "\" or r[\"_measurement\"] == \""), "\"")
	// Query and get complete result as a string
	fluxQuery := fmt.Sprintf(`from(bucket: "%s")
	|> range(%s)
	|> filter(fn: (r) => r["metric_report"] == "%s")
	|> filter(fn: (r) => %s)
	|> filter(fn: (r) => r["hostname"] == "%s")
	|> keep(columns: ["_time", "_value","_measurement"])
	|> pivot(rowKey:["_time"], columnKey: ["_measurement"], valueColumn: "_value")
	|> sort(columns: ["_time"], desc: false)`,
	qD.bucket,query_duration,qD.metricreport,query_measurement,qD.host_name)
	log.Infof("fluxQuery = %v",fluxQuery)

	return fluxQuery
}
