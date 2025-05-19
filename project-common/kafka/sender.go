package kafka

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

// LogData 代表一条日志数据，包括主题和具体内容。
// Topic 是日志的主题，用于在Kafka中区分不同类型的日志。
// Data 是日志的具体内容，以JSON格式存储。
type LogData struct {
	Topic string // 日志主题
	//json数据
	Data []byte
}

// KafkaWriter 是一个将日志数据写入Kafka的消息生产者。
// w 是用于写入Kafka的Writer实例。
// data 是一个LogData类型的通道，用于异步地传递日志数据到Kafka。
type KafkaWriter struct {
	w    *kafka.Writer // kafka.Writer实例
	data chan LogData  // 用于接收日志数据的通道。
}

// GetWriter 创建并返回一个KafkaWriter实例。
// 该函数接收一个字符串参数addr，用于指定Kafka的地址。
// 返回值是一个指向KafkaWriter的指针，它包含一个kafka.Writer实例和一个LogData类型的通道。
// 此函数还启动了一个后台goroutine，用于异步发送日志数据到Kafka。
func GetWriter(addr string) *KafkaWriter {
	// 创建一个kafka.Writer实例，配置其地址和负载均衡策略。
	w := &kafka.Writer{
		Addr:     kafka.TCP(addr),     // 使用TCP协议和提供的地址创建Kafka连接。
		Balancer: &kafka.LeastBytes{}, // 选择最少字节的分区作为负载均衡策略。
	}

	// 创建一个KafkaWriter实例，初始化其成员变量。
	k := &KafkaWriter{
		w:    w,                       // 嵌入之前创建的kafka.Writer实例。
		data: make(chan LogData, 100), // 创建一个缓冲大小为100的LogData类型通道，用于传递日志数据。
	}

	// 启动一个goroutine，在其中异步发送日志数据到Kafka。
	go k.sendKafka()

	// 返回KafkaWriter实例的指针。
	return k
}

// sendKafka 是用于向 Kafka 发送消息的核心方法。
// 该方法在一个无限循环中运行，监听通道中的数据，并尝试将数据发送到 Kafka。
// 每次发送最多重试 3 次，如果因为 Leader 不可用或超时导致失败，则会短暂等待后重试。
func (w *KafkaWriter) sendKafka() {
	for {
		select {
		case data := <-w.data: // 从通道中接收要发送的日志数据
			// 构建要发送的 Kafka 消息
			messages := []kafka.Message{
				{
					Topic: data.Topic,       // 消息的主题
					Key:   []byte("logMsg"), // 消息的键，统一设置为 "logMsg"
					Value: data.Data,        // 消息的实际内容
				},
			}
			var err error
			const retries = 3 // 设置最大重试次数为 3 次

			// 创建一个带有超时的上下文，防止发送过程无限阻塞
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel() // 确保在函数退出前释放上下文资源

			for i := 0; i < retries; i++ {
				// 尝试发送消息到 Kafka
				err = w.w.WriteMessages(ctx, messages...)
				if err == nil {
					break // 如果发送成功，跳出重试循环
				}

				// 如果错误是由于 Leader 不可用或者超时，则等待一段时间后重试
				if errors.Is(err, kafka.LeaderNotAvailable) || errors.Is(err, context.DeadlineExceeded) {
					time.Sleep(time.Millisecond * 250) // 等待 250 毫秒后重试
					continue
				}

				// 如果出现其他未知错误，记录日志
				if err != nil {
					log.Printf("kafka send writemessage err %s \n", err.Error())
				}
			}
		}
	}
}

// KafkaWriter的Send方法用于向Kafka发送日志数据。
// 该方法接收一个LogData类型的参数data，通过内部的channel（w.data）将数据发送到Kafka。
// 这种设计允许KafkaWriter以非阻塞的方式异步发送日志数据，提高了性能和响应速度。
func (w *KafkaWriter) Send(data LogData) {
	w.data <- data
}

// KafkaWriter的Close方法用于关闭与Kafka的连接。
// 如果KafkaWriter实例的w字段不为nil，则调用其Close方法来关闭连接。
// 这是为了确保在KafkaWriter对象被销毁时，所有到Kafka的连接都能被正确地关闭，避免资源泄露。
func (w *KafkaWriter) Close() {
	if w.w != nil {
		w.w.Close()
	}
}
