package job_queue

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"scheduler0/pkg/config"
	"scheduler0/pkg/db"
	"scheduler0/pkg/fsm"
	"scheduler0/pkg/models"
	"scheduler0/pkg/shared_repo"
	"testing"
	"time"
)

func Test_JobQueuesRepo_GetLastJobQueueLogForNode(t *testing.T) {
	scheduler0config := config.NewScheduler0Config()
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "job-queues-repo-test",
		Level: hclog.LevelFromString("DEBUG"),
	})
	sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
	scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)
	tempFile, err := ioutil.TempFile("", "test-db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
	sqliteDb.RunMigration()
	sqliteDb.OpenConnectionToExistingDB()
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
	defer cluster.Close()
	cluster.FullyConnect()
	scheduler0Store.UpdateRaft(cluster.Leader())

	// Create a new JobQueuesRepo instance
	jobQueuesRepo := NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)

	// Create job queue logs for testing
	logs := []models.JobQueueLog{
		{
			Id:              1,
			NodeId:          1,
			LowerBoundJobId: 1,
			UpperBoundJobId: 10,
			Version:         2,
			DateCreated:     time.Now(),
		},
		{
			Id:              2,
			NodeId:          1,
			LowerBoundJobId: 11,
			UpperBoundJobId: 20,
			Version:         2,
			DateCreated:     time.Now(),
		},
	}

	// Insert the job queue logs into the database
	for _, log := range logs {
		ds := scheduler0Store.GetDataStore()
		insertErr := insertJobQueueLog(logger, ds, log)
		if insertErr != nil {
			t.Fatalf("Failed to insert job queue log: %v", insertErr)
		}
	}

	// Define the node ID and version to query
	nodeID := uint64(1)
	version := uint64(2)

	// Call the GetLastJobQueueLogForNode method
	result := jobQueuesRepo.GetLastJobQueueLogForNode(nodeID, version)

	// Assert the number of retrieved job queue logs
	expectedCount := 2
	assert.Equal(t, expectedCount, len(result))

	// Create maps to track the expected log IDs and the retrieved log IDs
	expectedLogs := make(map[uint64]models.JobQueueLog)
	retrievedLogs := make(map[uint64]models.JobQueueLog)

	// Populate the expectedLogs map with log IDs as keys and corresponding logs as values
	for _, log := range logs {
		expectedLogs[log.Id] = log
	}

	// Populate the retrievedLogs map with log IDs from the result
	for _, log := range result {
		retrievedLogs[log.Id] = log
	}

	// Assert that the expectedLogs and retrievedLogs maps have the same number of elements
	assert.Equal(t, len(expectedLogs), len(retrievedLogs))

	// Assert the correctness of the retrieved job queue logs by comparing their properties with the expected logs
	for id, expectedLog := range expectedLogs {
		retrievedLog, ok := retrievedLogs[id]
		assert.True(t, ok) // Ensure the log with the given ID is present in the retrievedLogs map

		assert.Equal(t, expectedLog.Id, retrievedLog.Id)
		assert.Equal(t, expectedLog.NodeId, retrievedLog.NodeId)
		assert.Equal(t, expectedLog.LowerBoundJobId, retrievedLog.LowerBoundJobId)
		assert.Equal(t, expectedLog.UpperBoundJobId, retrievedLog.UpperBoundJobId)
	}
}

func Test_JobQueuesRepo_GetLastVersion(t *testing.T) {
	scheduler0config := config.NewScheduler0Config()
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "job-queues-repo-test",
		Level: hclog.LevelFromString("DEBUG"),
	})
	sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
	scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)
	tempFile, err := ioutil.TempFile("", "test-db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
	sqliteDb.RunMigration()
	sqliteDb.OpenConnectionToExistingDB()
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
	defer cluster.Close()
	cluster.FullyConnect()
	scheduler0Store.UpdateRaft(cluster.Leader())

	// Create a new JobQueuesRepo instance
	jobQueuesRepo := NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)

	// Create job queue versions for testing
	versions := []uint64{1, 3, 2}

	// Insert the job queue versions into the database
	for _, version := range versions {
		insertErr := insertJobQueueVersion(logger, scheduler0Store.GetDataStore(), version)
		if insertErr != nil {
			t.Fatalf("Failed to insert job queue version: %v", insertErr)
		}
	}

	// Call the GetLastVersion method
	result := jobQueuesRepo.GetLastVersion()

	// Assert the retrieved last version
	expectedVersion := uint64(3)
	assert.Equal(t, expectedVersion, result)
}

func Test_IncrementQueueVersion(t *testing.T) {
	scheduler0config := config.NewScheduler0Config()
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "job-queues-repo-test",
		Level: hclog.LevelFromString("DEBUG"),
	})
	sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
	scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)
	tempFile, err := ioutil.TempFile("", "test-db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
	sqliteDb.RunMigration()
	sqliteDb.OpenConnectionToExistingDB()
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
	defer cluster.Close()
	cluster.FullyConnect()
	scheduler0Store.UpdateRaft(cluster.Leader())

	// Create a new JobQueuesRepo instance
	jobQueuesRepo := NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)

	// Create job queue versions for testing
	versions := []uint64{1, 3, 2}

	// Insert the job queue versions into the database
	for _, version := range versions {
		insertErr := insertJobQueueVersion(logger, scheduler0Store.GetDataStore(), version)
		if insertErr != nil {
			t.Fatalf("Failed to insert job queue version: %v", insertErr)
		}
	}

	// Call the GetLastVersion method
	result := jobQueuesRepo.GetLastVersion()

	// Assert the retrieved last version
	expectedVersion := uint64(3)
	assert.Equal(t, expectedVersion, result)

	jobQueuesRepo.IncrementQueueVersion(4)

	result = jobQueuesRepo.GetLastVersion()

	expectedVersion = uint64(4)
	assert.Equal(t, expectedVersion, result)
}

