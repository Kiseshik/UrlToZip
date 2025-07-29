package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Kiseshik/UrlToZip.git/config"
	"github.com/Kiseshik/UrlToZip.git/service"
)

func TestArchiveFlow(t *testing.T) {
	cfg := config.LoadConfig()
	manager := service.NewTaskManager(cfg)

	server := httptest.NewServer(NewRouter(manager))
	defer server.Close()

	resp, err := http.Get(server.URL + "/task/create")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
	defer resp.Body.Close()

	var createResp struct {
		TaskID string `json:"task_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		t.Fatalf("invalid create response: %v", err)
	}

	if createResp.TaskID == "" {
		t.Fatal("empty task_id returned")
	}

	taskID := createResp.TaskID

	sampleUrls := []string{
		"https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf",
		"https://upload.wikimedia.org/wikipedia/commons/7/77/Delete_key1.jpg",
		"https://upload.wikimedia.org/wikipedia/commons/a/a0/Pierre-Person.jpg",
	}

	for _, u := range sampleUrls {
		addUrl := server.URL + "/task/add?task_id=" + url.QueryEscape(taskID) + "&url=" + url.QueryEscape(u)
		resp, err := http.Get(addUrl)
		if err != nil {
			t.Fatalf("failed to add file: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("non-200 from add: %v", resp.Status)
		}
	}

	var statusResp struct {
		Status     string            `json:"status"`
		Errors     map[string]string `json:"errors"`
		ArchiveURL string            `json:"archive_url"`
	}

	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		resp, err := http.Get(server.URL + "/task/status?task_id=" + taskID)
		if err != nil {
			t.Fatalf("failed to get status: %v", err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err := json.Unmarshal(body, &statusResp); err != nil {
			t.Fatalf("invalid status response: %v", err)
		}
		if statusResp.Status == "ready" {
			break
		}
	}

	if statusResp.Status != "ready" {
		t.Fatalf("task did not reach ready state: %v", statusResp.Status)
	}

	if statusResp.ArchiveURL == "" {
		t.Fatal("archive URL is empty in ready state")
	}

	resp, err = http.Get(statusResp.ArchiveURL)
	if err != nil {
		t.Fatalf("failed to download archive: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("non-200 on archive download: %v", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read archive: %v", err)
	}
	if len(data) < 100 {
		t.Fatal("archive too small, seems empty")
	}

	_ = os.WriteFile(filepath.Join(os.TempDir(), "test-output.zip"), data, 0644)
}
