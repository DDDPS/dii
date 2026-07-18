package exporter

import (
	"DII/pkg/provider"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type ComposeFile struct {
	Version  string               `yaml:"version"`
	Services map[string]*Service `yaml:"services"`
}

type Healthcheck struct {
	Test        []string `yaml:"test,omitempty"`
	Interval    string   `yaml:"interval,omitempty"`
	Timeout     string   `yaml:"timeout,omitempty"`
	Retries     int      `yaml:"retries,omitempty"`
	StartPeriod string   `yaml:"start_period,omitempty"`
}

type Service struct {
	Image         string            `yaml:"image,omitempty"`
	Platform      string            `yaml:"platform,omitempty"`
	ContainerName string            `yaml:"container_name,omitempty"`
	Restart       string            `yaml:"restart,omitempty"`
	User          string            `yaml:"user,omitempty"`
	WorkingDir    string            `yaml:"working_dir,omitempty"`
	StopSignal    string            `yaml:"stop_signal,omitempty"`
	Entrypoint    []string          `yaml:"entrypoint,omitempty"`
	Command       []string          `yaml:"command,omitempty"`
	Ports         []string          `yaml:"ports,omitempty"`
	Volumes       []string          `yaml:"volumes,omitempty"`
	Environment   map[string]string `yaml:"environment,omitempty"`
	Labels        map[string]string `yaml:"labels,omitempty"`
	Healthcheck   *Healthcheck      `yaml:"healthcheck,omitempty"`
}

func ExportCompose(im *provider.ImageMeta) (string, error) {
	if im == nil {
		return "", fmt.Errorf("imageMeta is nil")
	}

	imageName := "app"
	if len(im.RepoTags) > 0 {
		imageName = im.RepoTags[0]
	} else if im.ID != "" {
		imageName = im.ID
	}

	serviceName := getServiceName(im)

	service := &Service{
		Image: imageName,
	}

	if im.OS != "" && im.Architecture != "" {
		service.Platform = fmt.Sprintf("%s/%s", im.OS, im.Architecture)
	}

	if im.Config != nil {
		upperService := strings.ToUpper(serviceName)

		// container_name
		service.ContainerName = fmt.Sprintf("${%s_CONTAINER_NAME:-%s_container}", upperService, serviceName)

		// restart
		service.Restart = fmt.Sprintf("${%s_RESTART:-always}", upperService)

		// user
		if im.Config.User != "" {
			service.User = fmt.Sprintf("${%s_USER:-%s}", upperService, im.Config.User)
		}

		// working_dir
		if im.Config.WorkingDir != "" {
			service.WorkingDir = im.Config.WorkingDir
		}

		// stop_signal
		if im.Config.StopSignal != "" {
			service.StopSignal = im.Config.StopSignal
		}

		// healthcheck
		if im.Config.Healthcheck != nil {
			hc := im.Config.Healthcheck
			service.Healthcheck = &Healthcheck{
				Test: hc.Test,
			}
			if hc.Interval > 0 {
				service.Healthcheck.Interval = hc.Interval.String()
			}
			if hc.Timeout > 0 {
				service.Healthcheck.Timeout = hc.Timeout.String()
			}
			if hc.StartPeriod > 0 {
				service.Healthcheck.StartPeriod = hc.StartPeriod.String()
			}
			if hc.Retries > 0 {
				service.Healthcheck.Retries = hc.Retries
			}
		}

		// entrypoint
		if len(im.Config.Entrypoint) > 0 {
			service.Entrypoint = im.Config.Entrypoint
		}

		// command
		if len(im.Config.Cmd) > 0 {
			service.Command = im.Config.Cmd
		}

		// ports
		if len(im.Config.ExposedPorts) > 0 {
			var ports []string
			for p := range im.Config.ExposedPorts {
				ports = append(ports, p)
			}
			sort.Slice(ports, func(i, j int) bool {
				pi := parsePortNum(ports[i])
				pj := parsePortNum(ports[j])
				if pi != pj {
					return pi < pj
				}
				return ports[i] < ports[j]
			})

			for _, p := range ports {
				parts := strings.Split(p, "/")
				portNum := parts[0]
				proto := ""
				if len(parts) > 1 {
					proto = parts[1]
				}

				var portVal string
				if len(ports) == 1 {
					if proto != "" && strings.ToLower(proto) != "tcp" {
						portVal = fmt.Sprintf("${%s_PORT_%s:-%s}:%s/%s", upperService, strings.ToUpper(proto), portNum, portNum, proto)
					} else {
						portVal = fmt.Sprintf("${%s_PORT:-%s}:%s", upperService, portNum, portNum)
					}
				} else {
					if proto != "" && strings.ToLower(proto) != "tcp" {
						portVal = fmt.Sprintf("${%s_PORT_%s_%s:-%s}:%s/%s", upperService, portNum, strings.ToUpper(proto), portNum, portNum, proto)
					} else {
						portVal = fmt.Sprintf("${%s_PORT_%s:-%s}:%s", upperService, portNum, portNum, portNum)
					}
				}
				service.Ports = append(service.Ports, portVal)
			}
		}

		// volumes
		if len(im.Config.Volumes) > 0 {
			var vols []string
			for v := range im.Config.Volumes {
				vols = append(vols, v)
			}
			sort.Strings(vols)

			usedSegs := make(map[string]bool)
			for _, v := range vols {
				cleanSeg := cleanEnvVarName(getLastPathSegment(v))
				if cleanSeg == "" {
					cleanSeg = "DATA"
				}
				baseSeg := cleanSeg
				counter := 1
				for usedSegs[cleanSeg] {
					cleanSeg = fmt.Sprintf("%s_%d", baseSeg, counter)
					counter++
				}
				usedSegs[cleanSeg] = true

				volVal := fmt.Sprintf("${%s_VOL_%s:-./%s}:%s", upperService, cleanSeg, strings.ToLower(cleanSeg), v)
				service.Volumes = append(service.Volumes, volVal)
			}
		}

		// environment
		if len(im.Config.Env) > 0 {
			service.Environment = make(map[string]string)
			for _, envStr := range im.Config.Env {
				var key, val string
				if idx := strings.Index(envStr, "="); idx != -1 {
					key = envStr[:idx]
					val = envStr[idx+1:]
				} else {
					key = envStr
					val = ""
				}

				if isSystemEnv(key) {
					continue
				}

				service.Environment[key] = fmt.Sprintf("${%s:-%s}", key, val)
			}
			if len(service.Environment) == 0 {
				service.Environment = nil
			}
		}

		// labels
		if len(im.Config.Labels) > 0 {
			service.Labels = im.Config.Labels
		}
	}

	cf := &ComposeFile{
		Version: "3.8",
		Services: map[string]*Service{
			serviceName: service,
		},
	}

	out, err := yaml.Marshal(cf)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func getServiceName(im *provider.ImageMeta) string {
	if len(im.RepoTags) > 0 {
		tag := im.RepoTags[0]
		parts := strings.Split(tag, "/")
		namePart := parts[len(parts)-1]
		if idx := strings.Index(namePart, "@"); idx != -1 {
			namePart = namePart[:idx]
		}
		if idx := strings.Index(namePart, ":"); idx != -1 {
			namePart = namePart[:idx]
		}
		cleaned := cleanName(namePart)
		if cleaned != "" {
			return cleaned
		}
	}
	if im.ID != "" {
		id := strings.TrimPrefix(im.ID, "sha256:")
		if len(id) > 12 {
			id = id[:12]
		}
		cleaned := cleanName(id)
		if cleaned != "" {
			return "img-" + cleaned
		}
	}
	return "app"
}

func cleanName(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			sb.WriteRune(r)
		}
	}
	return strings.ToLower(sb.String())
}

