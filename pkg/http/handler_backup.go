package http

import "net/http"

func (h *httpHandler) backupStartHandler(w http.ResponseWriter, r *http.Request) {

	http.Error(w, "Not implmented", 500)
}
