package main

import (
	handler "traefik-based-dynamic-web-hosting/Handler"
	"traefik-based-dynamic-web-hosting/server"
)

func main() {
	dockerClient, err := handler.InitDockClient()
	if err != nil {
		panic(err)
	}
	defer dockerClient.Close()

	go server.StartServer(dockerClient)

	select {}
}