func cleanEnvVarName(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}
	res := sb.String()
	for strings.Contains(res, "__") {
		res = strings.ReplaceAll(res, "__", "_")
	}
	res = strings.Trim(res, "_")
	return strings.ToUpper(res)
}

func getLastPathSegment(p string) string {
	p = strings.Trim(p, "/")
	parts := strings.Split(p, "/")
	if len(parts) == 0 || parts[len(parts)-1] == "" {
		return "data"
	}
	return parts[len(parts)-1]
}

func parsePortNum(portStr string) int {
	parts := strings.Split(portStr, "/")
	val, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}
	return val
}

func isSystemEnv(key string) bool {
	sysKeys := map[string]bool{
		// OS & Shell basics
		"PATH":            true,
		"HOME":            true,
		"HOSTNAME":        true,
		"TERM":            true,
		"PWD":             true,
		"LANG":            true,
		"LC_ALL":          true,
		"SHELL":           true,
		"SHLVL":           true,
		"USER":            true,
		"LOGNAME":         true,
		"DEBIAN_FRONTEND": true,

		// Language Runtime defaults (Build-time only / Read-only)
		"GOLANG_VERSION":       true,
		"GOPATH":               true,
		"GOROOT":               true,
		"NODE_VERSION":         true,
		"YARN_VERSION":         true,
		"NPM_CONFIG_LOGLEVEL":  true,
		"PYTHON_VERSION":       true,
		"PYTHON_PIP_VERSION":   true,
		"PYTHONUNBUFFERED":     true,
		"PYTHONDONTWRITEBYTECODE": true,
		"JAVA_HOME":            true,
		"JAVA_VERSION":         true,
		"RUBY_VERSION":         true,
		"RUBY_MAJOR":           true,
	}
	return sysKeys[strings.ToUpper(key)]
}
