//Package metadata is used for metadata server actions
package metadata

import (
	"net/http"

	"github.com/Advaithdp02/s3lite/internal/storage"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func UploadHandler(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// TODO:
		// Read uploaded file
		// Save temporarily
		// store.Upload(...)

		w.Write([]byte("Upload endpoint"))
	}
}

func DownloadHandler(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// TODO:
		// Read filename
		// store.Download(...)

		w.Write([]byte("Download endpoint"))
	}
}

func ListHandler(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// TODO:
		// objects := store.List()

		w.Write([]byte("List endpoint"))
	}
}

func StatHandler(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// TODO:
		// store.Stat()

		w.Write([]byte("Stat endpoint"))
	}
}

func DeleteHandler(store *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// TODO:
		// store.Delete()

		w.Write([]byte("Delete endpoint"))
	}
}
