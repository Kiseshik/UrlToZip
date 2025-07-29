package main

import (
	"log"
	"net/http"

	"github.com/Kiseshik/UrlToZip.git/config"
	"github.com/Kiseshik/UrlToZip.git/handler"
	"github.com/Kiseshik/UrlToZip.git/service"
)

func NewRouter(taskManager *service.TaskManager) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/task/create", handler.CreateTaskHandler(taskManager))
	mux.HandleFunc("/task/add", handler.AddFileToTaskHandler(taskManager))
	mux.HandleFunc("/task/status", handler.GetTaskStatusHandler(taskManager))
	mux.HandleFunc("/task/download", handler.DownloadArchiveHandler(taskManager))
	return mux
}

func main() {
	config := config.LoadConfig()
	taskManager := service.NewTaskManager(config)
	NewRouter(taskManager)

	log.Printf("Server is running on port %s\n", config.Port)
	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
		log.Fatal(err)
	}
}
