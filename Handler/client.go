package handler

import "github.com/moby/moby/client"

type DockClient struct {
	Client *client.Client
}

func InitDockClient() (*DockClient, error) {
	apiClient, err := client.New(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &DockClient{Client: apiClient}, nil
}

func (d *DockClient) Close() {
	d.Client.Close()
}

func (d *DockClient) GetClient() *client.Client {
	return d.Client
}
