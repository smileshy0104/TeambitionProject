package kafka

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
)

// KafkaReader 是一个用于从Kafka主题读取消息的结构体。
// 它包含一个*kafka.Reader类型的成员R，用于实际的消息读取操作。
type KafkaReader struct {
	R *kafka.Reader
}

// readMsg 是KafkaReader的一个方法，用于不断读取消息。
// 该方法进入一个无限循环，尝试从Kafka中读取消息，并打印出来。
// 如果在读取消息过程中遇到错误，它会记录错误并继续尝试读取下一消息。
func (r *KafkaReader) readMsg() {
	for {
		m, err := r.R.ReadMessage(context.Background())
		if err != nil {
			log.Printf("kafka readMsg err %s \n", err.Error())
			continue
		}
		fmt.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
	}
}

// Close 关闭 Kafka 读取器
// 该方法主要负责关闭与 Kafka 的连接，释放相关资源
// 没有参数和返回值
func (r *KafkaReader) Close() {
	r.R.Close()
}

// GetReader 创建并返回一个KafkaReader实例，用于消费Kafka主题的消息。
// 该函数需要三个参数：
// 1. brokers: 一个包含Kafka代理地址的切片。
// 2. groupId: 消费者组的标识符，同一个组下的消费者会协同工作，共同消费主题队列中的内容。
// 3. topic: 消费者要订阅的Kafka主题。
// 返回值是一个指向KafkaReader的指针，通过它可以读取消息。
func GetReader(brokers []string, groupId, topic string) *KafkaReader {
	// 创建一个kafka.NewReader实例，配置包括：
	// 1. Brokers: 指定Kafka代理的地址列表。
	// 2. GroupID: 设置消费者组的ID，用于Kafka消费者组管理。
	// 3. Topic: 指定消费者要订阅的主题。
	// 4. MinBytes: 设置读取消息时的最小字节数，避免频繁读取小块数据。
	// 5. MaxBytes: 设置读取消息时的最大字节数，防止一次性读取过多数据导致内存压力。
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupId, //同一个组下的consumer 协同工作 共同消费topic队列中的内容
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	// 创建一个KafkaReader实例，将上述配置的kafka.NewReader实例赋给其R字段。
	k := &KafkaReader{R: r}

	// 启动一个goroutine异步读取消息，保持消费者活跃。
	go k.readMsg()

	// 返回KafkaReader实例的指针。
	return k
}
