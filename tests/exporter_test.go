package tests

import (
	"encoding/json"
	"strings"
	"testing"

	"DII/pkg/exporter"
	"DII/pkg/provider"

	v1 "github.com/moby/docker-image-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"gopkg.in/yaml.v3"
)

func createTestImageMeta() *provider.ImageMeta {
	return &provider.ImageMeta{
		ID:           "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		RepoTags:     []string{"test-image:latest", "test-image:v1.0"},
		OS:           "linux",
		Architecture: "amd64",
		Size:         123456789, // ~123.5 MB
		Created:      "2026-07-18T12:00:00Z",
		Config: &v1.DockerOCIImageConfig{
			ImageConfig: ocispec.ImageConfig{
				User:       "1001",
				WorkingDir: "/workspace",
				StopSignal: "SIGTERM",
				ExposedPorts: map[string]struct{}{
					"80/tcp":   {},
					"8080/tcp": {},
					"9000/udp": {},
				},
				Env: []string{
					"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
					"APP_ENV=production",
					"PORT=8080",
				},
				Entrypoint: []string{"/app/entrypoint.sh"},
				Cmd:        []string{"run", "--server"},
				Volumes: map[string]struct{}{
					"/data":     {},
					"/var/logs": {},
				},
				Labels: map[string]string{
					"maintainer": "test-developer",
					"version":    "1.0",
				},
			},
			DockerOCIImageConfigExt: v1.DockerOCIImageConfigExt{
				Healthcheck: &v1.HealthcheckConfig{
					Test:        []string{"CMD-SHELL", "curl -f http://localhost:80/health || exit 1"},
					Interval:    30000000000, // 30s
					Timeout:     5000000000,  // 5s
					StartPeriod: 10000000000, // 10s
					Retries:     3,
				},
			},
		},
	}
}

func TestExportJson(t *testing.T) {
	im := createTestImageMeta()
	out, err := exporter.ExportJson(im)
	if err != nil {
		t.Fatalf("ExportJson failed: %v", err)
	}

	// Verify it's valid JSON and contains key fields
	var parsed provider.ImageMeta
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("ExportJson returned invalid JSON: %v", err)
	}

	if parsed.ID != im.ID {
		t.Errorf("Expected ID %q, got %q", im.ID, parsed.ID)
	}
	if parsed.OS != im.OS {
		t.Errorf("Expected OS %q, got %q", im.OS, parsed.OS)
	}
}

func TestExportYaml(t *testing.T) {
	im := createTestImageMeta()
	out, err := exporter.ExportYaml(im)
	if err != nil {
		t.Fatalf("ExportYaml failed: %v", err)
	}

	// Verify it's valid YAML and contains key fields
	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("ExportYaml returned invalid YAML: %v", err)
	}

	if parsed["Id"] != im.ID {
		t.Errorf("Expected Id %q, got %v", im.ID, parsed["Id"])
	}
	if parsed["Size"] != "123.5MB" {
		t.Errorf("Expected Size %q, got %v", "123.5MB", parsed["Size"])
	}

	// Test nil ImageMeta
	_, err = exporter.ExportYaml(nil)
	if err == nil {
		t.Error("Expected error when ExportYaml is called with nil")
	}
}

func TestExportMarkdown(t *testing.T) {
	im := createTestImageMeta()
	out, err := exporter.ExportMarkdown(im)
	if err != nil {
		t.Fatalf("ExportMarkdown failed: %v", err)
	}

	expectedSubstrings := []string{
		"# Image Metadata",
		"Basic Information",
		"Config",
		"Labels",
		"sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		"test-image:latest",
		"test-image:v1.0",
		"123.5MB",
		"1001",
		"/workspace",
		"80/tcp",
		"8080/tcp",
		"9000/udp",
		"APP_ENV=production",
		"/app/entrypoint.sh",
		"run",
		"--server",
		"/data",
		"maintainer",
		"test-developer",
	}

	for _, sub := range expectedSubstrings {
		if !strings.Contains(out, sub) {
			t.Errorf("Expected markdown to contain %q, but it did not", sub)
		}
	}

	// Test nil ImageMeta
	_, err = exporter.ExportMarkdown(nil)
	if err == nil {
		t.Error("Expected error when ExportMarkdown is called with nil")
	}
}

func TestExportCompose(t *testing.T) {
	im := createTestImageMeta()
	out, err := exporter.ExportCompose(im)
	if err != nil {
		t.Fatalf("ExportCompose failed: %v", err)
	}

	// Verify it's valid YAML
	var cf exporter.ComposeFile
	if err := yaml.Unmarshal([]byte(out), &cf); err != nil {
		t.Fatalf("ExportCompose returned invalid YAML: %v", err)
	}

	serviceName := "test-image"
	s, ok := cf.Services[serviceName]
	if !ok {
		t.Fatalf("Expected service %q to exist in compose file, got %v", serviceName, cf.Services)
	}

	if s.Image != "test-image:latest" {
		t.Errorf("Expected image %q, got %q", "test-image:latest", s.Image)
	}

	// Check container_name environment variable structure
	if s.ContainerName != "${TEST-IMAGE_CONTAINER_NAME:-test-image_container}" {
		t.Errorf("Expected container_name %q, got %q", "${TEST-IMAGE_CONTAINER_NAME:-test-image_container}", s.ContainerName)
	}

	// Check environment logic: SYSTEM environment variables (like PATH) should be excluded
	if _, ok := s.Environment["PATH"]; ok {
		t.Error("System environment variable 'PATH' should be excluded from compose file environment")
	}
	if val, ok := s.Environment["APP_ENV"]; !ok || val != "${APP_ENV:-production}" {
		t.Errorf("Expected APP_ENV to be '${APP_ENV:-production}', got %q", val)
	}

	// Check ports sorting and environment variables mapping
	// 80/tcp -> ${TEST-IMAGE_PORT_80:-80}:80
	// 8080/tcp -> ${TEST-IMAGE_PORT_8080:-8080}:8080
	// 9000/udp -> ${TEST-IMAGE_PORT_9000_UDP:-9000}:9000/udp
	expectedPorts := []string{
		"${TEST-IMAGE_PORT_80:-80}:80",
		"${TEST-IMAGE_PORT_8080:-8080}:8080",
		"${TEST-IMAGE_PORT_9000_UDP:-9000}:9000/udp",
	}
	for _, p := range expectedPorts {
		found := false
		for _, sp := range s.Ports {
			if sp == p {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected port %q to be in compose ports, but got %v", p, s.Ports)
		}
	}

	// Check volumes mapping
	// Volumes are sorted: "/data" and "/var/logs"
	// For "/data": getLastPathSegment("/data") is "data" -> env suffix: "DATA" -> ${TEST-IMAGE_VOL_DATA:-./data}:/data
	// For "/var/logs": getLastPathSegment("/var/logs") is "logs" -> env suffix: "LOGS" -> ${TEST-IMAGE_VOL_LOGS:-./logs}:/var/logs
	expectedVolumes := []string{
		"${TEST-IMAGE_VOL_DATA:-./data}:/data",
		"${TEST-IMAGE_VOL_LOGS:-./logs}:/var/logs",
	}
	for _, v := range expectedVolumes {
		found := false
		for _, sv := range s.Volumes {
			if sv == v {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected volume %q to be in compose volumes, but got %v", v, s.Volumes)
		}
	}

	// Test nil ImageMeta
	_, err = exporter.ExportCompose(nil)
	if err == nil {
		t.Error("Expected error when ExportCompose is called with nil")
	}
}
