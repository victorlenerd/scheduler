package node

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	"github.com/robfig/cron"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"scheduler0/pkg/config"
	"scheduler0/pkg/db"
	"scheduler0/pkg/fsm"
	"scheduler0/pkg/models"
	async_task_repo "scheduler0/pkg/repository/async_task"
	job_repo "scheduler0/pkg/repository/job"
	job_execution_repo "scheduler0/pkg/repository/job_execution"
	job_queue_repo "scheduler0/pkg/repository/job_queue"
	project_repo "scheduler0/pkg/repository/project"
	"scheduler0/pkg/scheduler0time"
	"scheduler0/pkg/secrets"
	"scheduler0/pkg/service/async_task"
	"scheduler0/pkg/service/executor"
	"scheduler0/pkg/service/executor/executors"
	"scheduler0/pkg/service/job"
	"scheduler0/pkg/service/processor"
	"scheduler0/pkg/service/queue"
	"scheduler0/pkg/shared_repo"
	"scheduler0/pkg/utils"
	"testing"
	"time"
)

//func Test_CanAcceptRequest(t *testing.T) {
//	t.Cleanup(func() {
//		err := os.RemoveAll("./raft_data")
//		if err != nil {
//			fmt.Println("failed to remove raft_data dir for test", err)
//		}
//	})
//
//	logger := hclog.New(&hclog.LoggerOptions{
//		Name:  "node-service-test",
//		Level: hclog.LevelFromString("trace"),
//	})
//
//	// Create a temporary SQLite database file
//	tempFile, err := ioutil.TempFile("", "test-db")
//	if err != nil {
//		t.Fatalf("Failed to create temp file: %v", err)
//	}
//	defer os.Remove(tempFile.Name())
//
//	// Create a new SQLite database connection
//	sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
//	sqliteDb.RunMigration()
//	sqliteDb.OpenConnectionToExistingDB()
//
//	scheduler0config := config.NewScheduler0Config()
//	sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
//	scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)
//
//	scheduler0Secrets := secrets.NewScheduler0Secrets()
//	// Create a new FSM store
//	scheduler0Store := fsm.NewFSMStore(logger, scheduler0RaftActions, sqliteDb, nil, nil, nil, nil, sharedRepo)
//
//	// Create a mock raft cluster
//	cluster := raft.MakeClusterCustom(t, &raft.MakeClusterOpts{
//		Peers:          1,
//		Bootstrap:      true,
//		Conf:           raft.DefaultConfig(),
//		ConfigStoreFSM: false,
//		MakeFSMFunc: func() raft.FSM {
//			return scheduler0Store.GetFSM()
//		},
//	})
//	defer cluster.Close()
//	cluster.FullyConnect()
//	scheduler0Store.UpdateRaft(cluster.Leader())
//
//	ctx := context.Background()
//
//	jobRepo := job_repo.NewJobRepo(logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskManagerRepo := async_task_repo.NewAsyncTasksRepo(ctx, logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskService := async_task.NewAsyncTaskManager(ctx, logger, scheduler0Store, asyncTaskManagerRepo, scheduler0config)
//	jobQueueRepo := job_queue_repo.NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)
//	jobExecutionsRepo := job_execution_repo.NewExecutionsRepo(
//		logger,
//		scheduler0RaftActions,
//		scheduler0Store,
//	)
//
//	dispatcher := utils.NewDispatcher(
//		ctx,
//		int64(1),
//		int64(1),
//	)
//
//	dispatcher.Run()
//
//	queueService := queue.NewJobQueue(ctx, logger, scheduler0config, scheduler0RaftActions, scheduler0Store, jobQueueRepo)
//	httpJobExecutor := executors.NewMockHTTPExecutor(t)
//
//	jobExecutorService := executor.NewJobExecutor(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0RaftActions,
//		jobRepo,
//		jobExecutionsRepo,
//		jobQueueRepo,
//		httpJobExecutor,
//		dispatcher,
//	)
//
//	nodeHTTPClient := NewHTTPClient(logger, scheduler0config, scheduler0Secrets)
//	jobProcessor := processor.NewMockJobProcessorService(t)
//
//	nodeService := NewNode(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0Secrets,
//		scheduler0Store,
//		scheduler0RaftActions,
//		jobExecutorService,
//		queueService,
//		jobProcessor,
//		jobRepo,
//		sharedRepo,
//		jobExecutionsRepo,
//		asyncTaskService,
//		dispatcher,
//		nodeHTTPClient,
//		nil,
//	)
//
//	canAcceptRequest := nodeService.CanAcceptRequest()
//	assert.Equal(t, false, canAcceptRequest)
//	nodeService.beginAcceptingClientRequest()
//	canAcceptRequest = nodeService.CanAcceptRequest()
//	assert.Equal(t, true, canAcceptRequest)
//}
//
//func Test_CanAcceptClientWriteRequest(t *testing.T) {
//	t.Cleanup(func() {
//		err := os.RemoveAll("./raft_data")
//		if err != nil {
//			fmt.Println("failed to remove raft_data dir for test", err)
//		}
//	})
//
//	logger := hclog.New(&hclog.LoggerOptions{
//		Name:  "node-service-test",
//		Level: hclog.LevelFromString("trace"),
//	})
//
//	// Create a temporary SQLite database file
//	tempFile, err := ioutil.TempFile("", "test-db")
//	if err != nil {
//		t.Fatalf("Failed to create temp file: %v", err)
//	}
//	defer os.Remove(tempFile.Name())
//
//	// Create a new SQLite database connection
//	sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
//	sqliteDb.RunMigration()
//	sqliteDb.OpenConnectionToExistingDB()
//
//	scheduler0config := config.NewScheduler0Config()
//	sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
//	scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)
//
//	scheduler0Secrets := secrets.NewScheduler0Secrets()
//	// Create a new FSM store
//	scheduler0Store := fsm.NewFSMStore(logger, scheduler0RaftActions, sqliteDb, nil, nil, nil, nil, sharedRepo)
//
//	// Create a mock raft cluster
//	cluster := raft.MakeClusterCustom(t, &raft.MakeClusterOpts{
//		Peers:          1,
//		Bootstrap:      true,
//		Conf:           raft.DefaultConfig(),
//		ConfigStoreFSM: false,
//		MakeFSMFunc: func() raft.FSM {
//			return scheduler0Store.GetFSM()
//		},
//	})
//	defer cluster.Close()
//	cluster.FullyConnect()
//	scheduler0Store.UpdateRaft(cluster.Leader())
//
//	ctx := context.Background()
//
//	jobRepo := job_repo.NewJobRepo(logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskManagerRepo := async_task_repo.NewAsyncTasksRepo(ctx, logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskService := async_task.NewAsyncTaskManager(ctx, logger, scheduler0Store, asyncTaskManagerRepo, scheduler0config)
//	jobQueueRepo := job_queue_repo.NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)
//	jobExecutionsRepo := job_execution_repo.NewExecutionsRepo(
//		logger,
//		scheduler0RaftActions,
//		scheduler0Store,
//	)
//
//	dispatcher := utils.NewDispatcher(
//		ctx,
//		int64(1),
//		int64(1),
//	)
//
//	dispatcher.Run()
//
//	queueService := queue.NewJobQueue(ctx, logger, scheduler0config, scheduler0RaftActions, scheduler0Store, jobQueueRepo)
//	httpJobExecutor := executors.NewMockHTTPExecutor(t)
//
//	jobExecutorService := executor.NewJobExecutor(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0RaftActions,
//		jobRepo,
//		jobExecutionsRepo,
//		jobQueueRepo,
//		httpJobExecutor,
//		dispatcher,
//	)
//
//	nodeHTTPClient := NewHTTPClient(logger, scheduler0config, scheduler0Secrets)
//	jobProcessor := processor.NewMockJobProcessorService(t)
//
//	nodeService := NewNode(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0Secrets,
//		scheduler0Store,
//		scheduler0RaftActions,
//		jobExecutorService,
//		queueService,
//		jobProcessor,
//		jobRepo,
//		sharedRepo,
//		jobExecutionsRepo,
//		asyncTaskService,
//		dispatcher,
//		nodeHTTPClient,
//		nil,
//	)
//
//	canAcceptClientWriteRequest := nodeService.CanAcceptClientWriteRequest()
//	assert.Equal(t, false, canAcceptClientWriteRequest)
//	nodeService.beginAcceptingClientWriteRequest()
//	canAcceptClientWriteRequest = nodeService.CanAcceptClientWriteRequest()
//	assert.Equal(t, true, canAcceptClientWriteRequest)
//	nodeService.stopAcceptingClientWriteRequest()
//	canAcceptClientWriteRequest = nodeService.CanAcceptClientWriteRequest()
//	assert.Equal(t, false, canAcceptClientWriteRequest)
//}

