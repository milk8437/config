package config

import (
	"go.uber.org/zap"
	"github.com/pkg/errors"
	"github.com/bsm/sarama-cluster"
	"github.com/Shopify/sarama"
	"sync"
)

var (
	kafkaConsumers []*cluster.Consumer
	consumersWG    sync.WaitGroup
)

type ConsumerConfig struct {
	Topic    string
	GroupId  string `toml:"group_id"`
	Channels int32
}
type HandleMessage func(string) error

//创建消费者
//消费组提交消息失败会出现重复消费消息的问题，所以业务逻辑要处理重复消费问题
func CreateConsumer(key string, hm HandleMessage) {
	defer consumersWG.Done()
	consumersWG.Add(1)
	config := Prop.Kafka
	consumers := config.Consumers
	cc, ok := consumers[key]
	if !ok {
		panic(errors.Errorf("kafka consumer key:%s not exit!", key))
	}
	log := LOG.Named("KAFKA-CONSUMER-" + cc.Topic)
	log.Info("创建消费者",
		zap.String("group_id", cc.GroupId))

	//sarama不支持consumer-group消费
	clusterConfig := cluster.NewConfig()
	clusterConfig.Config.ClientID = config.ClientId
	clusterConfig.Config.Version = sarama.V1_0_0_0
	topics := []string{cc.Topic}

	consumer, err := cluster.NewConsumer(config.Address, cc.GroupId, topics, clusterConfig)
	if err != nil {
		panic(err)
	}

	//保存消费者
	kafkaConsumers = append(kafkaConsumers, consumer)

	h := &Handle{
		logger:  log,
		workers: make(chan int, cc.Channels),
		job:     hm,
	}

	log.Info("开始消费数据...",
		zap.Int32("channels", cc.Channels))
	for message := range consumer.Messages() {
		if message.Offset%1000 == 0 {
			log.Info("consuming ...",
				zap.ByteString("message", message.Value),
				zap.Int64("offset", message.Offset),
				zap.Int32("partition", message.Partition))
		}
		h.workers <- 1
		go h.run(string(message.Value))
		//提交消费组消费的位置
		//如果提交位置失败，会有重复消费的问题 幂等性
		consumer.MarkOffset(message, "")

	}
	close(h.workers)

	for range h.workers {
		log.Info("等待业务处理kafka消息执行完成...",
			zap.Int("worker", len(h.workers)))
	}
	log.Info("业务处理kafka消息执行完成!")
}

////此项目先关闭消费者，因为生产者数据来源于消费者的信息处理
//func Destroy() {
//	DestroyConsumer()
//	DestroyProducer()
//}

//consume 处理业务 池
type Handle struct {
	logger *zap.Logger
	//工作worker最大数
	workers chan int
	//处理函数
	job func(message string) error
}


func (h *Handle) run(message string ) {
	defer func() {
		<-h.workers
		if p := recover(); p != nil {
			//避免panic异常
			switch  p.(type) {
			case error:
				h.logger.Error("consumer 执行业务逻辑异常",
					zap.String("error", p.(error).Error()),
					zap.Error(p.(error)))
				break
			case string:
				h.logger.Error("consumer 执行业务逻辑异常",
					zap.String("error", p.(string)))
				break
			default:
				h.logger.Error("consumer 执行业务逻辑异常",
					zap.String("error", "unknown error"))
				break
			}
		}
	}()
	if err := h.job(message); err != nil {
		h.logger.Error("consumer 执行业务逻辑失败",
			zap.Error(err))
	}

}


//kafka消费者关闭
func DestroyConsumer() {
	log := LOG.Named("KAFKA-CONSUMER-DOWN")
	for _, kc := range kafkaConsumers {
		kc.Close()
	}
	consumersWG.Wait()
	log.Info("kafka consumer closed!")
}

