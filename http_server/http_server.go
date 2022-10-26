package http_server

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/unrolled/secure"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"os"
	"scheduler0/constants"
	"scheduler0/db"
	"scheduler0/fsm"
	controllers2 "scheduler0/http_server/controllers"
	"scheduler0/http_server/middlewares"
	"scheduler0/job_executor"
	"scheduler0/job_queue"
	"scheduler0/models"
	"scheduler0/peer"
	"scheduler0/process"
	repository2 "scheduler0/repository"
	service2 "scheduler0/service"
	"scheduler0/utils"
	"time"
)

// Start this will start the http server
func Start() {
	ctx := context.Background()
	logger := log.New(os.Stderr, "[http-server] ", log.LstdFlags)

	dir, err := os.Getwd()
	if err != nil {
		logger.Fatalln(fmt.Errorf("Fatal error getting working dir: %s \n", err))
	}
	dbFilePath := fmt.Sprintf("%v/%v", dir, constants.SqliteDbFileName)

	sqliteDb := db.NewSqliteDbConnection(dbFilePath)
	conn, err := sqliteDb.OpenConnection()
	if err != nil {
		logger.Fatal("Failed to open connection", err)
	}

	dbConnection := conn.(*sql.DB)
	configs := utils.GetScheduler0Configurations()

	err = dbConnection.Ping()
	if err != nil {
		logger.Fatalln(fmt.Errorf("ping error: restore failed to create db: %v", err))
	}

	utils.MakeDirIfNotExist(constants.RaftDir)

	dirPath := fmt.Sprintf("%v/%v", constants.RaftDir, configs.NodeId)
	dirPath, exists := utils.MakeDirIfNotExist(dirPath)

	p := peer.NewPeer(logger)
	fsmStr := fsm.NewFSMStore(sqliteDb, dbConnection, logger)

	if exists {
		p.RecoverPeer()
	}

	rft := p.NewRaft(fsmStr)
	fsmStr.Raft = rft

	if configs.Bootstrap == "true" && !exists {
		p.BootstrapNode(rft)
	}

	//repository
	credentialRepo := repository2.NewCredentialRepo(fsmStr)
	jobRepo := repository2.NewJobRepo(fsmStr)
	executionRepo := repository2.NewExecutionRepo(fsmStr)
	projectRepo := repository2.NewProjectRepo(fsmStr, jobRepo)

	jobExecutor := job_executor.NewJobExecutor(jobRepo)
	jobQueue := job_queue.NewJobQueue(rft, jobExecutor)

	//services
	credentialService := service2.NewCredentialService(credentialRepo, ctx)
	jobService := service2.NewJobService(jobRepo, jobQueue, ctx)
	executionService := service2.NewExecutionService(executionRepo)
	projectService := service2.NewProjectService(projectRepo)

	go func() {
		ticker := time.NewTicker(time.Millisecond * 100)
		startedJobs := false
		for !startedJobs {
			select {
			case <-ticker.C:
				vErr := rft.VerifyLeader()
				if vErr.Error() == nil {
					jobProcessor := process.NewJobProcessor(jobRepo, projectRepo, jobQueue)
					// StartJobs process to execute cron-server jobs
					logger.Println("starting Jobs")
					go jobProcessor.StartJobs()
					startedJobs = true
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case pendingJob := <-fsmStr.PendingJobs:
				logger.Println("running job:", pendingJob.ID)
				jobExecutor.Run([]models.JobModel{pendingJob})
			}
		}
	}()

	go jobExecutor.ListenToChannelsUpdates()

	// HTTP router setup
	router := mux.NewRouter()

	// Security middleware
	secureMiddleware := secure.New(secure.Options{FrameDeny: true})

	// Initialize controllers
	executionController := controllers2.NewExecutionsController(executionService)
	jobController := controllers2.NewJoBHTTPController(jobService)
	projectController := controllers2.NewProjectController(projectService)
	credentialController := controllers2.NewCredentialController(credentialService)
	healthCheckController := controllers2.NewHealthCheckController(rft)

	// Mount middleware
	middleware := middlewares.MiddlewareType{}

	router.Use(secureMiddleware.Handler)
	router.Use(mux.CORSMethodMiddleware(router))
	router.Use(middleware.ContextMiddleware)
	router.Use(middleware.AuthMiddleware(credentialService))

	// Executions Endpoint
	router.HandleFunc("/executions", executionController.ListExecutions).Methods(http.MethodGet)

	// Credentials Endpoint
	router.HandleFunc("/credentials", credentialController.CreateOneCredential).Methods(http.MethodPost)
	router.HandleFunc("/credentials", credentialController.ListCredentials).Methods(http.MethodGet)
	router.HandleFunc("/credentials/{uuid}", credentialController.GetOneCredential).Methods(http.MethodGet)
	router.HandleFunc("/credentials/{uuid}", credentialController.UpdateOneCredential).Methods(http.MethodPut)
	router.HandleFunc("/credentials/{uuid}", credentialController.DeleteOneCredential).Methods(http.MethodDelete)

	// Job Endpoint
	router.HandleFunc("/job", jobController.CreateOneJob).Methods(http.MethodPost)
	router.HandleFunc("/jobs", jobController.BatchCreateJobs).Methods(http.MethodPost)
	router.HandleFunc("/jobs", jobController.ListJobs).Methods(http.MethodGet)
	router.HandleFunc("/jobs/{uuid}", jobController.GetOneJob).Methods(http.MethodGet)
	router.HandleFunc("/jobs/{uuid}", jobController.UpdateOneJob).Methods(http.MethodPut)
	router.HandleFunc("/jobs/{uuid}", jobController.DeleteOneJob).Methods(http.MethodDelete)

	// Projects Endpoint
	router.HandleFunc("/projects", projectController.CreateOneProject).Methods(http.MethodPost)
	router.HandleFunc("/projects", projectController.ListProjects).Methods(http.MethodGet)
	router.HandleFunc("/projects/{uuid}", projectController.GetOneProject).Methods(http.MethodGet)
	router.HandleFunc("/projects/{uuid}", projectController.UpdateOneProject).Methods(http.MethodPut)
	router.HandleFunc("/projects/{uuid}", projectController.DeleteOneProject).Methods(http.MethodDelete)

	// Healthcheck Endpoint
	router.HandleFunc("/healthcheck", healthCheckController.HealthCheck).Methods(http.MethodGet)

	router.PathPrefix("/api-docs/").Handler(http.StripPrefix("/api-docs/", http.FileServer(http.Dir("./server/http_server/api-docs/"))))

	log.Println("Server is running on port", configs.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", configs.Port), router)
	if err != nil {
		logger.Fatal("failed to start http-server", err)
	}
}