func Test_ReturnUncommittedLogs(t *testing.T) {
	t.Cleanup(func() {
		err := os.RemoveAll("./raft_data")
		if err != nil {
			fmt.Println("failed to remove raft_data dir for test", err)
		}
	})

	bctx := context.Background()
	ctx, cancler := context.WithCancel(bctx)
	defer cancler()
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "node-service-test",
		Level: hclog.LevelFromString("trace"),
	})

	// Create a temporary SQLite database file
	tempFile, err := ioutil.TempFile("", "test-db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Create a new SQLite database connection
	sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
	sqliteDb.RunMigration()
	sqliteDb.OpenConnectionToExistingDB()

	scheduler0config := config.NewScheduler0Config()
	sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
	scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)

	scheduler0Secrets := secrets.NewScheduler0Secrets()
	// Create a new FSM store
	scheduler0Store := fsm.NewFSMStore(logger, scheduler0RaftActions, scheduler0config, sqliteDb, nil, nil, nil, nil, sharedRepo)

	// Create a mock raft cluster
	cluster := raft.MakeClusterCustom(t, &raft.MakeClusterOpts{
		Peers:          1,
		Bootstrap:      true,
		Conf:           raft.DefaultConfig(),
		ConfigStoreFSM: false,
		MakeFSMFunc: func() raft.FSM {
			return scheduler0Store.GetFSM()
		},
	})
	cluster.FullyConnect()
	scheduler0Store.UpdateRaft(cluster.Leader())

	jobRepo := job_repo.NewJobRepo(logger, scheduler0RaftActions, scheduler0Store)
	projectRepo := project_repo.NewProjectRepo(logger, scheduler0RaftActions, scheduler0Store, jobRepo)
	asyncTaskManagerRepo := async_task_repo.NewAsyncTasksRepo(ctx, logger, scheduler0RaftActions, scheduler0Store)
	asyncTaskService := async_task.NewAsyncTaskManager(ctx, logger, scheduler0Store, asyncTaskManagerRepo, scheduler0config)
	jobQueueRepo := job_queue_repo.NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)
	jobExecutionsRepo := job_execution_repo.NewExecutionsRepo(
		logger,
		scheduler0RaftActions,
		scheduler0Store,
	)

	dispatcher := utils.NewDispatcher(
		ctx,
		int64(1),
		int64(1),
	)

	dispatcher.Run()

	queueService := queue.NewJobQueue(ctx, logger, scheduler0config, scheduler0RaftActions, scheduler0Store, jobQueueRepo)
	httpJobExecutor := executors.NewMockHTTPExecutor(t)

	jobExecutorService := executor.NewJobExecutor(
		ctx,
		logger,
		scheduler0config,
		scheduler0RaftActions,
		jobRepo,
		jobExecutionsRepo,
		jobQueueRepo,
		httpJobExecutor,
		dispatcher,
	)
	jobService := job.NewJobService(ctx, logger, jobRepo, queueService, projectRepo, dispatcher, asyncTaskService)

	nodeHTTPClient := NewHTTPClient(logger, scheduler0config, scheduler0Secrets)
	jobProcessor := processor.NewMockJobProcessorService(t)

	nodeService := NewNode(
		ctx,
		logger,
		scheduler0config,
		scheduler0Secrets,
		scheduler0Store,
		scheduler0RaftActions,
		jobExecutorService,
		queueService,
		jobProcessor,
		jobRepo,
		sharedRepo,
		jobExecutionsRepo,
		asyncTaskService,
		dispatcher,
		nodeHTTPClient,
		nil,
		false,
	)

	requestId := "some-request-id"
	asyncTaskService.ListenForNotifications()
	// Define the input jobs
	jobs := []models.Job{}

	i := 1
	for i < 100 {
		jobs = append(jobs, models.Job{
			ID:        uint64(i),
			Spec:      "@every 1h",
			Timezone:  "America/New_York",
			ProjectID: 1,
		})
		i++
	}

	// Create the projects using the project repo
	project := models.Project{
		ID:          1,
		Name:        fmt.Sprintf("Project %d", 1),
		Description: fmt.Sprintf("Project %d description", 1),
	}
	_, createErr := projectRepo.CreateOne(&project)
	if createErr != nil {
		t.Fatalf("Failed to create project: %v", createErr)
	}

	// Call the BatchInsertJobs method of the job service
	_, batchErr := jobService.BatchInsertJobs("request123", jobs)
	if batchErr != nil {
		t.Fatalf("Failed to insert jobs: %v", batchErr)
	}

	time.Sleep(time.Second * time.Duration(2))

	schedule, parseErr := cron.Parse("@every 1h")
	if parseErr != nil {
		t.Fatal("cron spec error", parseErr)
	}

	schedulerTime := scheduler0time.GetSchedulerTime()
	now := schedulerTime.GetTime(time.Now())
	nextTime := schedule.Next(now)
	prevNextTime := nextTime.Add(-nextTime.Sub(now))
	lastTime := nextTime.Add(-nextTime.Sub(now)).Add(-nextTime.Sub(now))

	serr := os.Setenv("SCHEDULER0_NODE_ID", "1")
	defer os.Unsetenv("SCHEDULER0_NODE_ID")
	if serr != nil {
		t.Fatal("failed to set env", serr)
	}
	uncommittedExecutionsLogs := []models.JobExecutionLog{}
	i = 0
	for i < 24 {
		uncommittedExecutionsLogs = append(uncommittedExecutionsLogs, models.JobExecutionLog{
			JobId:                 jobs[i].ID,
			UniqueId:              fmt.Sprintf("%d-%d", jobs[i].ID, i),
			State:                 models.ExecutionLogSuccessState,
			LastExecutionDatetime: lastTime,
			NextExecutionDatetime: prevNextTime,
			NodeId:                1,
		})
		i++
	}
	err = sharedRepo.InsertExecutionLogs(sqliteDb, false, uncommittedExecutionsLogs)
	if err != nil {
		t.Fatal("failed to insert execution logs", err)
	}

	nodeService.GetUncommittedLogs(requestId)

	time.Sleep(time.Second * time.Duration(2))

	task, getTaskErr := asyncTaskService.GetTaskWithRequestIdNonBlocking(requestId)
	if getTaskErr != nil {
		t.Fatal("failed update an async task", getTaskErr)
	}
	data := []byte(task.Output)
	var asyncTaskRes models.LocalData
	marshalErr := json.Unmarshal(data, &asyncTaskRes)

	if marshalErr != nil {
		t.Fatal("failed to unmarshal", marshalErr)
	}

	assert.Equal(t, task.RequestId, requestId)
	assert.Equal(t, task.State, models.AsyncTaskSuccess)
	assert.Equal(t, len(asyncTaskRes.ExecutionLogs), len(uncommittedExecutionsLogs))
}

