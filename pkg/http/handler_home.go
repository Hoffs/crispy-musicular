package http

import "net/http"

func (h *httpHandler) homeHandler(w http.ResponseWriter, r *http.Request) {
	st, err := h.auth.GetState()
	// TODO: figure out something to handle these errors easier
	if err != nil {
		h.renderError(w, "No state found", err)
	}

	d := struct {
		User string
	}{
		User: st.User,
	}
	h.t.renderTemplate(w, "home.tmpl", &d)
}
