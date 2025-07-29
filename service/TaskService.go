package service

import (
	"archive/zip"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Kiseshik/UrlToZip.git/config"
	"github.com/google/uuid"
)

type TaskStatus string

const (
	Pending TaskStatus = "pending"
	Ready   TaskStatus = "ready"
	Failed  TaskStatus = "failed"
)

type Task struct {
	ID       string
	Files    []string
	Status   TaskStatus
	Archive  string
	Errors   map[string]string
	Mu       sync.Mutex
	doneChan chan struct{}
}

type TaskManager struct {
	Tasks         map[string]*Task
	Config        config.Config
	TaskMutex     sync.Mutex
	ProcessingSem chan struct{}
}

func NewTaskManager(cfg config.Config) *TaskManager {
	return &TaskManager{
		Tasks:         make(map[string]*Task),
		Config:        cfg,
		ProcessingSem: make(chan struct{}, cfg.MaxActiveTasks),
	}
}

func (tm *TaskManager) CreateTask() (*Task, error) {
	tm.TaskMutex.Lock()
	defer tm.TaskMutex.Unlock()

	if len(tm.ProcessingSem) >= tm.Config.MaxActiveTasks {
		return nil, errors.New("server is busy, try again later")
	}

	id := uuid.New().String()
	task := &Task{
		ID:       id,
		Status:   Pending,
		Errors:   make(map[string]string),
		doneChan: make(chan struct{}),
	}
	tm.Tasks[id] = task
	return task, nil
}

func (tm *TaskManager) AddFileToTask(taskID string, url string) error {
	task, ok := tm.Tasks[taskID]
	if !ok {
		return errors.New("task not found")
	}

	task.Mu.Lock()
	defer task.Mu.Unlock()

	if task.Status != Pending {
		return errors.New("task is not accepting files")
	}
	if len(task.Files) >= tm.Config.MaxFilesPerTask {
		return errors.New("file limit reached")
	}

	ext := filepath.Ext(url)
	if !tm.Config.AllowedExtensions[strings.ToLower(ext)] {
		return errors.New("file type not allowed")
	}

	task.Files = append(task.Files, url)

	if len(task.Files) == tm.Config.MaxFilesPerTask {
		go tm.processTask(task)
	}

	return nil
}

func (tm *TaskManager) processTask(task *Task) {
	tm.ProcessingSem <- struct{}{}
	defer func() { <-tm.ProcessingSem }()

	dir := os.TempDir()
	zipPath := filepath.Join(dir, task.ID+".zip")

	zipFile, err := os.Create(zipPath)
	if err != nil {
		task.Status = Failed
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, fileURL := range task.Files {
		resp, err := http.Get(fileURL)
		if err != nil || resp.StatusCode != 200 {
			task.Errors[fileURL] = "failed to download"
			continue
		}
		defer resp.Body.Close()

		filename := filepath.Base(fileURL)
		writer, err := zipWriter.Create(filename)
		if err != nil {
			task.Errors[fileURL] = "failed to add to archive"
			continue
		}
		if _, err := io.Copy(writer, resp.Body); err != nil {
			task.Errors[fileURL] = "failed to copy to archive"
			continue
		}
	}

	task.Mu.Lock()
	if len(task.Errors) == len(task.Files) {
		task.Status = Failed
	} else {
		task.Status = Ready
		task.Archive = zipPath
	}
	task.Mu.Unlock()

	close(task.doneChan)
}
