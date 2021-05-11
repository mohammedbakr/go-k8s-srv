package main

import (
	"os"
	"testing"

	"github.com/NeowayLabs/wabbit"
	"github.com/k8-proxy/k8-go-comm/pkg/minio"
	"github.com/k8-proxy/k8-go-comm/pkg/rabbitmq"
	zlog "github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

var (
	uri          = "amqp://guest:guest@localhost:5672/%2f"
	queueName    = "test-queue"
	exchange     = "test-exchange"
	exchangeType = "direct"
	body         = "body test"
	reliable     = true
)

type Delivery struct {
	data          []byte
	headers       wabbit.Option
	tag           uint64
	consumerTag   string
	originalRoute string
	messageId     string
	channel       wabbit.Channel
}

func TestProcessMessage(t *testing.T) {
	var err error

	minioEndpoint = os.Getenv("MINIO_ENDPOINT")
	minioAccessKey = os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey = os.Getenv("MINIO_SECRET_KEY")
	sourceMinioBucket = os.Getenv("MINIO_SOURCE_BUCKET")
	cleanMinioBucket = os.Getenv("MINIO_CLEAN_BUCKET")

	connection, err = rabbitmq.NewInstance(adaptationRequestQueueHostname, adaptationRequestQueuePort, messagebrokeruser, messagebrokerpassword)
	if err != nil {
		zlog.Fatal().Err(err).Msg("could not start rabbitmq connection ")
	}

	// Initiate a publisher on processing exchange

	// Start a consumer
	_, ch, err := rabbitmq.NewQueueConsumer(connection, AdpatationReuquestQueueName, AdpatationReuquestExchange, AdpatationReuquestRoutingKey)
	if err != nil {
		zlog.Fatal().Err(err).Msg("could not start  AdpatationReuquest consumer ")
	}
	defer ch.Close()

	_, outChannel, err := rabbitmq.NewQueueConsumer(connection, ProcessingOutcomeQueueName, ProcessingOutcomeExchange, ProcessingOutcomeRoutingKey)
	if err != nil {
		zlog.Fatal().Err(err).Msg("could not start ProcessingOutcome consumer ")

	}
	defer outChannel.Close()

	minioClient, err = minio.NewMinioClient(minioEndpoint, minioAccessKey, minioSecretKey, false)

	if err != nil {
		zlog.Fatal().Err(err).Msg("could not start minio client ")
	}

	err = createBucketIfNotExist(sourceMinioBucket)
	if err != nil {
		zlog.Error().Err(err).Msg(" sourceMinioBucket createBucketIfNotExist error")
	}

	err = createBucketIfNotExist(cleanMinioBucket)
	if err != nil {
		zlog.Error().Err(err).Msg("cleanMinioBucket createBucketIfNotExist error")
	}

	headers := make(amqp.Table)
	headers["file-id"] = "544"
	headers["source-file-location"] = "../source/file.pdf"
	headers["rebuilt-file-location"] = "../rebulid"
	var d amqp.Delivery
	d.ConsumerTag = "test-tag"
	d.Headers = headers
	d.ContentType = "text/plain"
	d.Body = []byte(body)
	t.Run("ProcessMessage", func(t *testing.T) {
		ProcessMessage(d)

	})
	type testSample struct {
		data          []byte
		headers       wabbit.Option
		tag           uint64
		consumerTag   string
		originalRoute string
		messageId     string
		channel       wabbit.Channel
	}
	sampleTable := []testSample{
		{
			data: []byte("teste"),
			headers: wabbit.Option{
				"contentType": "binary/fuzz",
			},
			tag: uint64(23473824),
		},
		{
			data: []byte("teste"),
			headers: wabbit.Option{
				"contentType": "binary/fuzz",
			},
			tag: uint64(23473824),
		},
	}

	for _, sample := range sampleTable {

		t.Run("ProcessMessage", func(t *testing.T) {

			if sample.headers["contentType"].(string) != "binary/fuzz" {
				t.Errorf("Headers value is nil")

			}

		})
	}
}