//func Test_getRaftConfiguration(t *testing.T) {
//	t.Cleanup(func() {
//		err := os.RemoveAll("./raft_data")
//		if err != nil {
//			fmt.Println("failed to remove raft_data dir for test", err)
//		}
//	})
//
//	ctx := context.Background()
//	logger := hclog.New(&hclog.LoggerOptions{
//		Name:  "node-service-test",
//		Level: hclog.LevelFromString("trace"),
//	})
//
//	// Create a temporary SQLite database file
//	tempFile, err := ioutil.TempFile("", "test-db")
//	if err != nil {
//		t.Fatalf("Failed to create temp file: %v", err)
//	}
//	defer os.Remove(tempFile.Name())
//
//	// Create a new SQLite database connection
//	sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
//	sqliteDb.RunMigration()
//	sqliteDb.OpenConnectionToExistingDB()
//
//	scheduler0config := config.NewScheduler0Config()
//	sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
//	scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)
//
//	scheduler0Secrets := secrets.NewScheduler0Secrets()
//	// Create a new FSM store
//	scheduler0Store := fsm.NewFSMStore(logger, scheduler0RaftActions, sqliteDb, nil, nil, nil, nil, sharedRepo)
//
//	// Create a mock raft cluster
//	cluster := raft.MakeClusterCustom(t, &raft.MakeClusterOpts{
//		Peers:          1,
//		Bootstrap:      true,
//		Conf:           raft.DefaultConfig(),
//		ConfigStoreFSM: false,
//		MakeFSMFunc: func() raft.FSM {
//			return scheduler0Store.GetFSM()
//		},
//	})
//	defer cluster.Close()
//	cluster.FullyConnect()
//	scheduler0Store.UpdateRaft(cluster.Leader())
//
//	jobRepo := job_repo.NewJobRepo(logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskManagerRepo := async_task_repo.NewAsyncTasksRepo(ctx, logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskService := async_task.NewAsyncTaskManager(ctx, logger, scheduler0Store, asyncTaskManagerRepo, scheduler0config)
//	jobQueueRepo := job_queue_repo.NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)
//	jobExecutionsRepo := job_execution_repo.NewExecutionsRepo(
//		logger,
//		scheduler0RaftActions,
//		scheduler0Store,
//	)
//
//	dispatcher := utils.NewDispatcher(
//		ctx,
//		int64(1),
//		int64(1),
//	)
//
//	dispatcher.Run()
//
//	queueService := queue.NewJobQueue(ctx, logger, scheduler0config, scheduler0RaftActions, scheduler0Store, jobQueueRepo)
//	httpJobExecutor := executors.NewMockHTTPExecutor(t)
//
//	jobExecutorService := executor.NewJobExecutor(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0RaftActions,
//		jobRepo,
//		jobExecutionsRepo,
//		jobQueueRepo,
//		httpJobExecutor,
//		dispatcher,
//	)
//	nodeHTTPClient := NewHTTPClient(logger, scheduler0config, scheduler0Secrets)
//	jobProcessor := processor.NewMockJobProcessorService(t)
//
//	nodeService := NewNode(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0Secrets,
//		scheduler0Store,
//		scheduler0RaftActions,
//		jobExecutorService,
//		queueService,
//		jobProcessor,
//		jobRepo,
//		sharedRepo,
//		jobExecutionsRepo,
//		asyncTaskService,
//		dispatcher,
//		nodeHTTPClient,
//		nil,
//	)
//
//	os.Setenv("SCHEDULER0_RAFT_ADDRESS", "http://localhost:3440")
//	defer os.Unsetenv("SCHEDULER0_RAFT_ADDRESS")
//
//	os.Setenv("SCHEDULER0_NODE_ID", "2")
//	defer os.Unsetenv("SCHEDULER0_NODE_ID")
//
//	os.Setenv("SCHEDULER0_REPLICAS", "[{\"raft_address\":\"http://localhost:3441\", \"nodeId\":1}]")
//	defer os.Unsetenv("SCHEDULER0_REPLICAS")
//
//	configurations := nodeService.getRaftConfiguration()
//	fmt.Println("TODO::configurations", configurations)
//	assert.Equal(t, 2, len(configurations.Servers))
//}
//
//func Test_getRandomFanInPeerHTTPAddresses(t *testing.T) {
//	t.Cleanup(func() {
//		err := os.RemoveAll("./raft_data")
//		if err != nil {
//			fmt.Println("failed to remove raft_data dir for test", err)
//		}
//	})
//
//	ctx := context.Background()
//	logger := hclog.New(&hclog.LoggerOptions{
//		Name:  "node-service-test",
//		Level: hclog.LevelFromString("trace"),
//	})
//
//	// Create a temporary SQLite database file
//	tempFile, err := ioutil.TempFile("", "test-db")
//	if err != nil {
//		t.Fatalf("Failed to create temp file: %v", err)
//	}
//	defer os.Remove(tempFile.Name())
//
//	// Create a new SQLite database connection
//	sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
//	sqliteDb.RunMigration()
//	sqliteDb.OpenConnectionToExistingDB()
//
//	scheduler0config := config.NewScheduler0Config()
//	sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
//	scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)
//
//	scheduler0Secrets := secrets.NewScheduler0Secrets()
//	// Create a new FSM store
//	scheduler0Store := fsm.NewFSMStore(logger, scheduler0RaftActions, sqliteDb, nil, nil, nil, nil, sharedRepo)
//
//	// Create a mock raft cluster
//	cluster := raft.MakeClusterCustom(t, &raft.MakeClusterOpts{
//		Peers:          5,
//		Bootstrap:      true,
//		Conf:           raft.DefaultConfig(),
//		ConfigStoreFSM: false,
//		MakeFSMFunc: func() raft.FSM {
//			return scheduler0Store.GetFSM()
//		},
//	})
//	defer cluster.Close()
//	cluster.FullyConnect()
//	scheduler0Store.UpdateRaft(cluster.Leader())
//
//	jobRepo := job_repo.NewJobRepo(logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskManagerRepo := async_task_repo.NewAsyncTasksRepo(ctx, logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskService := async_task.NewAsyncTaskManager(ctx, logger, scheduler0Store, asyncTaskManagerRepo, scheduler0config)
//	jobQueueRepo := job_queue_repo.NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)
//	jobExecutionsRepo := job_execution_repo.NewExecutionsRepo(
//		logger,
//		scheduler0RaftActions,
//		scheduler0Store,
//	)
//
//	dispatcher := utils.NewDispatcher(
//		ctx,
//		int64(1),
//		int64(1),
//	)
//
//	dispatcher.Run()
//
//	queueService := queue.NewJobQueue(ctx, logger, scheduler0config, scheduler0RaftActions, scheduler0Store, jobQueueRepo)
//	httpJobExecutor := executors.NewMockHTTPExecutor(t)
//
//	jobExecutorService := executor.NewJobExecutor(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0RaftActions,
//		jobRepo,
//		jobExecutionsRepo,
//		jobQueueRepo,
//		httpJobExecutor,
//		dispatcher,
//	)
//	nodeHTTPClient := NewHTTPClient(logger, scheduler0config, scheduler0Secrets)
//	jobProcessor := processor.NewMockJobProcessorService(t)
//
//	nodeService := NewNode(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0Secrets,
//		scheduler0Store,
//		scheduler0RaftActions,
//		jobExecutorService,
//		queueService,
//		jobProcessor,
//		jobRepo,
//		sharedRepo,
//		jobExecutionsRepo,
//		asyncTaskService,
//		dispatcher,
//		nodeHTTPClient,
//		nil,
//	)
//
//	os.Setenv("SCHEDULER0_RAFT_ADDRESS", strings.Split(cluster.Leader().String(), " ")[2])
//	defer os.Unsetenv("SCHEDULER0_RAFT_ADDRESS")
//
//	os.Setenv("SCHEDULER0_EXECUTION_LOG_FETCH_FAN_IN", "2")
//	defer os.Unsetenv("SCHEDULER0_EXECUTION_LOG_FETCH_FAN_IN")
//
//	followerHTTPAddresses := []string{
//		"http://localhost:34417",
//		"http://localhost:34416",
//		"http://localhost:34415",
//		"http://localhost:34414",
//	}
//
//	followerRaftAddresses := []string{
//		strings.Split(cluster.Followers()[0].String(), " ")[2],
//		strings.Split(cluster.Followers()[1].String(), " ")[2],
//		strings.Split(cluster.Followers()[2].String(), " ")[2],
//		strings.Split(cluster.Followers()[3].String(), " ")[2],
//	}
//
//	os.Setenv("SCHEDULER0_REPLICAS", fmt.Sprintf("["+
//		"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":5}, "+
//		"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":4}, "+
//		"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":3}, "+
//		"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":2}, "+
//		"{\"raft_address\":\"http://localhost:3442\", \"address\":\"http://localhost:34413\", \"nodeId\":1}"+
//		"]",
//		followerRaftAddresses[0], followerHTTPAddresses[0],
//		followerRaftAddresses[1], followerHTTPAddresses[1],
//		followerRaftAddresses[2], followerHTTPAddresses[2],
//		followerRaftAddresses[3], followerHTTPAddresses[3],
//	))
//	defer os.Unsetenv("SCHEDULER0_REPLICAS")
//
//	nodeService.FsmStore = scheduler0Store
//
//	tests := []struct {
//		Name        string
//		ExcludeList map[string]bool
//	}{{
//		Name:        "Empty Exclude List",
//		ExcludeList: map[string]bool{},
//	}, {
//		Name: "Not Empty Exclude List",
//		ExcludeList: map[string]bool{
//			followerHTTPAddresses[0]: true,
//			followerHTTPAddresses[1]: true,
//		},
//	}}
//
//	for _, test := range tests {
//		t.Run(test.Name, func(t *testing.T) {
//			httpAddresses := nodeService.getRandomFanInPeerHTTPAddresses(test.ExcludeList)
//
//			matchCount := 0
//
//			for _, httpAddress := range httpAddresses {
//				for _, followerHTTPAddress := range followerHTTPAddresses {
//					if followerHTTPAddress == httpAddress {
//						matchCount += 1
//					}
//				}
//			}
//
//			assert.Equal(t, 2, len(httpAddresses))
//			assert.Equal(t, 2, matchCount)
//
//			if len(test.ExcludeList) > 0 {
//				assert.NotContainsf(t, httpAddresses, followerHTTPAddresses[0], "should not contain excluded list")
//				assert.NotContainsf(t, httpAddresses, followerHTTPAddresses[1], "should not contain excluded list")
//			}
//		})
//	}
//}
//
//func Test_selectRandomPeersToFanIn(t *testing.T) {
//	t.Cleanup(func() {
//		err := os.RemoveAll("./raft_data")
//		if err != nil {
//			fmt.Println("failed to remove raft_data dir for test", err)
//		}
//	})
//
//	ctx := context.Background()
//	logger := hclog.New(&hclog.LoggerOptions{
//		Name:  "node-service-test",
//		Level: hclog.LevelFromString("trace"),
//	})
//
//	// Create a temporary SQLite database file
//	tempFile, err := ioutil.TempFile("", "test-db")
//	if err != nil {
//		t.Fatalf("Failed to create temp file: %v", err)
//	}
//	defer os.Remove(tempFile.Name())
//
//	// Create a new SQLite database connection
//	sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
//	sqliteDb.RunMigration()
//	sqliteDb.OpenConnectionToExistingDB()
//
//	scheduler0config := config.NewScheduler0Config()
//	sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
//	scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)
//
//	scheduler0Secrets := secrets.NewScheduler0Secrets()
//	// Create a new FSM store
//	scheduler0Store := fsm.NewFSMStore(logger, scheduler0RaftActions, sqliteDb, nil, nil, nil, nil, sharedRepo)
//
//	// Create a mock raft cluster
//	cluster := raft.MakeClusterCustom(t, &raft.MakeClusterOpts{
//		Peers:          5,
//		Bootstrap:      true,
//		Conf:           raft.DefaultConfig(),
//		ConfigStoreFSM: false,
//		MakeFSMFunc: func() raft.FSM {
//			return scheduler0Store.GetFSM()
//		},
//	})
//	defer cluster.Close()
//	cluster.FullyConnect()
//	scheduler0Store.UpdateRaft(cluster.Leader())
//
//	jobRepo := job_repo.NewJobRepo(logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskManagerRepo := async_task_repo.NewAsyncTasksRepo(ctx, logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskService := async_task.NewAsyncTaskManager(ctx, logger, scheduler0Store, asyncTaskManagerRepo, scheduler0config)
//	jobQueueRepo := job_queue_repo.NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)
//	jobExecutionsRepo := job_execution_repo.NewExecutionsRepo(
//		logger,
//		scheduler0RaftActions,
//		scheduler0Store,
//	)
//
//	dispatcher := utils.NewDispatcher(
//		ctx,
//		int64(1),
//		int64(1),
//	)
//
//	dispatcher.Run()
//
//	queueService := queue.NewJobQueue(ctx, logger, scheduler0config, scheduler0RaftActions, scheduler0Store, jobQueueRepo)
//	httpJobExecutor := executors.NewMockHTTPExecutor(t)
//
//	jobExecutorService := executor.NewJobExecutor(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0RaftActions,
//		jobRepo,
//		jobExecutionsRepo,
//		jobQueueRepo,
//		httpJobExecutor,
//		dispatcher,
//	)
//	nodeHTTPClient := NewHTTPClient(logger, scheduler0config, scheduler0Secrets)
//	jobProcessor := processor.NewMockJobProcessorService(t)
//
//	nodeService := NewNode(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0Secrets,
//		scheduler0Store,
//		scheduler0RaftActions,
//		jobExecutorService,
//		queueService,
//		jobProcessor,
//		jobRepo,
//		sharedRepo,
//		jobExecutionsRepo,
//		asyncTaskService,
//		dispatcher,
//		nodeHTTPClient,
//		nil,
//	)
//	os.Setenv("SCHEDULER0_RAFT_ADDRESS", "http://localhost:3442")
//	defer os.Unsetenv("SCHEDULER0_RAFT_ADDRESS")
//
//	os.Setenv("SCHEDULER0_EXECUTION_LOG_FETCH_FAN_IN", "2")
//	defer os.Unsetenv("SCHEDULER0_EXECUTION_LOG_FETCH_FAN_IN")
//
//	followerHTTPAddresses := []string{
//		"http://localhost:34417",
//		"http://localhost:34416",
//		"http://localhost:34415",
//		"http://localhost:34414",
//	}
//
//	followerRaftAddresses := []string{
//		strings.Split(cluster.Followers()[0].String(), " ")[2],
//		strings.Split(cluster.Followers()[1].String(), " ")[2],
//		strings.Split(cluster.Followers()[2].String(), " ")[2],
//		strings.Split(cluster.Followers()[3].String(), " ")[2],
//	}
//
//	os.Setenv("SCHEDULER0_REPLICAS", fmt.Sprintf("["+
//		"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":5}, "+
//		"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":4}, "+
//		"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":3}, "+
//		"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":2}, "+
//		"{\"raft_address\":\"http://localhost:3442\", \"address\":\"http://localhost:34413\", \"nodeId\":1}"+
//		"]",
//		followerRaftAddresses[0], followerHTTPAddresses[0],
//		followerRaftAddresses[1], followerHTTPAddresses[1],
//		followerRaftAddresses[2], followerHTTPAddresses[2],
//		followerRaftAddresses[3], followerHTTPAddresses[3],
//	))
//	defer os.Unsetenv("SCHEDULER0_REPLICAS")
//
//	nodeService.FsmStore = scheduler0Store
//
//	nodeService.fanIns.Store(followerHTTPAddresses[0], models.PeerFanIn{
//		PeerHTTPAddress: followerHTTPAddresses[0],
//	})
//	nodeService.fanIns.Store(followerHTTPAddresses[1], models.PeerFanIn{
//		PeerHTTPAddress: followerHTTPAddresses[1],
//	})
//
//	os.Setenv("SCHEDULER0_RAFT_ADDRESS", strings.Split(cluster.Leader().String(), " ")[2])
//
//	peerFanIns := nodeService.selectRandomPeersToFanIn()
//	count := 0
//	nodeService.fanIns.Range(func(key, value any) bool {
//		count += 1
//		return true
//	})
//	assert.Equal(t, 2, len(peerFanIns))
//	assert.Equal(t, 4, count)
//}
//
//func Test_fanInLocalDataFromPeers(t *testing.T) {
//	t.Cleanup(func() {
//		err := os.RemoveAll("./raft_data")
//		if err != nil {
//			fmt.Println("failed to remove raft_data dir for test", err)
//		}
//	})
//	ctx := context.Background()
//	logger := hclog.New(&hclog.LoggerOptions{
//		Name:  "node-service-test",
//		Level: hclog.LevelFromString("trace"),
//	})
//
//	// Create a temporary SQLite database file
//	tempFile, err := ioutil.TempFile("", "test-db")
//	if err != nil {
//		t.Fatalf("Failed to create temp file: %v", err)
//	}
//	defer os.Remove(tempFile.Name())
//
//	os.Setenv("SCHEDULER0_EXECUTION_LOG_FETCH_INTERVAL_SECONDS", "2")
//	defer os.Unsetenv("SCHEDULER0_EXECUTION_LOG_FETCH_INTERVAL_SECONDS")
//
//	// Create a new SQLite database connection
//	sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
//	sqliteDb.RunMigration()
//	sqliteDb.OpenConnectionToExistingDB()
//
//	scheduler0config := config.NewScheduler0Config()
//	sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
//	scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)
//
//	scheduler0Secrets := secrets.NewScheduler0Secrets()
//	// Create a new FSM store
//	scheduler0Store := fsm.NewFSMStore(logger, scheduler0RaftActions, sqliteDb, nil, nil, nil, nil, sharedRepo)
//
//	// Create a mock raft cluster
//	cluster := raft.MakeClusterCustom(t, &raft.MakeClusterOpts{
//		Peers:          6,
//		Bootstrap:      true,
//		Conf:           raft.DefaultConfig(),
//		ConfigStoreFSM: false,
//		MakeFSMFunc: func() raft.FSM {
//			return scheduler0Store.GetFSM()
//		},
//	})
//	defer cluster.Close()
//	cluster.FullyConnect()
//
//	scheduler0Store.UpdateRaft(cluster.Leader())
//
//	jobRepo := job_repo.NewJobRepo(logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskManagerRepo := async_task_repo.NewAsyncTasksRepo(ctx, logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskService := async_task.NewAsyncTaskManager(ctx, logger, scheduler0Store, asyncTaskManagerRepo, scheduler0config)
//	jobQueueRepo := job_queue_repo.NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)
//	jobExecutionsRepo := job_execution_repo.NewExecutionsRepo(
//		logger,
//		scheduler0RaftActions,
//		scheduler0Store,
//	)
//
//	dispatcher := utils.NewDispatcher(
//		ctx,
//		int64(1),
//		int64(1),
//	)
//
//	queueService := queue.NewJobQueue(ctx, logger, scheduler0config, scheduler0RaftActions, scheduler0Store, jobQueueRepo)
//	httpJobExecutor := executors.NewMockHTTPExecutor(t)
//
//	jobExecutorService := executor.NewJobExecutor(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0RaftActions,
//		jobRepo,
//		jobExecutionsRepo,
//		jobQueueRepo,
//		httpJobExecutor,
//		dispatcher,
//	)
//
//	os.Setenv("SCHEDULER0_RAFT_ADDRESS", strings.Split(cluster.Leader().String(), " ")[2])
//
//	os.Setenv("SCHEDULER0_EXECUTION_LOG_FETCH_FAN_IN", "2")
//
//	followerHTTPAddresses := []string{
//		"http://localhost:34417",
//		"http://localhost:34416",
//		"http://localhost:34415",
//		"http://localhost:34414",
//		"http://localhost:34413",
//	}
//
//	followerRaftAddresses := []string{
//		strings.Split(cluster.Followers()[0].String(), " ")[2],
//		strings.Split(cluster.Followers()[1].String(), " ")[2],
//		strings.Split(cluster.Followers()[2].String(), " ")[2],
//		strings.Split(cluster.Followers()[3].String(), " ")[2],
//		strings.Split(cluster.Followers()[4].String(), " ")[2],
//	}
//
//	tests := []struct {
//		cachedFanIns []struct {
//			httpAddress string
//			state       models.PeerFanInState
//		}
//		method    string
//		startPort string
//	}{
//		{
//			method:    "FetchUncommittedLogsFromPeersPhase1",
//			startPort: "localhost:34345",
//			cachedFanIns: []struct {
//				httpAddress string
//				state       models.PeerFanInState
//			}{
//				{
//					httpAddress: followerHTTPAddresses[0],
//					state:       models.PeerFanInStateNotStated,
//				},
//				{
//					httpAddress: followerHTTPAddresses[1],
//					state:       models.PeerFanInStateNotStated,
//				},
//				{
//					httpAddress: followerHTTPAddresses[2],
//					state:       models.PeerFanInStateNotStated,
//				},
//				{
//					httpAddress: followerHTTPAddresses[3],
//					state:       models.PeerFanInStateNotStated,
//				},
//			},
//		},
//		{
//			method:    "FetchUncommittedLogsFromPeersPhase2",
//			startPort: "localhost:34346",
//			cachedFanIns: []struct {
//				httpAddress string
//				state       models.PeerFanInState
//			}{
//				{
//					httpAddress: followerHTTPAddresses[0],
//					state:       models.PeerFanInStateGetRequestId,
//				},
//				{
//					httpAddress: followerHTTPAddresses[1],
//					state:       models.PeerFanInStateGetRequestId,
//				},
//				{
//					httpAddress: followerHTTPAddresses[2],
//					state:       models.PeerFanInStateGetRequestId,
//				},
//				{
//					httpAddress: followerHTTPAddresses[3],
//					state:       models.PeerFanInStateGetRequestId,
//				},
//				{
//					httpAddress: followerHTTPAddresses[4],
//					state:       models.PeerFanInStateGetRequestId,
//				},
//			},
//		},
//	}
//
//	for _, test := range tests {
//
//		t.Run(test.method, func(t *testing.T) {
//			nodeHTTPClient := NewMockNodeClient(t)
//			os.Setenv("SCHEDULER0_REPLICAS", fmt.Sprintf("["+
//				"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":6}, "+
//				"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":5}, "+
//				"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":4}, "+
//				"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":3}, "+
//				"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":2}, "+
//				"{\"raft_address\":\"%v\", \"address\":\"http://localhost:34410\", \"nodeId\":1}"+
//				"]",
//				followerRaftAddresses[0], followerHTTPAddresses[0],
//				followerRaftAddresses[1], followerHTTPAddresses[1],
//				followerRaftAddresses[2], followerHTTPAddresses[2],
//				followerRaftAddresses[3], followerHTTPAddresses[3],
//				followerRaftAddresses[4], followerHTTPAddresses[4],
//				test.startPort,
//			))
//			os.Setenv("SCHEDULER0_RAFT_ADDRESS", test.startPort)
//			jobProcessor := processor.NewMockJobProcessorService(t)
//
//			nodeService := NewNode(
//				ctx,
//				logger,
//				scheduler0config,
//				scheduler0Secrets,
//				scheduler0RaftActions,
//				scheduler0Store,
//				jobExecutorService,
//				queueService,
//				jobProcessor,
//				jobRepo,
//				sharedRepo,
//				jobExecutionsRepo,
//				asyncTaskService,
//				dispatcher,
//				nodeHTTPClient,
//				nil,
//			)
//
//			nodeService.FsmStore = scheduler0Store
//
//			for _, cachedHttpAddess := range test.cachedFanIns {
//				nodeService.fanIns.Store(cachedHttpAddess.httpAddress, models.PeerFanIn{
//					PeerHTTPAddress: cachedHttpAddess.httpAddress,
//					State:           cachedHttpAddess.state,
//				})
//			}
//
//			nodeHTTPClient.On(test.method, mock.Anything, mock.Anything, mock.Anything)
//
//			os.Setenv("SCHEDULER0_RAFT_ADDRESS", strings.Split(cluster.Leader().String(), " ")[2])
//			os.Setenv("SCHEDULER0_REPLICAS", fmt.Sprintf("["+
//				"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":6}, "+
//				"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":5}, "+
//				"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":4}, "+
//				"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":3}, "+
//				"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":2}, "+
//				"{\"raft_address\":\"%v\", \"address\":\"http://localhost:34410\", \"nodeId\":1}"+
//				"]",
//				followerRaftAddresses[0], followerHTTPAddresses[0],
//				followerRaftAddresses[1], followerHTTPAddresses[1],
//				followerRaftAddresses[2], followerHTTPAddresses[2],
//				followerRaftAddresses[3], followerHTTPAddresses[3],
//				followerRaftAddresses[4], followerHTTPAddresses[4],
//				strings.Split(cluster.Leader().String(), " ")[2],
//			))
//
//			nodeService.fanInLocalDataFromPeers()
//
//			time.Sleep(time.Second * time.Duration(3))
//			validFirstCallArgs := map[string]bool{
//				followerHTTPAddresses[0]: true,
//				followerHTTPAddresses[1]: true,
//				followerHTTPAddresses[2]: true,
//				followerHTTPAddresses[3]: true,
//				followerHTTPAddresses[4]: true,
//			}
//
//			firstCallArgs := nodeHTTPClient.Calls[0].Arguments.Get(2).([]models.PeerFanIn)
//
//			for _, firstCallArg := range firstCallArgs {
//				if ok := validFirstCallArgs[firstCallArg.PeerHTTPAddress]; !ok {
//					t.Fatalf("should not have performed %v for %v", test.method, firstCallArg)
//				}
//			}
//
//			for _, cachedHttpAddess := range test.cachedFanIns {
//				nodeService.fanIns.Delete(cachedHttpAddess.httpAddress)
//			}
//		})
//
//		//_, ok := nodeService.fanIns.Load(followerHTTPAddresses[2])
//		//if ok {
//		//	t.Fatalf("should have removed %v from store", followerHTTPAddresses[2])
//		//}
//	}
//	os.Unsetenv("SCHEDULER0_RAFT_ADDRESS")
//	os.Unsetenv("SCHEDULER0_EXECUTION_LOG_FETCH_FAN_IN")
//	os.Unsetenv("SCHEDULER0_REPLICAS")
//}
//
//func Test_handleLeaderChange(t *testing.T) {
//	t.Run("HandleLeaderChange", func(t *testing.T) {
//		t.Cleanup(func() {
//			err := os.RemoveAll("./raft_data")
//			if err != nil {
//				fmt.Println("failed to remove raft_data dir for test", err)
//			}
//			err = os.RemoveAll("./sqlite_data")
//			if err != nil {
//				fmt.Println("failed to remove sqlite_data dir for test", err)
//			}
//		})
//		ctx, cancelr := context.WithCancel(context.Background())
//		defer cancelr()
//		logger := hclog.New(&hclog.LoggerOptions{
//			Name:  "node-service-test",
//			Level: hclog.LevelFromString("trace"),
//		})
//
//		// Create a temporary SQLite database file
//		tempFile, err := ioutil.TempFile("", "test-db")
//		if err != nil {
//			t.Fatalf("Failed to create temp file: %v", err)
//		}
//		defer os.Remove(tempFile.Name())
//
//		// Create a new SQLite database connection
//		sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
//		sqliteDb.RunMigration()
//		sqliteDb.OpenConnectionToExistingDB()
//
//		scheduler0config := config.NewScheduler0Config()
//		sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
//		scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)
//
//		scheduler0Secrets := secrets.NewScheduler0Secrets()
//		scheduler0Store := fsm.NewFSMStore(logger, scheduler0RaftActions, sqliteDb, nil, nil, nil, nil, sharedRepo)
//
//		cluster := raft.MakeClusterCustom(t, &raft.MakeClusterOpts{
//			Peers:          2,
//			Bootstrap:      true,
//			Conf:           raft.DefaultConfig(),
//			ConfigStoreFSM: false,
//			MakeFSMFunc: func() raft.FSM {
//				return scheduler0Store.GetFSM()
//			},
//		})
//		cluster.FullyConnect()
//
//		clusterLeader := cluster.Leader()
//		scheduler0Store.UpdateRaft(clusterLeader)
//
//		jobRepo := job_repo.NewJobRepo(logger, scheduler0RaftActions, scheduler0Store)
//		projectRepo := project_repo.NewProjectRepo(logger, scheduler0RaftActions, scheduler0Store, jobRepo)
//		asyncTaskManagerRepo := async_task_repo.NewAsyncTasksRepo(ctx, logger, scheduler0RaftActions, scheduler0Store)
//		asyncTaskService := async_task.NewAsyncTaskManager(ctx, logger, scheduler0Store, asyncTaskManagerRepo, scheduler0config)
//		jobQueueRepo := job_queue_repo.NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)
//		jobExecutionsRepo := job_execution_repo.NewExecutionsRepo(
//			logger,
//			scheduler0RaftActions,
//			scheduler0Store,
//		)
//
//		dispatcher := utils.NewDispatcher(
//			ctx,
//			int64(1),
//			int64(1),
//		)
//
//		dispatcher.Run()
//
//		queueService := queue.NewJobQueue(ctx, logger, scheduler0config, scheduler0RaftActions, scheduler0Store, jobQueueRepo)
//		httpJobExecutor := executors.NewMockHTTPExecutor(t)
//
//		jobExecutorService := executor.NewJobExecutor(
//			ctx,
//			logger,
//			scheduler0config,
//			scheduler0RaftActions,
//			jobRepo,
//			jobExecutionsRepo,
//			jobQueueRepo,
//			httpJobExecutor,
//			dispatcher,
//		)
//		nodeHTTPClient := NewMockNodeClient(t)
//		nodeHTTPClient.On("StopJobs", mock.Anything, mock.Anything, mock.Anything).Return(nil)
//
//		jobService := job.NewJobService(ctx, logger, jobRepo, queueService, projectRepo, dispatcher, asyncTaskService)
//		jobProcessor := processor.NewMockJobProcessorService(t)
//
//		nodeService := NewNode(
//			ctx,
//			logger,
//			scheduler0config,
//			scheduler0Secrets,
//			scheduler0Store,
//			scheduler0RaftActions,
//			jobExecutorService,
//			queueService,
//			jobProcessor,
//			jobRepo,
//			sharedRepo,
//			jobExecutionsRepo,
//			asyncTaskService,
//			dispatcher,
//			nodeHTTPClient,
//			nil,
//		)
//		nodeService.FsmStore = scheduler0Store
//		os.Setenv("SCHEDULER0_PORT", "34410")
//		defer os.Unsetenv("SCHEDULER0_PORT")
//		os.Setenv("SCHEDULER0_PROTOCOL", "http")
//		defer os.Unsetenv("SCHEDULER0_PROTOCOL")
//		os.Setenv("SCHEDULER0_HOST", "localhost")
//		defer os.Unsetenv("SCHEDULER0_HOST")
//		os.Setenv("SCHEDULER0_EXECUTION_LOG_FETCH_INTERVAL_SECONDS", "2")
//		defer os.Unsetenv("SCHEDULER0_EXECUTION_LOG_FETCH_INTERVAL_SECONDS")
//		os.Setenv("SCHEDULER0_REPLICAS", fmt.Sprintf("["+
//			"{\"raft_address\":\"%v\", \"address\":\"http://localhost:34410\", \"nodeId\":1},"+
//			"{\"raft_address\":\"%v\", \"address\":\"http://localhost:34411\", \"nodeId\":2}]",
//			strings.Split(cluster.Leader().String(), " ")[2],
//			strings.Split(cluster.Followers()[0].String(), " ")[2],
//		))
//
//		asyncTaskService.ListenForNotifications()
//		// Define the input jobs
//		jobs := []models.Job{}
//
//		i := 1
//		for i < 100 {
//			jobs = append(jobs, models.Job{
//				ID:        uint64(i),
//				Spec:      "@every 1h",
//				Timezone:  "America/New_York",
//				ProjectID: 1,
//			})
//			i++
//		}
//
//		// Create the projects using the project repo
//		project := models.Project{
//			ID:          1,
//			Name:        fmt.Sprintf("Project %d", 1),
//			Description: fmt.Sprintf("Project %d description", 1),
//		}
//		_, createErr := projectRepo.CreateOne(&project)
//		if createErr != nil {
//			t.Fatalf("Failed to create project: %v", createErr)
//		}
//
//		// Call the BatchInsertJobs method of the job service
//		_, batchErr := jobService.BatchInsertJobs("request123", jobs)
//		if batchErr != nil {
//			t.Fatalf("Failed to insert jobs: %v", batchErr)
//		}
//
//		time.Sleep(time.Second * time.Duration(2))
//
//		schedule, parseErr := cron.Parse("@every 1h")
//		if parseErr != nil {
//			t.Fatal("cron spec error", parseErr)
//		}
//
//		schedulerTime := scheduler0time.GetSchedulerTime()
//		now := schedulerTime.GetTime(time.Now())
//		nextTime := schedule.Next(now)
//		prevNextTime := nextTime.Add(-nextTime.Sub(now))
//		lastTime := nextTime.Add(-nextTime.Sub(now)).Add(-nextTime.Sub(now))
//
//		serr := os.Setenv("SCHEDULER0_NODE_ID", "1")
//		defer os.Unsetenv("SCHEDULER0_NODE_ID")
//		if serr != nil {
//			t.Fatal("failed to set env", serr)
//		}
//		uncommittedExecutionsLogs := []models.JobExecutionLog{}
//		i = 0
//		for i < 24 {
//			uncommittedExecutionsLogs = append(uncommittedExecutionsLogs, models.JobExecutionLog{
//				JobId:                 jobs[i].ID,
//				UniqueId:              fmt.Sprintf("%d-%d", jobs[i].ID, i),
//				State:                 models.ExecutionLogSuccessState,
//				LastExecutionDatetime: lastTime,
//				NextExecutionDatetime: prevNextTime,
//				NodeId:                1,
//			})
//			i++
//		}
//		err = sharedRepo.InsertExecutionLogs(sqliteDb, false, uncommittedExecutionsLogs)
//		if err != nil {
//			t.Fatal("failed to insert execution logs", err)
//		}
//
//		nodeService.handleRaftLeadershipChanges(true)
//
//		time.Sleep(time.Second * time.Duration(2))
//
//		committedLogs, err := sharedRepo.GetExecutionLogs(sqliteDb, true)
//		if err != nil {
//			t.Fatal("failed to get committed execution logs", err)
//		}
//
//		assert.Equal(t, 2, len(nodeService.jobQueue.GetJobAllocations()))
//		assert.Equal(t, false, nodeService.jobQueue.GetSingleNodeMode())
//		assert.Equal(t, false, nodeService.jobExecutor.GetSingleNodeMode())
//		assert.Equal(t, false, nodeService.asyncTaskManager.GetSingleNodeMode())
//		assert.Equal(t, len(uncommittedExecutionsLogs), len(committedLogs))
//	})
//}
//
//func Test_commitFetchedUnCommittedLogs(t *testing.T) {
//	t.Cleanup(func() {
//		err := os.RemoveAll("./raft_data")
//		if err != nil {
//			fmt.Println("failed to remove raft_data dir for test", err)
//		}
//	})
//	ctx := context.Background()
//	logger := hclog.New(&hclog.LoggerOptions{
//		Name:  "node-service-test",
//		Level: hclog.LevelFromString("trace"),
//	})
//
//	// Create a temporary SQLite database file
//	tempFile, err := ioutil.TempFile("", "test-db")
//	if err != nil {
//		t.Fatalf("Failed to create temp file: %v", err)
//	}
//	defer os.Remove(tempFile.Name())
//
//	os.Setenv("SCHEDULER0_EXECUTION_LOG_FETCH_INTERVAL_SECONDS", "2")
//	defer os.Unsetenv("SCHEDULER0_EXECUTION_LOG_FETCH_INTERVAL_SECONDS")
//
//	// Create a new SQLite database connection
//	sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
//	sqliteDb.RunMigration()
//	sqliteDb.OpenConnectionToExistingDB()
//
//	scheduler0config := config.NewScheduler0Config()
//	sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
//	scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)
//
//	scheduler0Secrets := secrets.NewScheduler0Secrets()
//	// Create a new FSM store
//	scheduler0Store := fsm.NewFSMStore(logger, scheduler0RaftActions, sqliteDb, nil, nil, nil, nil, sharedRepo)
//
//	// Create a mock raft cluster
//	cluster := raft.MakeClusterCustom(t, &raft.MakeClusterOpts{
//		Peers:          2,
//		Bootstrap:      true,
//		Conf:           raft.DefaultConfig(),
//		ConfigStoreFSM: false,
//		MakeFSMFunc: func() raft.FSM {
//			return scheduler0Store.GetFSM()
//		},
//	})
//	defer cluster.Close()
//	cluster.FullyConnect()
//
//	scheduler0Store.UpdateRaft(cluster.Leader())
//
//	jobRepo := job_repo.NewJobRepo(logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskManagerRepo := async_task_repo.NewAsyncTasksRepo(ctx, logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskService := async_task.NewAsyncTaskManager(ctx, logger, scheduler0Store, asyncTaskManagerRepo, scheduler0config)
//	jobQueueRepo := job_queue_repo.NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)
//	jobExecutionsRepo := job_execution_repo.NewExecutionsRepo(
//		logger,
//		scheduler0RaftActions,
//		scheduler0Store,
//	)
//
//	dispatcher := utils.NewDispatcher(
//		ctx,
//		int64(1),
//		int64(1),
//	)
//
//	queueService := queue.NewJobQueue(ctx, logger, scheduler0config, scheduler0RaftActions, scheduler0Store, jobQueueRepo)
//	httpJobExecutor := executors.NewMockHTTPExecutor(t)
//
//	jobExecutorService := executor.NewJobExecutor(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0RaftActions,
//		jobRepo,
//		jobExecutionsRepo,
//		jobQueueRepo,
//		httpJobExecutor,
//		dispatcher,
//	)
//
//	leaderRaftAddress := strings.Split(cluster.Leader().String(), " ")[2]
//	os.Setenv("SCHEDULER0_RAFT_ADDRESS", leaderRaftAddress)
//	defer os.Unsetenv("SCHEDULER0_RAFT_ADDRESS")
//
//	followerHTTPAddresses := []string{
//		"http://localhost:34416",
//	}
//
//	followerRaftAddresses := []string{
//		strings.Split(cluster.Followers()[0].String(), " ")[2],
//	}
//	leaderHttpAddress := "http://localhost:34410"
//
//	nodeHTTPClient := NewMockNodeClient(t)
//	os.Setenv("SCHEDULER0_REPLICAS", fmt.Sprintf("["+
//		"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":2}, "+
//		"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":1}]",
//		followerRaftAddresses[0], followerHTTPAddresses[0],
//		leaderRaftAddress, leaderHttpAddress,
//	))
//	defer os.Unsetenv("SCHEDULER0_REPLICAS")
//	jobProcessor := processor.NewMockJobProcessorService(t)
//
//	nodeService := NewNode(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0Secrets,
//		scheduler0Store,
//		scheduler0RaftActions,
//		jobExecutorService,
//		queueService,
//		jobProcessor,
//		jobRepo,
//		sharedRepo,
//		jobExecutionsRepo,
//		asyncTaskService,
//		dispatcher,
//		nodeHTTPClient,
//		nil,
//	)
//
//	nodeService.FsmStore = scheduler0Store
//
//	nodeService.fanIns.Store(followerHTTPAddresses[0], models.PeerFanIn{
//		PeerHTTPAddress: followerHTTPAddresses[0],
//		State:           models.PeerFanInStateComplete,
//	})
//
//	var receivedFanIn models.PeerFanIn
//	go func() {
//		for {
//			select {
//			case fanedInPeer := <-nodeService.fanInCh:
//				receivedFanIn = fanedInPeer
//			}
//		}
//	}()
//
//	nodeService.commitFetchedUnCommittedLogs([]models.PeerFanIn{
//		{
//			PeerHTTPAddress: followerHTTPAddresses[0],
//			State:           models.PeerFanInStateComplete,
//		},
//	})
//
//	time.Sleep(time.Second * time.Duration(2))
//
//	_, found := nodeService.fanIns.Load(followerHTTPAddresses[0])
//	assert.Equal(t, false, found)
//	assert.Equal(t, int(receivedFanIn.State), models.PeerFanInStateComplete)
//	assert.Equal(t, receivedFanIn.PeerHTTPAddress, followerHTTPAddresses[0])
//}
//
//func Test_handleUncommittedAsyncTasks(t *testing.T) {
//	t.Cleanup(func() {
//		err := os.RemoveAll("./raft_data")
//		if err != nil {
//			fmt.Println("failed to remove raft_data dir for test", err)
//		}
//	})
//	ctx := context.Background()
//	logger := hclog.New(&hclog.LoggerOptions{
//		Name:  "node-service-test",
//		Level: hclog.LevelFromString("trace"),
//	})
//
//	// Create a temporary SQLite database file
//	tempFile, err := ioutil.TempFile("", "test-db")
//	if err != nil {
//		t.Fatalf("Failed to create temp file: %v", err)
//	}
//	defer os.Remove(tempFile.Name())
//
//	os.Setenv("SCHEDULER0_EXECUTION_LOG_FETCH_INTERVAL_SECONDS", "2")
//	defer os.Unsetenv("SCHEDULER0_EXECUTION_LOG_FETCH_INTERVAL_SECONDS")
//
//	// Create a new SQLite database connection
//	sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
//	sqliteDb.RunMigration()
//	sqliteDb.OpenConnectionToExistingDB()
//
//	scheduler0config := config.NewScheduler0Config()
//	sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
//	scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)
//
//	scheduler0Secrets := secrets.NewScheduler0Secrets()
//	// Create a new FSM store
//	scheduler0Store := fsm.NewFSMStore(logger, scheduler0RaftActions, sqliteDb, nil, nil, nil, nil, sharedRepo)
//
//	// Create a mock raft cluster
//	cluster := raft.MakeClusterCustom(t, &raft.MakeClusterOpts{
//		Peers:          2,
//		Bootstrap:      true,
//		Conf:           raft.DefaultConfig(),
//		ConfigStoreFSM: false,
//		MakeFSMFunc: func() raft.FSM {
//			return scheduler0Store.GetFSM()
//		},
//	})
//	defer cluster.Close()
//	cluster.FullyConnect()
//
//	scheduler0Store.UpdateRaft(cluster.Leader())
//
//	jobRepo := job_repo.NewJobRepo(logger, scheduler0RaftActions, scheduler0Store)
//	projectRepo := project_repo.NewProjectRepo(logger, scheduler0RaftActions, scheduler0Store, jobRepo)
//	asyncTaskManagerRepo := async_task_repo.NewAsyncTasksRepo(ctx, logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskService := async_task.NewAsyncTaskManager(ctx, logger, scheduler0Store, asyncTaskManagerRepo, scheduler0config)
//	jobQueueRepo := job_queue_repo.NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)
//	jobExecutionsRepo := job_execution_repo.NewExecutionsRepo(
//		logger,
//		scheduler0RaftActions,
//		scheduler0Store,
//	)
//
//	dispatcher := utils.NewDispatcher(
//		ctx,
//		int64(1),
//		int64(1),
//	)
//
//	dispatcher.Run()
//
//	queueService := queue.NewJobQueue(ctx, logger, scheduler0config, scheduler0RaftActions, scheduler0Store, jobQueueRepo)
//	httpJobExecutor := executors.NewMockHTTPExecutor(t)
//
//	jobExecutorService := executor.NewJobExecutor(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0RaftActions,
//		jobRepo,
//		jobExecutionsRepo,
//		jobQueueRepo,
//		httpJobExecutor,
//		dispatcher,
//	)
//
//	leaderRaftAddress := strings.Split(cluster.Leader().String(), " ")[2]
//	os.Setenv("SCHEDULER0_RAFT_ADDRESS", leaderRaftAddress)
//	defer os.Unsetenv("SCHEDULER0_RAFT_ADDRESS")
//
//	followerHTTPAddresses := []string{
//		"http://localhost:34416",
//	}
//
//	followerRaftAddresses := []string{
//		strings.Split(cluster.Followers()[0].String(), " ")[2],
//	}
//	leaderHttpAddress := "http://localhost:34410"
//
//	ctx, canceler := context.WithCancel(context.Background())
//	defer canceler()
//	nodeHTTPClient := NewMockNodeClient(t)
//	os.Setenv("SCHEDULER0_REPLICAS", fmt.Sprintf("["+
//		"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":2}, "+
//		"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":1}]",
//		followerRaftAddresses[0], followerHTTPAddresses[0],
//		leaderRaftAddress, leaderHttpAddress,
//	))
//	defer os.Unsetenv("SCHEDULER0_REPLICAS")
//	jobProcessor := processor.NewMockJobProcessorService(t)
//	nodeService := NewNode(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0Secrets,
//		scheduler0Store,
//		scheduler0RaftActions,
//		jobExecutorService,
//		queueService,
//		jobProcessor,
//		jobRepo,
//		sharedRepo,
//		jobExecutionsRepo,
//		asyncTaskService,
//		dispatcher,
//		nodeHTTPClient,
//		nil,
//	)
//
//	nodeService.FsmStore = scheduler0Store
//
//	project := models.Project{
//		ID:          1,
//		Name:        fmt.Sprintf("Project %d", 1),
//		Description: fmt.Sprintf("Project %d description", 1),
//	}
//	projectId, createErr := projectRepo.CreateOne(&project)
//	if createErr != nil {
//		t.Fatalf("Failed to create project: %v", createErr)
//	}
//
//	fmt.Println("projectId", projectId)
//
//	createJobJsonPayload := "[{" +
//		"\"spec\": \"@every 1m\"," +
//		"\"projectId\": 1," +
//		"\"executionType\": \"http\"," +
//		"\"data\": \"{}\"," +
//		"\"timezone\": \"America/New_York\"," +
//		"\"callbackUrl\": \"http://localhost:3000/callback\"" +
//		"}]"
//
//	asyncTaskService.ListenForNotifications()
//	ids, createTErr := asyncTaskService.AddTasks(createJobJsonPayload, "request-id-1", "create_job")
//	if createTErr != nil {
//		t.Fatalf("failed to create async task error: %v", createTErr)
//	}
//
//	nodeService.handleUncommittedAsyncTasks([]models.AsyncTask{
//		{
//			RequestId: "request-id-1",
//			Service:   "create_job",
//			Input:     createJobJsonPayload,
//		},
//	})
//
//	time.Sleep(time.Second * 1)
//
//	taskCh, _, getTErr := asyncTaskService.GetTaskBlocking(ids[0])
//	if getTErr != nil {
//		t.Fatalf("failed to get task error: %v", getTErr)
//	}
//
//	time.Sleep(time.Second * 1)
//
//	var taskRes models.AsyncTask
//	go func() {
//		taskRes = <-taskCh
//	}()
//
//	time.Sleep(time.Second * 1)
//
//	var jobRes utils.FSMResponse
//	umerr := json.Unmarshal([]byte(taskRes.Output), &jobRes)
//	if umerr != nil {
//		t.Fatalf("failed to unmarshal job error: %v", umerr)
//	}
//
//	jobIds := jobRes.Data.([]interface{})
//
//	assert.Equal(t, float64(1), jobIds[0])
//}
//
//func Test_authRaftConfiguration(t *testing.T) {
//	t.Cleanup(func() {
//		err := os.RemoveAll("./raft_data")
//		if err != nil {
//			fmt.Println("failed to remove raft_data dir for test", err)
//		}
//	})
//	ctx := context.Background()
//	logger := hclog.New(&hclog.LoggerOptions{
//		Name:  "node-service-test",
//		Level: hclog.LevelFromString("trace"),
//	})
//
//	// Create a temporary SQLite database file
//	tempFile, err := ioutil.TempFile("", "test-db")
//	if err != nil {
//		t.Fatalf("Failed to create temp file: %v", err)
//	}
//	defer os.Remove(tempFile.Name())
//
//	// Create a new SQLite database connection
//	sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
//	sqliteDb.RunMigration()
//	sqliteDb.OpenConnectionToExistingDB()
//
//	scheduler0config := config.NewScheduler0Config()
//	sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
//	scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)
//
//	scheduler0Secrets := secrets.NewScheduler0Secrets()
//	// Create a new FSM store
//	scheduler0Store := fsm.NewFSMStore(logger, scheduler0RaftActions, sqliteDb, nil, nil, nil, nil, sharedRepo)
//
//	// Create a mock raft cluster
//	cluster := raft.MakeClusterCustom(t, &raft.MakeClusterOpts{
//		Peers:          2,
//		Bootstrap:      true,
//		Conf:           raft.DefaultConfig(),
//		ConfigStoreFSM: false,
//		MakeFSMFunc: func() raft.FSM {
//			return scheduler0Store.GetFSM()
//		},
//	})
//	defer cluster.Close()
//	cluster.FullyConnect()
//
//	scheduler0Store.UpdateRaft(cluster.Leader())
//
//	jobRepo := job_repo.NewJobRepo(logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskManagerRepo := async_task_repo.NewAsyncTasksRepo(ctx, logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskService := async_task.NewAsyncTaskManager(ctx, logger, scheduler0Store, asyncTaskManagerRepo, scheduler0config)
//	jobQueueRepo := job_queue_repo.NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)
//	jobExecutionsRepo := job_execution_repo.NewExecutionsRepo(
//		logger,
//		scheduler0RaftActions,
//		scheduler0Store,
//	)
//
//	dispatcher := utils.NewDispatcher(
//		ctx,
//		int64(1),
//		int64(1),
//	)
//
//	dispatcher.Run()
//
//	queueService := queue.NewJobQueue(ctx, logger, scheduler0config, scheduler0RaftActions, scheduler0Store, jobQueueRepo)
//	httpJobExecutor := executors.NewMockHTTPExecutor(t)
//
//	jobExecutorService := executor.NewJobExecutor(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0RaftActions,
//		jobRepo,
//		jobExecutionsRepo,
//		jobQueueRepo,
//		httpJobExecutor,
//		dispatcher,
//	)
//
//	leaderRaftAddress := strings.Split(cluster.Leader().String(), " ")[2]
//	os.Setenv("SCHEDULER0_RAFT_ADDRESS", leaderRaftAddress)
//	defer os.Unsetenv("SCHEDULER0_RAFT_ADDRESS")
//
//	followerHTTPAddresses := []string{
//		"http://localhost:34416",
//	}
//
//	followerRaftAddresses := []string{
//		strings.Split(cluster.Followers()[0].String(), " ")[2],
//	}
//	leaderHttpAddress := "http://localhost:34410"
//
//	nodeHTTPClient := NewMockNodeClient(t)
//	os.Setenv("SCHEDULER0_REPLICAS", fmt.Sprintf("["+
//		"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":2}, "+
//		"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":1}]",
//		followerRaftAddresses[0], followerHTTPAddresses[0],
//		leaderRaftAddress, leaderHttpAddress,
//	))
//	defer os.Unsetenv("SCHEDULER0_REPLICAS")
//
//	jobProcessor := processor.NewMockJobProcessorService(t)
//	nodeService := NewNode(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0Secrets,
//		scheduler0Store,
//		scheduler0RaftActions,
//		jobExecutorService,
//		queueService,
//		jobProcessor,
//		jobRepo,
//		sharedRepo,
//		jobExecutionsRepo,
//		asyncTaskService,
//		dispatcher,
//		nodeHTTPClient,
//		nil,
//	)
//
//	nodeService.FsmStore = scheduler0Store
//
//	status := Status{
//		IsAuth:   true,
//		IsAlive:  true,
//		IsLeader: false,
//	}
//	nodeHTTPClient.On("ConnectNode", mock.Anything).Return(&status, nil)
//
//	os.Setenv("SCHEDULER0_PORT", "34410")
//	defer os.Unsetenv("SCHEDULER0_PORT")
//	os.Setenv("SCHEDULER0_PROTOCOL", "http")
//	defer os.Unsetenv("SCHEDULER0_PROTOCOL")
//	os.Setenv("SCHEDULER0_HOST", "localhost")
//	defer os.Unsetenv("SCHEDULER0_HOST")
//
//	cfg := nodeService.authRaftConfiguration()
//
//	time.Sleep(time.Second * 1)
//	firstCallArgs := nodeHTTPClient.Calls[0].Arguments.Get(0).(config.RaftNode)
//
//	assert.Equal(t, 2, len(cfg.Servers))
//	assert.Equal(t, followerRaftAddresses[0], string(cfg.Servers[1].Address))
//	assert.Equal(t, firstCallArgs.Address, followerHTTPAddresses[0])
//}
//
//func Test_recoverRaftState(t *testing.T) {
//	t.Cleanup(func() {
//		err := os.RemoveAll("./raft_data")
//		if err != nil {
//			fmt.Println("failed to remove raft_data dir for test", err)
//		}
//		err = os.RemoveAll("./sqlite_data")
//		if err != nil {
//			fmt.Println("failed to remove sqlite_data dir for test", err)
//		}
//	})
//	ctx := context.Background()
//	logger := hclog.New(&hclog.LoggerOptions{
//		Name:  "node-service-test",
//		Level: hclog.LevelFromString("trace"),
//	})
//
//	// Create a temporary SQLite database file
//	tempFile, err := ioutil.TempFile("", "test-db")
//	if err != nil {
//		t.Fatalf("Failed to create temp file: %v", err)
//	}
//	defer os.Remove(tempFile.Name())
//
//	// Create a new SQLite database connection
//	sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
//	sqliteDb.RunMigration()
//	sqliteDb.OpenConnectionToExistingDB()
//
//	scheduler0config := config.NewScheduler0Config()
//	sharedRepo := shared_repo.NewMockSharedRepo(t)
//	scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)
//
//	scheduler0Secrets := secrets.NewScheduler0Secrets()
//	// Create a new FSM store
//	scheduler0Store := fsm.NewFSMStore(logger, scheduler0RaftActions, sqliteDb, nil, nil, nil, nil, sharedRepo)
//
//	// Create a mock raft cluster
//	cluster := raft.MakeClusterCustom(t, &raft.MakeClusterOpts{
//		Peers:          2,
//		Bootstrap:      true,
//		Conf:           raft.DefaultConfig(),
//		ConfigStoreFSM: false,
//		MakeFSMFunc: func() raft.FSM {
//			return scheduler0Store.GetFSM()
//		},
//	})
//	defer cluster.Close()
//	cluster.FullyConnect()
//
//	scheduler0Store.UpdateRaft(cluster.Leader())
//
//	jobRepo := job_repo.NewJobRepo(logger, scheduler0RaftActions, scheduler0Store)
//	asyncTaskService := async_task.NewMockAsyncTaskService(t)
//	jobQueueRepo := job_queue_repo.NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)
//	jobExecutionsRepo := job_execution_repo.NewExecutionsRepo(
//		logger,
//		scheduler0RaftActions,
//		scheduler0Store,
//	)
//
//	dispatcher := utils.NewDispatcher(
//		ctx,
//		int64(1),
//		int64(1),
//	)
//
//	dispatcher.Run()
//
//	queueService := queue.NewJobQueue(ctx, logger, scheduler0config, scheduler0RaftActions, scheduler0Store, jobQueueRepo)
//	httpJobExecutor := executors.NewMockHTTPExecutor(t)
//
//	jobExecutorService := executor.NewJobExecutor(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0RaftActions,
//		jobRepo,
//		jobExecutionsRepo,
//		jobQueueRepo,
//		httpJobExecutor,
//		dispatcher,
//	)
//
//	leaderRaftAddress := strings.Split(cluster.Leader().String(), " ")[2]
//	os.Setenv("SCHEDULER0_RAFT_ADDRESS", "localhost:34410")
//	defer os.Unsetenv("SCHEDULER0_RAFT_ADDRESS")
//
//	followerHTTPAddresses := []string{
//		"http://localhost:34416",
//	}
//
//	followerRaftAddresses := []string{
//		strings.Split(cluster.Followers()[0].String(), " ")[2],
//	}
//	leaderHttpAddress := "http://localhost:34410"
//	nodeHTTPClient := NewMockNodeClient(t)
//	os.Setenv("SCHEDULER0_REPLICAS", fmt.Sprintf("["+
//		"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":2}, "+
//		"{\"raft_address\":\"%v\", \"address\":\"%v\", \"nodeId\":1}]",
//		followerRaftAddresses[0], followerHTTPAddresses[0],
//		leaderRaftAddress, leaderHttpAddress,
//	))
//	defer os.Unsetenv("SCHEDULER0_REPLICAS")
//
//	jobProcessor := processor.NewMockJobProcessorService(t)
//	nodeService := NewNode(
//		ctx,
//		logger,
//		scheduler0config,
//		scheduler0Secrets,
//		scheduler0Store,
//		scheduler0RaftActions,
//		jobExecutorService,
//		queueService,
//		jobProcessor,
//		jobRepo,
//		sharedRepo,
//		jobExecutionsRepo,
//		asyncTaskService,
//		dispatcher,
//		nodeHTTPClient,
//		nil,
//	)
//
//	executionLogs := []models.JobExecutionLog{
//		{
//			UniqueId: "some-unique-id",
//		},
//	}
//	sharedRepo.On("GetExecutionLogs", mock.Anything, false).Return(executionLogs, nil)
//	asyncTasks := []models.AsyncTask{
//		{
//			RequestId: "some-request-id",
//		},
//	}
//	asyncTaskService.On("GetUnCommittedTasks").Return(asyncTasks, nil)
//
//	sharedRepo.On("InsertExecutionLogs", mock.Anything, false, executionLogs).Return(nil)
//	sharedRepo.On("InsertAsyncTasksLogs", mock.Anything, false, asyncTasks).Return(nil)
//
//	nodeService.FsmStore = scheduler0Store
//
//	dir, err := os.Getwd()
//	if err != nil {
//		t.Fatalf(fmt.Sprintf("fatal error getting working dir: %s \n", err))
//	}
//
//	err = os.Mkdir(fmt.Sprintf("%v/sqlite_data", dir), os.ModePerm)
//	if err != nil {
//		t.Fatalf("failed to create sqlite_data dir %v", err)
//	}
//
//	nodeService.ConnectRaftLogsAndTransport()
//	nodeService.recoverRaftState()
//}

