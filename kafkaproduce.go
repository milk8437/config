package config

import (
	"github.com/Shopify/sarama"
	"sync"
	"go.uber.org/zap"
)

type KafkaConfig struct {
	Address   []string
	ClientId  string `toml:"client_id"`
	Topics    map[string]string
	Consumers map[string]*ConsumerConfig
}

var (
	AsyncProducer sarama.AsyncProducer
	producerWG    sync.WaitGroup
)
//构建kafka客户端
func CreateProducer() {
	kc := Prop.Kafka
	var err error

	//构建producer
	//默认协程数256
	kafkaConfig := sarama.NewConfig()
	//生产消息成功回调,默认为false
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.ClientID = kc.ClientId
	kafkaConfig.Version = sarama.V1_0_0_0
	kafkaConfig.Producer.Flush.MaxMessages = 100
	//sarama.MaxRequestSize = 10 * 1024 * 1024
	AsyncProducer, err = sarama.NewAsyncProducer(kc.Address, kafkaConfig)
	if err != nil {
		panic(err)
	}
	//创建producer成功后，监控生产消息的返回值
	producerWG.Add(1)
	go returnError()
	producerWG.Add(1)
	go returnSuccess()
}

//获取生产数据提交失败的
func returnError() {
	defer producerWG.Done()
	log := LOG.Named("PRODUCER")
	//获取生产数据错误的
	for err := range AsyncProducer.Errors() {
		log.Error("produce failed!!!",
			zap.String("topic", err.Msg.Topic),
			zap.String("error", err.Error()))
	}
	log.Info("kafka return error channel closed!")
}

//获取生产数据的提交结果 ，成功的
func returnSuccess() {
	defer producerWG.Done()
	log := LOG.Named("PRODUCER")
	for success := range AsyncProducer.Successes() {
		if success.Offset%1000 == 0 {
			log.Info("produce success ...",
				zap.Int64("offset", success.Offset),
				zap.String("topic", success.Topic))
		}
	}
	log.Info("kafka return success channel closed!")
}

//关闭生产者
//AsyncClose 此方法会等 errors和success都处理完再结束
func DestroyProducer() {
	AsyncProducer.AsyncClose()
	producerWG.Wait()
	log := LOG.Named("PRODUCER")
	log.Info("kafka producer closed !")
}
