package stack

import (
	"context"
	"github.com/docker/go-connections/nat"
	docker "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
	"time"
)


func ElasticSearch(version string) *Service {
	return startContainer(
		"docker.elastic.co/elasticsearch/elasticsearch:"+version,
		"9200",
		map[string]string{
			"transport.host": "127.0.0.1",
			"http.port": "9200",
		},
		nil)
}

func Kibana(version string) *Service {
	return startContainer(
		"docker.elastic.co/kibana/kibana:"+version,
		"5601",
		nil,
		nil)
}

func APMServer(version string, opts ...string) *Service {
	opts = append(opts, "-E", "apm-server.host=0.0.0.0:8200")
	cmd := []string{"apm-server", "-e"}
	for _, opt := range opts {
		cmd = append(cmd, "-E", opt)
	}
	return startContainer(
		"docker.elastic.co/apm/apm-server:"+version,
		"8200",
		nil,
		cmd)
}

func APMServerSubCommand(version string, log string, args []string, opts ...string) *Service {
	cmd := []string{"apm-server", "-e"}
	for _, arg := range args {
		cmd = append(cmd, arg)
	}
	for _, opt := range opts {
		cmd = append(cmd, "-E", opt)
	}
	req := docker.ContainerRequest{
		Image: "docker.elastic.co/apm/apm-server:"+version,
		SkipReaper: true,
		WaitingFor:   wait.ForLog(log),
	}
	req.Cmd = cmd
	return launch(req)
}



func startContainer(image, port string, env map[string]string, cmd[]string) *Service {
	req := docker.ContainerRequest{
		Image: image,
		Env: env,
		ExposedPorts: []string{port},
		SkipReaper: true,
		WaitingFor:   wait.ForAll((&wait.HTTPStrategy{
			Port:              nat.Port(port),
			StatusCodeMatcher: func(status int) bool { return status == http.StatusOK },
			Path:              "/",
		}).WithStartupTimeout(time.Second * 10)),
	}
	if cmd != nil {
		req.Cmd = cmd
	}
	return launch(req)
}

func launch(req docker.ContainerRequest) *Service {
	ctx := context.Background()
	container, err := docker.GenericContainer(ctx,
		docker.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
	if err != nil {
		panic(err)
	}

	logConsumer := LogConsumer{}
	if container.StartLogProducer(ctx) != nil {
		panic(err)
	}
	container.FollowOutput(&logConsumer)

	return &Service{
		ctx:         ctx,
		Container:   container,
		logConsumer: logConsumer,
	}
}

type Service struct {
	ctx context.Context
	docker.Container
	logConsumer LogConsumer
}

func (s *Service) Endpoint() string {
	url, err := s.Container.Endpoint(s.ctx, "http")
	if err != nil {
		panic(err)
	}
	return url
}

func (s *Service) Stop() error {
	s.StopLogProducer()
	return s.Terminate(s.ctx)
}

func (s *Service) Logs() []string {
	return s.logConsumer.Messages
}

type LogConsumer struct {
	Messages []string
}

func (c *LogConsumer) Accept(log docker.Log) {
	c.Messages = append(c.Messages, string(log.Content))
}
