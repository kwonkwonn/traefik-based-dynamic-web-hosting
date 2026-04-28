package handler

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"gopkg.in/yaml.v3"
)

// ContainerSpec is the YAML schema for defining a container.
//
// Example:
//
//	name: my-app
//	image: nginx:latest
//	ports:
//	  - "8080:80"
//	volumes:
//	  - "/data:/data"
//	environment:
//	  - "ENV=production"
//	labels:
//	  traefik.enable: "true"
//	restart: unless-stopped
type ContainerSpec struct {
	Name        string            `yaml:"name"`
	Image       string            `yaml:"image"`
	Command     []string          `yaml:"command"`
	Ports       []string          `yaml:"ports"`
	Volumes     []string          `yaml:"volumes"`
	Environment []string          `yaml:"environment"`
	Labels      map[string]string `yaml:"labels"`
	Restart     string            `yaml:"restart"`
	NetworkMode string            `yaml:"network_mode"`
}

type ContainerHandler struct {
	docker *DockClient
}

func NewContainerHandler(docker *DockClient) *ContainerHandler {
	return &ContainerHandler{docker: docker}
}

func LoadContainerSpec(path string) (*ContainerSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	var spec ContainerSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	if spec.Image == "" {
		return nil, fmt.Errorf("image is required")
	}
	return &spec, nil
}

func (ch *ContainerHandler) CreateFromSpec(spec *ContainerSpec) (string, error) {
	cfg := &container.Config{
		Image:  spec.Image,
		Cmd:    spec.Command,
		Env:    spec.Environment,
		Labels: spec.Labels,
	}

	hostCfg := &container.HostConfig{
		Binds:         spec.Volumes,
		PortBindings:  parsePortBindings(spec.Ports),
		RestartPolicy: parseRestartPolicy(spec.Restart),
	}
	if spec.NetworkMode != "" {
		hostCfg.NetworkMode = container.NetworkMode(spec.NetworkMode)
	}

	ctx := context.Background()
	result, err := ch.docker.GetClient().ContainerCreate(ctx, client.ContainerCreateOptions{
		Name:       spec.Name,
		Config:     cfg,
		HostConfig: hostCfg,
		NetworkingConfig: &network.NetworkingConfig{},
	})
	if err != nil {
		return "", fmt.Errorf("container create: %w", err)
	}
	return result.ID, nil
}

func (ch *ContainerHandler) StartContainer(id string) error {
	ctx := context.Background()
	_, err := ch.docker.GetClient().ContainerStart(ctx, id, client.ContainerStartOptions{})
	return err
}

func (ch *ContainerHandler) RunFromSpec(spec *ContainerSpec) (string, error) {
	id, err := ch.CreateFromSpec(spec)
	if err != nil {
		return "", err
	}
	if err := ch.StartContainer(id); err != nil {
		return "", fmt.Errorf("container start: %w", err)
	}
	return id, nil
}

func (ch *ContainerHandler) RunFromFile(path string) (string, error) {
	spec, err := LoadContainerSpec(path)
	if err != nil {
		return "", err
	}
	return ch.RunFromSpec(spec)
}

// parsePortBindings converts ["8080:80", "443:443/tcp"] to PortMap.
func parsePortBindings(ports []string) network.PortMap {
	pm := make(network.PortMap)
	for _, p := range ports {
		hostPort, containerPort, ok := strings.Cut(p, ":")
		if !ok {
			continue
		}
		proto := "tcp"
		if s, pr, found := strings.Cut(containerPort, "/"); found {
			containerPort = s
			proto = pr
		}
		port, err := network.ParsePort(containerPort + "/" + proto)
		if err != nil {
			continue
		}
		pm[port] = append(pm[port], network.PortBinding{HostPort: hostPort})
	}
	return pm
}

func parseRestartPolicy(policy string) container.RestartPolicy {
	switch policy {
	case "always":
		return container.RestartPolicy{Name: container.RestartPolicyAlways}
	case "on-failure":
		return container.RestartPolicy{Name: container.RestartPolicyOnFailure}
	case "unless-stopped":
		return container.RestartPolicy{Name: container.RestartPolicyUnlessStopped}
	default:
		return container.RestartPolicy{Name: container.RestartPolicyDisabled}
	}
}
