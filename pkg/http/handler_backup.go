package http

import (
	"fmt"
	"net/http"
)

func (h *httpHandler) backupStartHandler(w http.ResponseWriter, r *http.Request) {
	go h.backuper.CreateBackup()

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Backup started")
}
