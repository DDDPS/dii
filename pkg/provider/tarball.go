package provider

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/v1/tarball"
	v1 "github.com/moby/docker-image-spec/specs-go/v1"
)

type gzipReadCloser struct {
	zr *gzip.Reader
	f  *os.File
}

func (g *gzipReadCloser) Read(p []byte) (n int, err error) {
	return g.zr.Read(p)
}

func (g *gzipReadCloser) Close() error {
	err1 := g.zr.Close()
	err2 := g.f.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

func openTarball(tarPath string) (io.ReadCloser, error) {
	f, err := os.Open(tarPath)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(tarPath, ".tar.gz") || strings.HasSuffix(tarPath, ".tgz") {
		zr, err := gzip.NewReader(f)
		if err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				log.Printf("warning: failed to close tarball file %s: %v", tarPath, closeErr)
			}
			return nil, err
		}
		return &gzipReadCloser{zr: zr, f: f}, nil
	}

	return f, nil
}

func GetMetaTarball(ctx context.Context, tarPath string) (*ImageMeta, error) {
	//support .tar .tar.gz .tgz
	img, err := tarball.Image(func() (io.ReadCloser, error) {
		return openTarball(tarPath)
	}, nil)
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
	rc, err := openTarball(tarPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rc.Close(); closeErr != nil {
			log.Printf("warning: failed to close tarball file %s: %v", tarPath, closeErr)
		}
	}()

	tr := tar.NewReader(rc)
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
