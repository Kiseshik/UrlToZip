package handler

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/Kiseshik/UrlToZip.git/service"
)

func CreateTaskHandler(manager *service.TaskManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := manager.CreateTask()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{
			"task_id": t.ID,
		})
	}
}

func AddFileToTaskHandler(manager *service.TaskManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := r.URL.Query().Get("task_id")
		url := r.URL.Query().Get("url")

		if taskID == "" || url == "" {
			http.Error(w, "missing task_id or url", http.StatusBadRequest)
			return
		}

		err := manager.AddFileToTask(taskID, url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func GetTaskStatusHandler(manager *service.TaskManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := r.URL.Query().Get("task_id")
		if taskID == "" {
			http.Error(w, "missing task_id", http.StatusBadRequest)
			return
		}

		t, ok := manager.Tasks[taskID]
		if !ok {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}

		t.Mu.Lock()
		defer t.Mu.Unlock()

		resp := map[string]interface{}{
			"status": t.Status,
			"errors": t.Errors,
		}
		if t.Status == service.Ready {
			resp["archive_url"] = "http://" + r.Host + "/task/download?task_id=" + t.ID
		}
		json.NewEncoder(w).Encode(resp)
	}
}

func DownloadArchiveHandler(manager *service.TaskManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := r.URL.Query().Get("task_id")
		if taskID == "" {
			http.Error(w, "missing task_id", http.StatusBadRequest)
			return
		}

		t, ok := manager.Tasks[taskID]
		if !ok {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}

		t.Mu.Lock()
		defer t.Mu.Unlock()

		if t.Status != service.Ready || t.Archive == "" {
			http.Error(w, "archive not ready", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(t.Archive)+"\"")
		http.ServeFile(w, r, t.Archive)
	}
}
