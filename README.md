**Structure-**
This structure is inspired from clean code architecture
```bash
│
├── src // put all handlers here
│   ├── app
│   │   │── config
│   │   │   │── yaml 
│   │   │   │   └── config.yaml //define custom config for the app
│   │   │   │── json 
│   │   │   │   └── metrics.json //define metrics details
│   │   │   └── config.go  //config read
│   │   │── logging
│   │   │   │── appcontext 
│   │   │   │   └── appContext.go // context management across app
│   │   │   └── auditlog.go // framework for audit log on every request
│   │   │── operation
│   │   │   └── operation.go //save results in csv file
│   │   │── output
│   │   │   └── log // contains logs
│   │	│── rest
│   │	│   │── generated //swagger gen code 
│   │	│   │── handler
│   │	│   │   │── logger.go // logger initializer
│   │	│   │   │── myApp.go // rest api handler
│   │	│   │   └── responseHandler.go   //marshal and respone of type success/error
│   │	│   │── api.yaml //openapi v3 yaml - restapi definition
│   │	│   │── apiserver.go // rest api server listener
│   │	│   └── routers.go  //define navigation for rest apis
│   │   │── utils
│   │   │   │── influxUtil.go  //Influx query generator
│   │   │   │── postgresUtil.go  //Postgres query generator
│   │   │   │── kafkaUtil.go  //Kafka utils
│   │   │   │── commonUtil.go  //common utils
│   │   │   └── httpUtils.go //json response generator 
└───└───└── main.go //entry point   
```

**Installation-**
1)  Go 1.13 or later is required.
2)  Few of the libraries that are used as part of this project are
    - github.com/sirupsen/logrus - logging library
    - gopkg.in/yaml.v2 - Yaml config reader
    - github.com/google/uuid - Unique ID library used in audit log framework
    - gopkg.in/natefinch/lumberjack.v2 - Lumberjack is a Go package for writing logs to rolling files
    - github.com/rifflock/lfshook - Local Filesystem Hook for Logrus
    - github.com/influxdata/influxdb-client-go/v2 - InfluxDB Client Go
3)  Apache2 at port 8000


**Rest APIs-**
1) **Get Idracs List-**
**url -** http://100.67.30.48:8080/v1/idracs/
**headers -** Content-Type:application/json
**method -** GET
**response -**
        {
          "idrac": ["10.239.56.12", "100.67.30.26", "100.67.30.133", "100.67.30.162"]
        }

2) **Get Metrics List-**
**url -** http://100.67.30.48:8080/v1/metrics/
**headers -** Content-Type:application/json
**method -** GET
**response -**
        {
          "measurements": [{
            "measurement": ["dell_poweredge_SystemAvgInletTempHour", "dell_poweredge_SystemMaxInletTempHour"],
            "metrics": "AggregationMetrics"
          }, {
            "measurement": ["dell_poweredge_AvgFrequencyAcrossCores", "dell_poweredge_CPUC0ResidencyHigh"],
            "metrics": "CPUMemMetrics"
          }, {
            "measurement": ["dell_poweredge_TemperatureReading"],
            "metrics": "CPUSensor"
          }, {
            "measurement": ["dell_poweredge_RPMReading"],
            "metrics": "FanSensor"
          }]
        }


3) **Get Metrics Data-**
**url -** http://100.67.30.48:8080/v1/idracs/metrics/
**headers-** Content-Type:application/json
**method -** POST
**payload -**
**Method 1-**
        {
        "bucket": "telegraf",
        "measurement": ["dell_poweredge_AggregateUsage", "dell_poweredge_CPUUsage", "dell_poweredge_IOUsage"],
        "metricreport": "SystemUsage",
        "host": "100.67.30.193",
        "duration": "1h"
        }

**Method 2-**
        {
        "bucket": "telegraf",
        "measurement": ["dell_poweredge_SystemAvgInletTempHour", "dell_poweredge_SystemMaxInletTempHour"],
        "metricreport": "AggregationMetrics",
        "host": "100.67.30.193",
        "start": "2022-01-30T00:00:00Z",
        "stop": "2022-02-01T00:00:00Z"
        }
**response -**
        {
        "File name": "http://100.67.30.48:8000/AggregationMetrics-100.67.30.193.csv",
        "Idrac": "100.67.30.193"
        }

**Note-**
1)  **Examples of duration types**
1ns // 1 nanosecond
1us // 1 microsecond
1ms // 1 millisecond
1s  // 1 second
1m  // 1 minute
1h  // 1 hour
1d  // 1 day
1w  // 1 week
1mo // 1 calendar month
1y  // 1 calendar year
3d12h4m25s // 3 days, 12 hours, 4 minutes, and 25 seconds

2)  **Examples for particular time duration**
    YYYY-MM-DD
    YYYY-MM-DDT00:00:00Z
    YYYY-MM-DDT00:00:00.000Z

