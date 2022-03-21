package utils

import (
	log "github.com/sirupsen/logrus"
	"github.com/Shopify/sarama"
	"os"
	"os/signal"
	"strings"
	"encoding/json"
	"time"
)

type Tags struct {
	Context_ID    string `json:"context_id"`
	Host          string `json:"host"`
	HostName     string `json:"hostname"`
	MetricLabel  string `json:"metric_label"`
	MetricReport string `json:"metric_report"`
}
type Fields struct {
	MetricValue float64 `json:"metric_value"`
}
type Metric_data struct {
	Name      string `json:"name"`
	Tags      `json:"tags"`
	Fields    `json:"fields"`
	Time int64 `json:"timestamp"`
}

func KafkaConsumer(brokers []string) {

	config := sarama.NewConfig()

	// Create new consumer
	master, err := sarama.NewConsumer(brokers, config)
	checkError("Error to create Consumer", err)


	defer func() {
		if err := master.Close(); err != nil {
			checkError("Error", err)
		}
	}()

	topics, _ := master.Topics()
	//topics :=[]string{"telegraf_SystemUsage"}
	consumer, errors := consume(topics, master)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Count how many message processed
	msgCount := 0

	// Get signnal for finish
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-consumer:
				msgCount++
				log.Debug("Received messages %v", string(msg.Value))

				// Unmarshal or Decode the JSON to the interface.
				var md Metric_data
				json.Unmarshal([]byte(msg.Value), &md)
				log.Debug("Metric data = %v",md)

				Time := time.Unix(md.Time, 0).Format(time.RFC3339)
				
				log.Debug("Time : %v, Name :%v, Metric_Report :%v, Hostname :%v, Value :%v", 
				Time, md.Name, md.MetricReport, md.HostName, md.MetricValue)

			case consumerError := <-errors:
				msgCount++
				log.Error("Received consumerError %v %v %v ", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
				doneCh <- struct{}{}
			case <-signals:
				log.Infof("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh
	log.Infof("Processed %v messages", msgCount)

}

func consume(topics []string, master sarama.Consumer) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError) {
	consumers := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)
	for _, topic := range topics {
		if strings.Contains(topic, "__consumer_offsets") {
			continue
		}
		partitions, _ := master.Partitions(topic)
    // this only consumes partition no 1, you would probably want to consume all partitions
		consumer, err := master.ConsumePartition(topic, partitions[0], sarama.OffsetOldest)
		if nil != err {
			log.Infof("Topic %v Partitions: %v", topic, partitions)
			checkError("Error", err)
		}
		log.Infof(" Start consuming topic %v", topic)
		go func(topic string, consumer sarama.PartitionConsumer) {
			for {
				select {
				case consumerError := <-consumer.Errors():
					errors <- consumerError
					checkError("consumerError ", consumerError.Err)

				case msg := <-consumer.Messages():
					consumers <- msg
					log.Infof("Got message on topic %v %v", topic, string(msg.Value))
				}
			}
		}(topic, consumer)
	}

	return consumers, errors
}
