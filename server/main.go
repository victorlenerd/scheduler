package main

import (
	"cron-server/server/controllers"
	"cron-server/server/middlewares"
	"cron-server/server/migrations"
	"cron-server/server/misc"
	"cron-server/server/models"
	"cron-server/server/process"
	"github.com/gorilla/mux"
	"github.com/unrolled/secure"
	"log"
	"net/http"
)

func main() {
	pool, err := migrations.NewPool(migrations.CreateConnection, migrations.MaxConnections)
	misc.CheckErr(err)

	// Setup logging
	log.SetFlags(0)
	log.SetOutput(new(misc.LogWriter))

	// Set time zone, create database and run migrations
	models.Setup(pool)

	// Start process to execute cron-server jobs
	go process.Start(pool)

	// HTTP router setup
	router := mux.NewRouter()

	// Security middleware
	secureMiddleware := secure.New(secure.Options{FrameDeny: true})

	// Initialize controllers
	executionController := controllers.ExecutionController{Pool: *pool}
	jobController := controllers.JobController{Pool: *pool}
	projectController := controllers.ProjectController{Pool: *pool}
	credentialController := controllers.CredentialController{Pool: *pool}

	// Mount middleware
	middleware := middlewares.MiddlewareType{}

	router.Use(secureMiddleware.Handler)
	router.Use(mux.CORSMethodMiddleware(router))
	router.Use(middleware.ContextMiddleware)
	router.Use(middleware.AuthMiddleware(pool))

	// Executions Endpoint
	router.HandleFunc("/executions", executionController.GetAll).Methods(http.MethodGet)
	router.HandleFunc("/executions/{id}", executionController.GetOne).Methods(http.MethodGet)

	// Credentials Endpoint
	router.HandleFunc("/credentials", credentialController.CreateOne).Methods(http.MethodPost)
	router.HandleFunc("/credentials", credentialController.List).Methods(http.MethodGet)
	router.HandleFunc("/credentials/{id}", credentialController.GetOne).Methods(http.MethodGet)
	router.HandleFunc("/credentials/{id}", credentialController.UpdateOne).Methods(http.MethodPut)
	router.HandleFunc("/credentials/{id}", credentialController.DeleteOne).Methods(http.MethodDelete)

	// Job Endpoint
	router.HandleFunc("/jobs", jobController.GetAllOrCreateOne).Methods(http.MethodPost, http.MethodGet)
	router.HandleFunc("/jobs/{id}", jobController.GetOne).Methods(http.MethodGet)
	router.HandleFunc("/jobs/{id}", jobController.UpdateOne).Methods(http.MethodPut)
	router.HandleFunc("/jobs/{id}", jobController.DeleteOne).Methods(http.MethodDelete)

	// Projects Endpoint
	router.HandleFunc("/projects", projectController.GetAllOrCreateOne).Methods(http.MethodPost, http.MethodGet)
	router.HandleFunc("/projects/{id}", projectController.GetOne).Methods(http.MethodGet)
	router.HandleFunc("/projects/{id}", projectController.UpdateOne).Methods(http.MethodPut)
	router.HandleFunc("/projects/{id}", projectController.DeleteOne).Methods(http.MethodDelete)

	log.Println("Server is running on port", misc.GetPort(), misc.GetClientHost())
	err = http.ListenAndServe(misc.GetPort(), router)
	misc.CheckErr(err)
}
