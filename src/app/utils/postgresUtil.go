package utils

import (
	"context"
	log "github.com/sirupsen/logrus"
	"fmt"
	"strings"
	"dataframe-service/src/app/config"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type postgresQueryData struct{
	measurement []string `measurement`
	metricreport string `metrics report`
	host_name string `host name`
	duration string `query duration`
	start string `query start time`
	stop string `query stop time`
}

func NewPostgres(measurement []string,metrics_report string,host_name string,duration string,start string,stop string) postgresQueryData{
	e := postgresQueryData{
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
func(qD postgresQueryData) ExecuteQuery() [][]string {

	pg_user := config.GetConfig().DataframeConfig.Postgres.PgUser
	pg_password := config.GetConfig().DataframeConfig.Postgres.PgPassword
	pg_host := config.GetConfig().DataframeConfig.Postgres.PgHost
	pg_port := config.GetConfig().DataframeConfig.Postgres.PgPort
	pg_db := config.GetConfig().DataframeConfig.Postgres.PgDb
	
	ctx := context.Background()
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
	pg_user,pg_password,pg_host,pg_port,pg_db)
	log.Infof("Connection string = %v",connStr)

	dbpool, err := pgxpool.Connect(ctx, connStr)
	checkError("Unable to connect to database", err)
	defer dbpool.Close()

	// Generate SQL Query
	query := qD.GenerateQuery()

	//Execute query on TimescaleDB
	rows, e := dbpool.Query(ctx, query)
	checkError("Unable to execute query", e)

	if e != nil {
		log.Error("Error to get rows %v",e)
	}
	defer rows.Close()
	log.Infof("Successfully executed query and rows = %v",rows)

	var result [][]string
	for rows.Next() {
		data := []string{}
		val,e:=rows.Values()
		if e!=nil{
			log.Error("Error to get row data %v",e)
		}else{
			duration := fmt.Sprint(val[0].(time.Time).Format(time.RFC3339))
			data = append(data,duration)
			for i:=1;i<len(val);i++{
				 data=append(data,fmt.Sprintf("%v", val[i]))
			}
			result = append(result,data)
		}
	}
	return result
}

func(qD postgresQueryData) GenerateQuery() string{
	select_metric := []string{}
	select_metric = append(select_metric, fmt.Sprintf("%s.time qtime", qD.measurement[0]))
	for _, v := range qD.measurement {
		select_metric = append(select_metric, fmt.Sprintf("%s.metric_value %s", v, v))
	}

	from_metric := []string{}
	for _, v := range qD.measurement {
		from_metric = append(from_metric, fmt.Sprintf("\"%s\" %s", v, v))
	}
	
	where_metric := []string{}
	where_metric = append(where_metric, fmt.Sprintf("%s.time between '%s' and '%s' ",
		qD.measurement[0], qD.start, qD.stop))
	for _, v := range qD.measurement {
		where_metric = append(where_metric, fmt.Sprintf("%s.hostname='%s'", v, qD.host_name))
	}
	for i := 1; i < len(qD.measurement); i++ {
		where_metric = append(where_metric, fmt.Sprintf("%s.time=%s.time", qD.measurement[i-1],
		qD.measurement[i]))
	}

	query_order := fmt.Sprintf("%s.time", qD.measurement[0])

	query := fmt.Sprintf(`select %s from %s where %s order by %s asc;`,
		strings.Join(select_metric, ","), strings.Join(from_metric, ","),
		strings.Join(where_metric, " and "), query_order)

		log.Infof("postgresQuery = %v",query)

	return query
}