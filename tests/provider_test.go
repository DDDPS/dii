package tests

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"DII/pkg/provider"
)

// createMockTarball creates a minimal valid docker save tarball for testing
func createMockTarball(t *testing.T) string {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "mock_image.tar")

	f, err := os.Create(tarPath)
	if err != nil {
		t.Fatalf("failed to create temp tar file: %v", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			t.Logf("warning: failed to close temp tar file: %v", closeErr)
		}
	}()

	tw := tar.NewWriter(f)
	defer func() {
		if closeErr := tw.Close(); closeErr != nil {
			t.Logf("warning: failed to close tar writer: %v", closeErr)
		}
	}()

	// 1. Create a mock layer1.tar (empty tar archive)
	var layerBuf bytes.Buffer
	layerTw := tar.NewWriter(&layerBuf)
	if closeErr := layerTw.Close(); closeErr != nil {
		t.Fatalf("failed to close layerTw: %v", closeErr)
	}
	layerBytes := layerBuf.Bytes()

	// Write layer1.tar to main tar
	header := &tar.Header{
		Name: "layer1.tar",
		Mode: 0600,
		Size: int64(len(layerBytes)),
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("failed to write layer1.tar header: %v", err)
	}
	if _, err := tw.Write(layerBytes); err != nil {
		t.Fatalf("failed to write layer1.tar content: %v", err)
	}

	// 2. Create config.json
	configJSON := `{
		"config": {
			"User": "1000",
			"Env": ["TEST_ENV=true", "PATH=/bin"],
			"ExposedPorts": {
				"80/tcp": {}
			}
		},
		"os": "linux",
		"architecture": "amd64",
		"created": "2026-07-18T12:00:00Z"
	}`
	header = &tar.Header{
		Name: "config.json",
		Mode: 0600,
		Size: int64(len(configJSON)),
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("failed to write config.json header: %v", err)
	}
	if _, err := tw.Write([]byte(configJSON)); err != nil {
		t.Fatalf("failed to write config.json content: %v", err)
	}

	// 3. Create manifest.json
	manifestJSON := `[
		{
			"Config": "config.json",
			"RepoTags": ["my-tarball-image:v1.0", "my-tarball-image:latest"],
			"Layers": ["layer1.tar"]
		}
	]`
	header = &tar.Header{
		Name: "manifest.json",
		Mode: 0600,
		Size: int64(len(manifestJSON)),
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("failed to write manifest.json header: %v", err)
	}
	if _, err := tw.Write([]byte(manifestJSON)); err != nil {
		t.Fatalf("failed to write manifest.json content: %v", err)
	}

	return tarPath
}

func TestGetMetaTarball(t *testing.T) {
	tarPath := createMockTarball(t)
	ctx := context.Background()

	meta, err := provider.GetMetaTarball(ctx, tarPath)
	if err != nil {
		t.Fatalf("GetMetaTarball failed: %v", err)
	}

	if meta.OS != "linux" {
		t.Errorf("Expected OS 'linux', got %q", meta.OS)
	}
	if meta.Architecture != "amd64" {
		t.Errorf("Expected Architecture 'amd64', got %q", meta.Architecture)
	}
	if len(meta.RepoTags) != 2 || meta.RepoTags[0] != "my-tarball-image:v1.0" {
		t.Errorf("Expected RepoTags ['my-tarball-image:v1.0', 'my-tarball-image:latest'], got %v", meta.RepoTags)
	}
	if meta.Config == nil {
		t.Fatal("Expected Config to be non-nil")
	}
	if meta.Config.User != "1000" {
		t.Errorf("Expected User '1000', got %q", meta.Config.User)
	}
	if _, ok := meta.Config.ExposedPorts["80/tcp"]; !ok {
		t.Error("Expected ExposedPort 80/tcp to be present")
	}

	// Test with non-existent file
	_, err = provider.GetMetaTarball(ctx, "non_existent_file.tar")
	if err == nil {
		t.Error("Expected error for non-existent tar path, got nil")
	}
}

func TestGetImageMetaFallback(t *testing.T) {
	tarPath := createMockTarball(t)
	ctx := context.Background()

	// 1. Test fallback when file ends with .tar and exists -> should use tarball provider
	meta, err := provider.GetImageMeta(ctx, tarPath, "")
	if err != nil {
		t.Fatalf("GetImageMeta failed with default mode on tarball: %v", err)
	}
	if meta.OS != "linux" || meta.Architecture != "amd64" {
		t.Errorf("GetImageMeta did not resolve to tarball correctly: %v", meta)
	}

	// 2. Test fallback with non-existent tag (should try daemon, then registry, and fail since it does not exist)
	_, err = provider.GetImageMeta(ctx, "nonexistent-image-xyz:9.9.9", "")
	if err == nil {
		t.Error("Expected GetImageMeta to fail for non-existent image with empty mode, but got nil")
	}
}

func TestGetMetaRegistryInvalidImage(t *testing.T) {
	ctx := context.Background()
	_, err := provider.GetMetaRegistry(ctx, "invalid_image_name_#@!")
	if err == nil {
		t.Error("Expected GetMetaRegistry to fail for invalid image name, got nil")
	}
}

func TestGetMetaTarballGzipped(t *testing.T) {
	tarPath := createMockTarball(t)
	ctx := context.Background()

	tarData, err := os.ReadFile(tarPath)
	if err != nil {
		t.Fatalf("failed to read mock tarball: %v", err)
	}

	tmpDir := t.TempDir()
	for _, ext := range []string{".tar.gz", ".tgz"} {
		gzPath := filepath.Join(tmpDir, "mock_image"+ext)
		f, err := os.Create(gzPath)
		if err != nil {
			t.Fatalf("failed to create gzip file: %v", err)
		}

		gw := gzip.NewWriter(f)
		if _, err := gw.Write(tarData); err != nil {
			t.Fatalf("failed to write gzip content: %v", err)
		}
		if err := gw.Close(); err != nil {
			t.Fatalf("failed to close gzip writer: %v", err)
		}
		if err := f.Close(); err != nil {
			t.Fatalf("failed to close gzip file: %v", err)
		}

		meta, err := provider.GetMetaTarball(ctx, gzPath)
		if err != nil {
			t.Fatalf("GetMetaTarball failed for %s: %v", ext, err)
		}

		if meta.OS != "linux" {
			t.Errorf("Expected OS 'linux', got %q for %s", meta.OS, ext)
		}
		if len(meta.RepoTags) != 2 || meta.RepoTags[0] != "my-tarball-image:v1.0" {
			t.Errorf("Expected RepoTags, got %v for %s", meta.RepoTags, ext)
		}

		metaFallback, err := provider.GetImageMeta(ctx, gzPath, "")
		if err != nil {
			t.Fatalf("GetImageMeta fallback failed for %s: %v", ext, err)
		}
		if metaFallback.OS != "linux" {
			t.Errorf("Expected fallback OS 'linux', got %q for %s", metaFallback.OS, ext)
		}
	}
}
