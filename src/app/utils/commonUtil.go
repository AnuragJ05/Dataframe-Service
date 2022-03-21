package utils

import (
	"context"
	log "github.com/sirupsen/logrus"
	"fmt"
	"dataframe-service/src/app/config"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/jackc/pgx/v4/pgxpool"
)

func GetIdracs() map[string][]string{

	idracs_list := make(map[string][]string)

	if config.GetConfig().DataframeConfig.Destination.ActiveDb == "influxdb"{
		url := config.GetConfig().DataframeConfig.Data.URL
		token := config.GetConfig().DataframeConfig.Data.Token
		organization := config.GetConfig().DataframeConfig.Data.Organization
		
		// Create a new client using an InfluxDB server base URL and an authentication token
		client := influxdb2.NewClient(url, token)

		// Ensures background processes finishes
		defer client.Close()

		// Get query client
		queryAPI := client.QueryAPI(organization)

		// Query and get complete result as a string
		fluxQuery := fmt.Sprintf(`from(bucket: "telegraf")
		|> range(start: -15m)
		|> keep(columns: ["hostname"])
		|> distinct(column: "hostname")`)

		// Use default dialect
		result, err := queryAPI.Query(context.Background(), fluxQuery)
		checkError("Cannot get the data", err)

		// Iterate over query response
		for result.Next() {
			value := fmt.Sprint(result.Record().ValueByKey("_value"))
			if value != "<nil>"{
				idracs_list["idrac"] = append(idracs_list["idrac"], value)
			}
		}

	} else if config.GetConfig().DataframeConfig.Destination.ActiveDb == "postgres"{

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

		//create index
		create_index_query := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS index ON 
		"dell_poweredge_AggregateUsage"(time, hostname);
		`)
		index, e1 := dbpool.Query(ctx, create_index_query)
		checkError("Error to create index", e1)

		defer index.Close()
		log.Infof("Successfully executed query")

		get_host_query := fmt.Sprintf(`select distinct(hostname) from "dell_poweredge_AggregateUsage";
		`)
		log.Infof("query= %v",get_host_query) 

		//Execute query on TimescaleDB
		rows, e2 := dbpool.Query(ctx, get_host_query)
		checkError("Error to query data", e2)

		defer rows.Close()
		log.Infof("Successfully executed query")

		for rows.Next() {
			val,e:=rows.Values()
			if e!=nil{
				log.Error("Error to get row data %v",e)
			}else{
				idracs_list["idrac"] = append(idracs_list["idrac"], val[0].(string))
			}
		}
	}
	log.Infof("Idracs List = %v",idracs_list)
	return idracs_list
}


func checkError(message string, err error) {
    if err != nil {
        log.Error("%v : %v",message, err)
    }
}