package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/moby/moby/client"
)

type ImageHandler struct {
	docker *DockClient
}

func NewImageHandler(docker *DockClient) *ImageHandler {
	return &ImageHandler{docker: docker}
}

type ImagePullRequest struct {
	Name string `json:"name"`
}

type ImagePullResponse struct {
	Status string `json:"status"`
	Image  string `json:"image"`
}

func (ih *ImageHandler) PullImage(w http.ResponseWriter, r *http.Request) {
	var req ImagePullRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reader, err := ih.docker.GetClient().ImagePull(r.Context(), req.Name, client.ImagePullOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()
	io.Copy(io.Discard, reader)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ImagePullResponse{Status: "pulled", Image: req.Name})
}
