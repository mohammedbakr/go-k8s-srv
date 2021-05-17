package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	min7 "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/suite"
)

type RabbitmqTestSuite struct {
	suite.Suite
}

// var body = "body test"

func (s *RabbitmqTestSuite) TestprocessmsgMessage() {
	s.T().Run("K8 srv massge", func(t *testing.T) {
		pool, err := dockertest.NewPool("")
		if err != nil {
			log.Fatalf("Could not connect to docker: %s", err)
		}
		opts := dockertest.RunOptions{
			Repository: "rabbitmq",
			Tag:        "latest",
			Env: []string{
				"host=root",
			},
			ExposedPorts: []string{"5672"},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"5672": {
					{HostPort: "5672"},
				},
			},
		}
		resource, err := pool.RunWithOptions(&opts)
		if err != nil {
			log.Fatalf("Could not start resource: %s", err.Error())
		}
		time.Sleep(20 * time.Second)
		// Get a connection //rabbitmq
		connection, err := amqp.Dial("amqp://localhost:5672")
		if err != nil {
			log.Fatalf("%s", err)
		}
		defer connection.Close()
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	})
}
func TestRabbitmqTestSuite(t *testing.T) {
	suite.Run(t, new(RabbitmqTestSuite))
}

// Minio server
func TestProcessmsgMessage(t *testing.T) {

	minioAccessKey = os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey = os.Getenv("MINIO_SECRET_KEY")

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	options := &dockertest.RunOptions{
		Repository: "minio/minio",
		Tag:        "latest",
		Cmd:        []string{"server", "/data"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"9000/tcp": {{HostPort: "9000"}},
		},
		Env: []string{fmt.Sprintf("MINIO_ACCESS_KEY=%s", minioAccessKey), fmt.Sprintf("MINIO_SECRET_KEY=%s", minioSecretKey)},
	}

	resource, err := pool.RunWithOptions(options)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	endpoint := fmt.Sprintf("localhost:%s", resource.GetPort("9000/tcp"))
	if err := pool.Retry(func() error {
		url := fmt.Sprintf("http://%s/minio/health/live", endpoint)
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf(err.Error())
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status code not OK")
		}
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	time.Sleep(20 * time.Second)
	// now we can instantiate minio client
	minioClient, err := min7.New(endpoint, &min7.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Println("Failed to create minio client:", err)
		return
	}
	// log.Printf("%#v\n", minioClient) // minioClient is now set up

	// now we can use the client, for example, to list the buckets
	buckets, err := minioClient.ListBuckets(context.Background())
	if err != nil {
		log.Fatalf("error while listing buckets: %v", err)
	}
	fmt.Printf("buckets: %+v", buckets)

	// When you're done, kill and remove the container
	if err = pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

// func TestProcessMessage(t *testing.T) {
// 	var err error
// 	// Start a consumer
// 	_, ch, err := rabbitmq.NewQueueConsumer(connection, AdpatationReuquestQueueName, AdpatationReuquestExchange, AdpatationReuquestRoutingKey)
// 	if err != nil {
// 		zlog.Fatal().Err(err).Msg("could not start  AdpatationReuquest consumer ")
// 	}
// 	defer ch.Close()

// 	_, outChannel, err := rabbitmq.NewQueueConsumer(connection, ProcessingOutcomeQueueName, ProcessingOutcomeExchange, ProcessingOutcomeRoutingKey)
// 	if err != nil {
// 		zlog.Fatal().Err(err).Msg("could not start ProcessingOutcome consumer ")

// 	}
// 	defer outChannel.Close()

// 	minioClient, err = minio.NewMinioClient(minioEndpoint, minioAccessKey, minioSecretKey, false)

// 	if err != nil {
// 		zlog.Fatal().Err(err).Msg("could not start minio client ")
// 	}

// 	err = createBucketIfNotExist(sourceMinioBucket)
// 	if err != nil {
// 		zlog.Error().Err(err).Msg(" sourceMinioBucket createBucketIfNotExist error")
// 	}

// 	err = createBucketIfNotExist(cleanMinioBucket)
// 	if err != nil {
// 		zlog.Error().Err(err).Msg("cleanMinioBucket createBucketIfNotExist error")
// 	}

// 	headers := make(amqp.Table)
// 	headers["file-id"] = "544"
// 	headers["source-file-location"] = "../source/file.pdf"
// 	headers["rebuilt-file-location"] = "../rebulid"
// 	var d amqp.Delivery
// 	d.ConsumerTag = "test-tag"
// 	d.Headers = headers
// 	d.ContentType = "text/plain"
// 	d.Body = []byte(body)
// 	t.Run("ProcessMessage", func(t *testing.T) {
// 		ProcessMessage(d)

// 	})
// 	type testSample struct {
// 		data    []byte
// 		headers wabbit.Option
// 		tag     uint64
// 	}
// 	sampleTable := []testSample{
// 		{
// 			data: []byte("teste"),
// 			headers: wabbit.Option{
// 				"contentType": "binary/fuzz",
// 			},
// 			tag: uint64(23473824),
// 		},
// 		{
// 			data: []byte("teste"),
// 			headers: wabbit.Option{
// 				"contentType": "binary/fuzz",
// 			},
// 			tag: uint64(23473824),
// 		},
// 	}

// 	for _, sample := range sampleTable {
// 		t.Run("ProcessMessage", func(t *testing.T) {
// 			if sample.headers["contentType"].(string) != "binary/fuzz" {
// 				t.Errorf("Headers value is nil")
// 			}
// 		})
// 	}
// }
