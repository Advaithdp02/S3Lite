// Package metadata is used for metadata server actions
package metadata

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Advaithdp02/s3lite/internal/storage"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func UploadHandler(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		filename := filepath.Base(header.Filename)
		if filename == "." || filename == "/" {
			http.Error(w, "invalid filename", http.StatusBadRequest)
			return
		}

		tmp, err := os.CreateTemp("", "upload-*")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(tmp.Name())

		if _, err := io.Copy(tmp, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmp.Close()

		finalPath := filepath.Join(filepath.Dir(tmp.Name()), filename)
		if err := os.Rename(tmp.Name(), finalPath); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(finalPath)

		if err := store.Upload(finalPath); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("uploaded"))
	}
}

func DownloadHandler(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		filename := r.URL.Query().Get("file")
		if filename == "" {
			http.Error(w, "missing file", http.StatusBadRequest)
			return
		}

		tmpDir, err := os.MkdirTemp("", "download-*")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		go func() {
			time.Sleep(5 * time.Second)
			os.RemoveAll(tmpDir)
		}()

		if err := store.Download(filename, tmpDir); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		path := filepath.Join(tmpDir, filename)

		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		http.ServeFile(w, r, path)
	}
}

func ListHandler(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		objects, err := store.List()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(objects); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func StatHandler(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		filename := r.URL.Query().Get("file")
		if filename == "" {
			http.Error(w, "missing file parameter", http.StatusBadRequest)
			return
		}

		meta, err := store.Stat(filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(meta)
	}
}

func DeleteHandler(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		filename := r.URL.Query().Get("file")
		if filename == "" {
			http.Error(w, "missing file parameter", http.StatusBadRequest)
			return
		}

		if err := store.Delete(filename); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("deleted"))
	}
}
