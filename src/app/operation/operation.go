package operation

import (
	"github.com/influxdata/influxdb-client-go/v2/api"
	"fmt"
	"os"
	"encoding/csv"
	log "github.com/sirupsen/logrus"
	"time"
	"strings"
)

func CreateInfluxDBCSV(file_name string,result *api.QueryTableResult, measurement_list []string) {
	//open file
	f,err := os.Create(file_name)
	checkError("Error", err)

	//setting headers
	headers := fmt.Sprintf("%s%s", "_time,", strings.Join(measurement_list, ","))
	fmt.Fprintln(f,headers)
	f.Close()

	//open and write to csv file
	file, err := os.OpenFile(file_name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	checkError("Cannot open the file", err)
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// Iterate over query response
	for result.Next() {
		data := []string{}
		t := result.Record().Time()
		duration := fmt.Sprint(t.Format(time.RFC3339))
		data = append(data,duration)
		for _,value:= range measurement_list{
			measurement := fmt.Sprint(result.Record().ValueByKey(value))
			data = append(data,measurement)
		}
		e := writer.Write(data)
		log.Infof("data = %v",data)
		checkError("Cannot write to file", e)
	}
	// check for an error
	if result.Err() != nil {
		log.Error("query parsing error: %s\n", result.Err().Error())
	}
}

func CreatePostgresCSV(file_name string,result [][]string, measurement_list []string) {
	//open file
	f,err := os.Create(file_name)
	checkError("Error", err)

	//setting headers
	headers := fmt.Sprintf("%s%s", "_time,", strings.Join(measurement_list, ","))
	fmt.Fprintln(f,headers)
	f.Close()

	//open and write to csv file
	file, err := os.OpenFile(file_name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	checkError("Cannot open the file", err)
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _,x := range result{
		e := writer.Write(x)
		checkError("Cannot write to file", e)
	}
	   
}

func checkError(message string, err error) {
    if err != nil {
        log.Fatal(message, err)
    }
}