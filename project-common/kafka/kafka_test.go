package kafka

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"testing"
	"time"
)

// TestProducer 测试生产者功能，用于向Kafka主题发送消息
func TestProducer(t *testing.T) {
	// 测试生产者功能，用于向Kafka主题发送消息

	// 定义要发送消息的主题和分区
	topic := "my-topic" // 指定Kafka主题名称
	partition := 0      // 指定分区编号

	// 连接到指定Kafka主题分区的Leader Broker
	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err) // 如果连接失败则记录错误并终止程序
	}

	// 设置写入操作的截止时间，防止无限期阻塞
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// 发送多条消息到Kafka
	_, err = conn.WriteMessages(
		kafka.Message{Value: []byte("one!")},   // 第一条消息内容为 "one!"
		kafka.Message{Value: []byte("two!")},   // 第二条消息内容为 "two!"
		kafka.Message{Value: []byte("three!")}, // 第三条消息内容为 "three!"
	)
	if err != nil {
		log.Fatal("failed to write messages:", err) // 如果写入消息失败则记录错误并终止程序
	}

	// 关闭连接以释放资源
	if err := conn.Close(); err != nil {
		log.Fatal("failed to close writer:", err) // 如果关闭连接失败则记录错误并终止程序
	}
}

// TestConsumer 测试消费者功能，用于从Kafka主题中读取消息
func TestConsumer(t *testing.T) {
	// 消费者测试函数，用于从Kafka主题中读取消息

	// 定义要消费消息的主题和分区
	topic := "my-topic" // Kafka主题名称
	partition := 0      // 分区编号

	// 连接到指定Kafka主题分区的Leader Broker
	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err) // 如果连接失败则记录错误并终止程序
	}

	// 设置读取操作的截止时间，防止无限期阻塞
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// 开始读取一批消息，最小读取10KB，最大1MB
	batch := conn.ReadBatch(10e3, 1e6) // fetch 10KB min, 1MB max

	// 创建一个缓冲区，用于存储每次读取的消息（每个消息最多10KB）
	b := make([]byte, 10e3) // 10KB max per message

	// 循环读取批次中的所有消息
	for {
		n, err := batch.Read(b) // 将消息读入缓冲区
		if err != nil {
			break // 如果读取完成或发生错误，则退出循环
		}
		// 打印读取到的消息内容
		fmt.Println(string(b[:n]))
	}

	// 关闭批次以释放相关资源
	if err := batch.Close(); err != nil {
		log.Fatal("failed to close batch:", err) // 如果关闭失败则记录错误并终止程序
	}

	// 关闭连接以释放资源
	if err := conn.Close(); err != nil {
		log.Fatal("failed to close connection:", err) // 如果关闭失败则记录错误并终止程序
	}
}
