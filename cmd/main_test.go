package main
import (
    "log"
    "testing"
    "time"
    "github.com/ory/dockertest/v3"
    "github.com/ory/dockertest/v3/docker"
    "github.com/streadway/amqp"
    "github.com/stretchr/testify/suite"
)
type Rabbitmq struct {
    conn    *amqp.Connection
    channel *amqp.Channel
}
type RabbitmqTestSuite struct {
    suite.Suite
    queue *Rabbitmq
}
var (
    user     = "docker"
    password = "secret"
    db       = "user"
    port     = "5672"
    dialect  = "mysql"
    dsn      = "%s:%s@tcp(localhost:%s)/%s"
    idleConn = 25
    maxConn  = 25
)
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
                    {HostPort: port},
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