package traefik

import "fmt"

const prefix = "traefik"

func RouterRule(name string) string {
	return fmt.Sprintf("%s/http/routers/%s/rule", prefix, name)
}

func RouterEntrypoint(name string) string {
	return fmt.Sprintf("%s/http/routers/%s/entrypoints/0", prefix, name)
}

func RouterService(name string) string {
	return fmt.Sprintf("%s/http/routers/%s/service", prefix, name)
}

func ServiceURL(name string) string {
	return fmt.Sprintf("%s/http/services/%s/loadbalancer/servers/0/url", prefix, name)
}

func AllRouterKeys(name string) []string {
	return []string{
		RouterRule(name),
		RouterEntrypoint(name),
		RouterService(name),
		ServiceURL(name),
	}
}
