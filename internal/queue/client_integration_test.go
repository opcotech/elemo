package queue_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/queue"
	"github.com/opcotech/elemo/internal/testutil"
	"github.com/opcotech/elemo/internal/transport/async"
)

type AsynqClientIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	client *queue.Client
	worker *async.Worker
}

func (s *AsynqClientIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	var err error

	s.client, err = queue.NewClient(
		queue.WithClientConfig(&config.WorkerConfig{
			Concurrency: 10,
			Broker:      s.RedisConf.RedisConfig,
		}),
	)
	s.Require().NoError(err)

	systemHealthCheckTaskHandler, err := async.NewSystemHealthCheckTaskHandler()
	s.Require().NoError(err)

	async.SetRateLimiter(0, 0)
	s.worker, err = async.NewWorker(
		async.WithWorkerConfig(&config.WorkerConfig{
			Concurrency: 10,
			Broker:      s.RedisConf.RedisConfig,
		}),
		async.WithWorkerTaskHandler(queue.TaskTypeSystemHealthCheck, systemHealthCheckTaskHandler),
	)
	s.Require().NoError(err)

	go func() {
		s.Require().NoError(s.worker.Start())
	}()
}

func (s *AsynqClientIntegrationTestSuite) SetupTest() {}

func (s *AsynqClientIntegrationTestSuite) TearDownTest() {}

func (s *AsynqClientIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
	s.worker.Shutdown()
}

func (s *AsynqClientIntegrationTestSuite) TestEnqueue() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	task, err := queue.NewSystemHealthCheckTask()
	s.Require().NoError(err)

	info, err := s.client.Enqueue(ctx, task)
	s.Require().NoError(err)
	s.Assert().Equal(task.Type(), info.Type)
	s.Assert().Equal(task.Payload(), info.Payload)
	s.Assert().Equal(asynq.TaskStatePending, info.State)
}

func (s *AsynqClientIntegrationTestSuite) TestGetTaskInfo() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	task, err := queue.NewSystemHealthCheckTask()
	s.Require().NoError(err)

	info, err := s.client.Enqueue(ctx, task)
	s.Require().NoError(err)

	for info.State != asynq.TaskStateCompleted {
		s.Require().NoError(ctx.Err())
		info, err = s.client.GetTaskInfo(info.Queue, info.ID)
		s.Require().NoError(err)
	}

	s.Assert().Equal(task.Type(), info.Type)
	s.Assert().Equal(task.Payload(), info.Payload)
	s.Assert().Equal(asynq.TaskStateCompleted, info.State)
}

func (s *AsynqClientIntegrationTestSuite) TestPing() {
	s.Require().NoError(s.client.Ping(context.Background()))
}

func (s *AsynqClientIntegrationTestSuite) Test_Z_Close() { // The test suite is run in alphabetical order, so we need to run this test last.
	s.Require().NoError(s.client.Close(context.Background()))
}

func TestAsynqClientIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AsynqClientIntegrationTestSuite))
}
