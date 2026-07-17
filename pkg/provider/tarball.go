package provider

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/go-containerregistry/pkg/v1/tarball"
	v1 "github.com/moby/docker-image-spec/specs-go/v1"
)

func GetMetaTarball(ctx context.Context, tarPath string) (*ImageMeta, error) {
	img, err := tarball.ImageFromPath(tarPath, nil)
	if err != nil {
		return nil, fmt.Errorf("read tarball error: %v", err)
	}

	config, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("get config file error: %v", err)
	}

	digest, err := img.Digest()
	if err != nil {
		return nil, fmt.Errorf("get image digest error: %v", err)
	}

	// Calculate size by summing up layer sizes
	var size int64
	manifest, err := img.Manifest()
	if err == nil && manifest != nil {
		for _, layer := range manifest.Layers {
			size += layer.Size
		}
	}
	if size == 0 {
		if fileInfo, err := os.Stat(tarPath); err == nil {
			size = fileInfo.Size()
		}
	}

	// (ggcr) v1.Config -> (moby) v1.DockerOCIImageConfig
	configBytes, err := json.Marshal(config.Config)
	if err != nil {
		return nil, fmt.Errorf("marshal config error: %v", err)
	}

	var dockerConfig v1.DockerOCIImageConfig
	if err := json.Unmarshal(configBytes, &dockerConfig); err != nil {
		return nil, fmt.Errorf("unmarshal config error: %v", err)
	}

	tags, _ := getTagsFromTarball(tarPath)

	imageMate := &ImageMeta{
		ID:           digest.String(),
		RepoTags:     tags,
		OS:           config.OS,
		Architecture: config.Architecture,
		Size:         size,
		Created:      config.Created.Format(time.RFC3339),
		Config:       &dockerConfig,
	}

	return imageMate, nil
}

func getTagsFromTarball(tarPath string) ([]string, error) {
	f, err := os.Open(tarPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tr := tar.NewReader(f)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if header.Name == "manifest.json" {
			type manifestEntry struct {
				RepoTags []string `json:"RepoTags"`
			}
			var entries []manifestEntry
			if err := json.NewDecoder(tr).Decode(&entries); err != nil {
				return nil, err
			}
			var tags []string
			for _, entry := range entries {
				tags = append(tags, entry.RepoTags...)
			}
			return tags, nil
		}
	}
	return nil, nil
}
