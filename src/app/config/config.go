package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"gopkg.in/yaml.v2"
	log "github.com/sirupsen/logrus"
)

var configFile Config

//Dataframe Config
type DataframeConfig struct {
	Data	Influxconfig `yaml:"influx"`
	Log         LogConfig      `yaml:"log"`
	File	FileConfig	`yaml:"file"`
	Postgres	Postgresconfig	`yaml:"postgres"`
	Kafka	Kafkaconfig	`yaml:"kafka"`
	Destination		Destinationconfig	`yaml:"destination"`
}

//Influx Config
type Influxconfig struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
	Organization string `yaml:"organization"`
}

//Postgres Config
type Postgresconfig struct {
	PgUser   string `yaml:"pg_user"`
	PgPassword string `yaml:"pg_password"`
	PgHost string `yaml:"pg_host"`
	PgPort string `yaml:"pg_port"`
	PgDb string `yaml:"pg_db"`
}

//Kafka Config
type Kafkaconfig struct {
	Brokers   []string `yaml:"brokers"`
}

//Destination Config
type Destinationconfig struct {
	ActiveDb   string `yaml:"active_db"`
}

//Log Config
type LogConfig struct {
	Location   string `yaml:"location"`
	Level      string `yaml:"level"`
	MaxBackups int    `yaml:"maxbackups"`
	MaxAge     int    `yaml:"maxage"`
}

//File config
type FileConfig struct {
	File_path   string `yaml:"file_path"`
	File_url      string `yaml:"file_url"`
}

//GetConfig get config
func GetConfig() Config {
	return configFile
}

//Config config
type Config struct {
	DataframeConfig DataframeConfig `yaml:"dataframeconfig"`
}

//ReadYamlConfigFile Initial Function
func ReadYamlConfigFile() error {
	var config = DataframeConfig{}
	// dataframe Config
	yamlFile, err := ioutil.ReadFile(getConfigPath())
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return err
	}
	configFile = Config{DataframeConfig: config}
	return nil
}

func getConfigPath() string {
	return GetConfigPath() + "/config/yaml/config.yaml"
}

//GetConfigPath get config path
func GetConfigPath() string {
	ex, err := os.Executable()
	if err != nil {
		log.Error("error to get config %v",err)
	}
	exPath := filepath.Dir(ex)
	return exPath
}