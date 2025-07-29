package main

import (
	"log"
	"net/http"

	"github.com/Kiseshik/UrlToZip.git/config"
	"github.com/Kiseshik/UrlToZip.git/handler"
	"github.com/Kiseshik/UrlToZip.git/service"
)

func main() {
	config := config.LoadConfig()
	taskManager := service.NewTaskManager(config)

	http.HandleFunc("/task/create", handler.CreateTaskHandler(taskManager))
	http.HandleFunc("/task/add", handler.AddFileToTaskHandler(taskManager))
	http.HandleFunc("/task/status", handler.GetTaskStatusHandler(taskManager))
	http.HandleFunc("/task/download", handler.DownloadArchiveHandler(taskManager))

	log.Printf("Server is running on port %s\n", config.Port)
	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
		log.Fatal(err)
	}
}
