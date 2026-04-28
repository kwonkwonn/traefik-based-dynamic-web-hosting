package server

import (
	"log"
	"net/http"

	handler "traefik-based-dynamic-web-hosting/Handler"
)

func newMux(docker *handler.DockClient) *http.ServeMux {
	mux := http.NewServeMux()

	imageHandler := handler.NewImageHandler(docker)
	mux.HandleFunc("POST /images/pull", imageHandler.PullImage)

	return mux
}

func StartServer(docker *handler.DockClient) {
	mux := newMux(docker)

	log.Println("server listening on :8090")
	if err := http.ListenAndServe(":8090", mux); err != nil {
		log.Fatal(err)
	}
}
