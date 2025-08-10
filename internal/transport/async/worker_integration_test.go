package async_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/testutil"
	"github.com/opcotech/elemo/internal/transport/async"
)

type AsynqWorkerIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	worker *async.Worker
}

func (s *AsynqWorkerIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	var err error

	async.SetRateLimiter(0, 0)
	s.worker, err = async.NewWorker(
		async.WithWorkerConfig(&config.WorkerConfig{
			Concurrency: 10,
			Broker:      s.RedisConf.RedisConfig,
		}),
	)
	s.Require().NoError(err)
}

func (s *AsynqWorkerIntegrationTestSuite) SetupTest() {}

func (s *AsynqWorkerIntegrationTestSuite) TearDownTest() {}

func (s *AsynqWorkerIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *AsynqWorkerIntegrationTestSuite) TestStartShutdown() {
	// start worker in background
	go func() {
		s.Require().NoError(s.worker.Start())
	}()

	// wait for worker to start and crash if needed
	time.Sleep(3 * time.Second)

	// shutdown worker
	s.worker.Shutdown()
}

func TestAsynqWorkerIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AsynqWorkerIntegrationTestSuite))
}