func Test_StopJobs(t *testing.T) {
	t.Run("StopJobs", func(t *testing.T) {
		t.Cleanup(func() {
			err := os.RemoveAll("./raft_data")
			if err != nil {
				fmt.Println("failed to remove raft_data dir for test", err)
			}
			err = os.RemoveAll("./sqlite_data")
			if err != nil {
				fmt.Println("failed to remove sqlite_data dir for test", err)
			}
		})
		ctx, cancelr := context.WithCancel(context.Background())
		defer cancelr()
		logger := hclog.New(&hclog.LoggerOptions{
			Name:  "node-service-test",
			Level: hclog.LevelFromString("trace"),
		})

		// Create a temporary SQLite database file
		tempFile, err := ioutil.TempFile("", "test-db")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		// Create a new SQLite database connection
		sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
		sqliteDb.RunMigration()
		sqliteDb.OpenConnectionToExistingDB()

		scheduler0config := config.NewScheduler0Config()
		sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
		scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)

		scheduler0Secrets := secrets.NewScheduler0Secrets()
		scheduler0Store := fsm.NewFSMStore(logger, scheduler0RaftActions, scheduler0config, sqliteDb, nil, nil, nil, nil, sharedRepo)

		cluster := raft.MakeClusterCustom(t, &raft.MakeClusterOpts{
			Peers:          2,
			Bootstrap:      true,
			Conf:           raft.DefaultConfig(),
			ConfigStoreFSM: false,
			MakeFSMFunc: func() raft.FSM {
				return scheduler0Store.GetFSM()
			},
		})
		cluster.FullyConnect()

		clusterLeader := cluster.Leader()
		scheduler0Store.UpdateRaft(clusterLeader)

		jobRepo := job_repo.NewJobRepo(logger, scheduler0RaftActions, scheduler0Store)
		asyncTaskManagerRepo := async_task_repo.NewAsyncTasksRepo(ctx, logger, scheduler0RaftActions, scheduler0Store)
		asyncTaskService := async_task.NewAsyncTaskManager(ctx, logger, scheduler0Store, asyncTaskManagerRepo, scheduler0config)
		jobQueueRepo := job_queue_repo.NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)
		jobExecutionsRepo := job_execution_repo.NewExecutionsRepo(
			logger,
			scheduler0RaftActions,
			scheduler0Store,
		)
		dispatcher := utils.NewDispatcher(
			ctx,
			int64(1),
			int64(1),
		)

		dispatcher.Run()

		queueService := queue.NewJobQueue(ctx, logger, scheduler0config, scheduler0RaftActions, scheduler0Store, jobQueueRepo)
		jobExecutorService := executor.NewMockJobExecutorService(t)
		jobExecutorService.On("StopAll")
		nodeHTTPClient := NewMockNodeClient(t)
		jobProcessor := processor.NewMockJobProcessorService(t)
		nodeService := NewNode(
			ctx,
			logger,
			scheduler0config,
			scheduler0Secrets,
			scheduler0Store,
			scheduler0RaftActions,
			jobExecutorService,
			queueService,
			jobProcessor,
			jobRepo,
			sharedRepo,
			jobExecutionsRepo,
			asyncTaskService,
			dispatcher,
			nodeHTTPClient,
			nil,
			false,
		)
		nodeService.StopJobs()
	})
}