func Test_InsertJobQueueLogs(t *testing.T) {
	scheduler0config := config.NewScheduler0Config()
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "job-queues-repo-test",
		Level: hclog.LevelFromString("DEBUG"),
	})
	sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
	scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)
	tempFile, err := ioutil.TempFile("", "test-db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
	sqliteDb.RunMigration()
	sqliteDb.OpenConnectionToExistingDB()
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
	defer cluster.Close()
	cluster.FullyConnect()
	scheduler0Store.UpdateRaft(cluster.Leader())

	// Create a new JobQueuesRepo instance
	jobQueuesRepo := NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)

	// Create job queue versions for testing
	versions := []uint64{1}

	// Insert the job queue versions into the database
	for _, version := range versions {
		insertErr := insertJobQueueVersion(logger, scheduler0Store.GetDataStore(), version)
		if insertErr != nil {
			t.Fatalf("Failed to insert job queue version: %v", insertErr)
		}
	}

	jobQueuesRepo.InsertJobQueueLogs([]models.JobQueueLog{{
		NodeId:          1,
		LowerBoundJobId: 1,
		UpperBoundJobId: 20,
		Version:         1,
		DateCreated:     time.Now(),
	}, {
		NodeId:          2,
		LowerBoundJobId: 21,
		UpperBoundJobId: 29,
		Version:         1,
		DateCreated:     time.Now(),
	}})

	jobQueueLogsForNode1 := jobQueuesRepo.GetLastJobQueueLogForNode(1, 1)
	jobQueueLogsForNode2 := jobQueuesRepo.GetLastJobQueueLogForNode(2, 1)

	assert.Equal(t, uint64(1), jobQueueLogsForNode1[0].LowerBoundJobId)
	assert.Equal(t, uint64(20), jobQueueLogsForNode1[0].UpperBoundJobId)

	assert.Equal(t, uint64(21), jobQueueLogsForNode2[0].LowerBoundJobId)
	assert.Equal(t, uint64(29), jobQueueLogsForNode2[0].UpperBoundJobId)
}

func Test_GetJobQueueByLastInsertedAndRowsAffected(t *testing.T) {
	scheduler0config := config.NewScheduler0Config()
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "job-queues-repo-test",
		Level: hclog.LevelFromString("DEBUG"),
	})
	sharedRepo := shared_repo.NewSharedRepo(logger, scheduler0config)
	scheduler0RaftActions := fsm.NewScheduler0RaftActions(sharedRepo, nil)
	tempFile, err := ioutil.TempFile("", "test-db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	sqliteDb := db.NewSqliteDbConnection(logger, tempFile.Name())
	sqliteDb.RunMigration()
	sqliteDb.OpenConnectionToExistingDB()
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
	defer cluster.Close()
	cluster.FullyConnect()
	scheduler0Store.UpdateRaft(cluster.Leader())

	// Create a new JobQueuesRepo instance
	jobQueuesRepo := NewJobQueuesRepo(logger, scheduler0RaftActions, scheduler0Store)

	// Create job queue versions for testing
	versions := []uint64{1}

	// Insert the job queue versions into the database
	for _, version := range versions {
		insertErr := insertJobQueueVersion(logger, scheduler0Store.GetDataStore(), version)
		if insertErr != nil {
			t.Fatalf("Failed to insert job queue version: %v", insertErr)
		}
	}

	jobQueuesRepo.InsertJobQueueLogs([]models.JobQueueLog{{
		NodeId:          1,
		LowerBoundJobId: 1,
		UpperBoundJobId: 20,
		Version:         1,
		DateCreated:     time.Now(),
	}, {
		NodeId:          2,
		LowerBoundJobId: 21,
		UpperBoundJobId: 29,
		Version:         1,
		DateCreated:     time.Now(),
	}})

	jobQueueLogs := jobQueuesRepo.GetJobQueueByLastInsertedAndRowsAffected(2, 2)

	assert.Equal(t, 2, len(jobQueueLogs))
}

func insertJobQueueLog(logger hclog.Logger, dataStore db.DataStore, log models.JobQueueLog) error {
	insertBuilder := sq.Insert(JobQueuesTableName).
		Columns(
			JobQueueIdColumn,
			JobQueueNodeIdColumn,
			JobQueueLowerBoundJobId,
			JobQueueUpperBound,
			JobQueueVersion,
			JobQueueDateCreatedColumn,
		).
		Values(
			log.Id,
			log.NodeId,
			log.LowerBoundJobId,
			log.UpperBoundJobId,
			log.Version,
			log.DateCreated,
		).
		RunWith(dataStore.GetOpenConnection())

	_, err := insertBuilder.Exec()
	if err != nil {
		logger.Error("failed to insert job queue log", err)
		return err
	}

	return nil
}

func insertJobQueueVersion(logger hclog.Logger, dataStore db.DataStore, version uint64) error {
	insertBuilder := sq.Insert(JobQueuesVersionTableName).
		Columns(JobQueueVersion).
		Columns(JobNumberOfActiveNodesVersion).
		Values(version, 1).
		RunWith(dataStore.GetOpenConnection())

	_, err := insertBuilder.Exec()
	if err != nil {
		logger.Error("failed to insert job queue version", err)
		return err
	}

	return nil
}
