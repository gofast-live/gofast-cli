package clients

import "github.com/gofast-live/gofast-cli/v2/cmd/gof/config"

const (
	Svelte   = "svelte"
	Tanstack = "tanstack"
)

type Spec struct {
	Name                 string
	DisplayName          string
	ServiceDir           string
	ComposeFile          string
	Port                 string
	PaymentsRouteSubpath string
	FilesRouteSubpath    string
	EmailsRouteSubpath   string
}

var specs = map[string]Spec{
	Svelte: {
		Name:                 Svelte,
		DisplayName:          "Svelte",
		ServiceDir:           "service-svelte",
		ComposeFile:          "docker-compose.svelte.yml",
		Port:                 "3000",
		PaymentsRouteSubpath: "src/routes/(app)/payments",
		FilesRouteSubpath:    "src/routes/(app)/files",
		EmailsRouteSubpath:   "src/routes/(app)/emails",
	},
	Tanstack: {
		Name:                 Tanstack,
		DisplayName:          "TanStack",
		ServiceDir:           "service-tanstack",
		ComposeFile:          "docker-compose.tanstack.yml",
		Port:                 "3000",
		PaymentsRouteSubpath: "src/routes/_layout/payments",
		FilesRouteSubpath:    "src/routes/_layout/files.tsx",
		EmailsRouteSubpath:   "src/routes/_layout/emails.tsx",
	},
}

func SpecFor(name string) (Spec, bool) {
	spec, ok := specs[name]
	return spec, ok
}

func All() []Spec {
	return []Spec{
		specs[Svelte],
		specs[Tanstack],
	}
}

func Enabled(cfg *config.Config) []Spec {
	if cfg == nil {
		return nil
	}

	var enabled []Spec
	for _, service := range cfg.Services {
		spec, ok := SpecFor(service.Name)
		if ok {
			enabled = append(enabled, spec)
		}
	}
	return enabled
}

func HasAny(cfg *config.Config) bool {
	return len(Enabled(cfg)) > 0
}
