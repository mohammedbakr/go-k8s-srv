package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/suite"
)

type Rabbitmq struct {
	// conn    *amqp.Connection
	// channel *amqp.Channel
}

type RabbitmqTestSuite struct {
	suite.Suite
	// queue *Rabbitmq
}

// var (
// 	user     = "docker"
// 	password = "secret"
// 	db       = "user"
// port = "5672"

// 	dialect  = "mysql"
// 	dsn      = "%s:%s@tcp(localhost:%s)/%s"
// 	idleConn = 25
// 	maxConn  = 25
// )

func (s *RabbitmqTestSuite) TestprocessmsgMessage() {
	s.T().Run("K8 srv1 massge", func(t *testing.T) {
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
		conn, err := amqp.Dial("amqp://localhost:5672")
		if err != nil {
			log.Fatalf("%s", err)
		}
		defer conn.Close()
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
		Env: []string{"MINIO_ACCESS_KEY=MYACCESSKEY", "MINIO_SECRET_KEY=MYSECRETKEY"},
	}

	resource, err := pool.RunWithOptions(options)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	endpoint := fmt.Sprintf("localhost:%s", resource.GetPort("9000/tcp"))
	// or you could use the following, because we mapped the port 9000 to the port 9000 on the host
	// endpoint := "localhost:9000"

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	// the minio client does not do service discovery for you (i.e. it does not check if connection can be established), so we have to use the health check
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
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4("MYACCESSKEY", "MYSECRETKEY", ""),
		Secure: false,
	})
	if err != nil {
		log.Println("Failed to create minio client:", err)
		return
	}
	log.Printf("%#v\n", minioClient) // minioClient is now set up

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