func Test_StartJobs(t *testing.T) {
	t.Run("StartJobs", func(t *testing.T) {
		t.Cleanup(func() {
			err := os.RemoveAll("./raft_data")
			if err != nil {
				fmt.Println("failed to remove raft_data dir for test", err)
			}
			err = os.RemoveAll("./sqlite_data")
			if err != nil {
				fmt.Println("failed to remove sqlite_data dir for test", err)
			}
		})
		ctx, cancelr := context.WithCancel(context.Background())
		defer cancelr()
		logger := hclog.New(&hclog.LoggerOptions{
			Name:  "node-service-test",
			Level: hclog.LevelFromString("trace"),
		})

		// Create a temporary SQLite database file
		tempFile, err := ioutil.TempFile("", "test-db")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		// Create a new SQLite database connection
		sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
		sqliteDb.RunMigration()
		sqliteDb.OpenConnectionToExistingDB()

		scheduler0config := config.NewScheduler0Config()
		sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
		scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)

		scheduler0Secrets := secrets.NewScheduler0Secrets()
		scheduler0Store := fsm.NewFSMStore(logger, scheduler0RaftActions, scheduler0config, sqliteDb, nil, nil, nil, nil, sharedRepo)

		cluster := raft.MakeClusterCustom(t, &raft.MakeClusterOpts{
			Peers:          2,
			Bootstrap:      true,
			Conf:           raft.DefaultConfig(),
			ConfigStoreFSM: false,
			MakeFSMFunc: func() raft.FSM {
				return scheduler0Store.GetFSM()
			},
		})
		cluster.FullyConnect()

		clusterLeader := cluster.Leader()
		scheduler0Store.UpdateRaft(clusterLeader)

		jobRepo := job_repo.NewJobRepo(logger, scheduler0RaftActions, scheduler0Store)
		asyncTaskManagerRepo := async_task_repo.NewAsyncTasksRepo(ctx, logger, scheduler0RaftActions, scheduler0Store)
		asyncTaskService := async_task.NewAsyncTaskManager(ctx, logger, scheduler0Store, asyncTaskManagerRepo, scheduler0config)
		jobQueueRepo := job_queue_repo.NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)
		jobExecutionsRepo := job_execution_repo.NewExecutionsRepo(
			logger,
			scheduler0RaftActions,
			scheduler0Store,
		)
		dispatcher := utils.NewDispatcher(
			ctx,
			int64(1),
			int64(1),
		)
		dispatcher.Run()
		queueService := queue.NewJobQueue(ctx, logger, scheduler0config, scheduler0RaftActions, scheduler0Store, jobQueueRepo)
		jobExecutorService := executor.NewMockJobExecutorService(t)
		nodeHTTPClient := NewMockNodeClient(t)
		jobProcessor := processor.NewMockJobProcessorService(t)
		jobProcessor.On("RecoverJobs")
		nodeService := NewNode(
			ctx,
			logger,
			scheduler0config,
			scheduler0Secrets,
			scheduler0Store,
			scheduler0RaftActions,
			jobExecutorService,
			queueService,
			jobProcessor,
			jobRepo,
			sharedRepo,
			jobExecutionsRepo,
			asyncTaskService,
			dispatcher,
			nodeHTTPClient,
			nil,
			false,
		)
		nodeService.StartJobs()
	})
}
