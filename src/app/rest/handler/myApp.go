package handler

import (
	"fmt"
	"net/http"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"dataframe-service/src/app/utils"
	"dataframe-service/src/app/operation"
	"dataframe-service/src/app/config"
	"io/ioutil"
	"os"
)

type data struct{
	Bucket string `json:"bucket"`
	Measurement []string `json:"measurement"`
	MetricReport string `json:"metricreport"`
	Host string `json:"host"`
	Duration string `json:"duration"`
	Start string `json:"start"`
	Stop string	`json:"stop"`
}

//Get idracs metrics
func IdracsMetrics(w http.ResponseWriter, r *http.Request) {
	setupHeader(w,r)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	success := true
	statusMsg := ""
	alog, _ := LogInit(r, "Get Idrac Metrics")
	defer alog.LogMessageEnd(&success, &statusMsg)
	alog.LogMessageReceived()
	var i data
	reqBody, _ := ioutil.ReadAll(r.Body)
    log.Infof("req = %v",string(reqBody))
	json.Unmarshal(reqBody, &i)
	log.Infof("i = %v",i)

	file_path := config.GetConfig().DataframeConfig.File.File_path
	file_name := fmt.Sprintf("%s-%s.csv",i.MetricReport,i.Host)
	log.Infof("file name = %v",file_name)
	
	if config.GetConfig().DataframeConfig.Destination.ActiveDb == "influxdb"{
		//Initialize Query Param
		inQ := utils.NewInfluxDB(i.Bucket,i.Measurement,i.MetricReport,i.Host,i.Duration,i.Start,i.Stop)
		log.Infof("influxQuery = %v",inQ)

		//Execute Query
		result := inQ.ExecuteQuery()
		log.Infof("result = %v",result)
		//Create CSV
		operation.CreateInfluxDBCSV(file_path+file_name,result,i.Measurement)

	} else if config.GetConfig().DataframeConfig.Destination.ActiveDb == "postgres"{
		//Initialize Query Param
		pgQ := utils.NewPostgres(i.Measurement,i.MetricReport,i.Host,i.Duration,i.Start,i.Stop)
		log.Infof("postgresQuery = %v",pgQ)

		//Execute Query
		result := pgQ.ExecuteQuery()
		
		//Create CSV
		operation.CreatePostgresCSV(file_path+file_name,result,i.Measurement)
	} else if config.GetConfig().DataframeConfig.Destination.ActiveDb == "kafka"{

		//Need to configured...

	}

	file_name_res := map[string]string{
		"idrac": i.Host,
		"csv": config.GetConfig().DataframeConfig.File.File_url + file_name,
	}
	HandleSuccessResponse(w, file_name_res)
	statusMsg = "Get successful"
}

//Get idracs list
func IdracsList(w http.ResponseWriter, r *http.Request) {
	setupHeader(w,r)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	success := true
	statusMsg := ""
	alog, _ := LogInit(r, "Get Idrac List")
	defer alog.LogMessageEnd(&success, &statusMsg)
	alog.LogMessageReceived()
	idrac_list := utils.GetIdracs()
	HandleSuccessResponse(w, idrac_list)
	statusMsg = "Get successful"
}


//Get metrics list
func MetricsList(w http.ResponseWriter, r *http.Request) {
	setupHeader(w,r)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	success := true
	statusMsg := ""
	alog, _ := LogInit(r, "Get Metrics List")
	defer alog.LogMessageEnd(&success, &statusMsg)
	alog.LogMessageReceived()
	
	// Open our jsonFile
    jsonFile, err := os.Open("config/json/metrics.json")
    // if we os.Open returns an error then handle it
    if err != nil {
        log.Error(err)
    }
    log.Info("Successfully Opened metrics.json")
    // defer the closing of our jsonFile so that we can parse it later on
    defer jsonFile.Close()

    byteValue, _ := ioutil.ReadAll(jsonFile)

    var metrics_list map[string]interface{}
    json.Unmarshal([]byte(byteValue), &metrics_list)
	log.Infof("Metric list = %v",metrics_list)
	HandleSuccessResponse(w, metrics_list)
	statusMsg = "Get successful"
}

func setupHeader(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	rw.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
