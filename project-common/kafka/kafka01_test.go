package kafka

import (
	"encoding/json"
	"testing"
	"time"
)

func TestProducer01(t *testing.T) {
	w := GetWriter("localhost:9092")
	m := make(map[string]string)
	m["projectCode"] = "1300"
	bytes, _ := json.Marshal(m)
	w.Send(LogData{
		Topic: "msproject_log",
		Data:  bytes,
	})
	time.Sleep(2 * time.Second)
}

func TestConsumer01(t *testing.T) {
	GetReader([]string{"localhost:9092"}, "group1", "msproject_log")
	for {

	}
}
