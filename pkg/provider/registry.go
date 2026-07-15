package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	v1 "github.com/moby/docker-image-spec/specs-go/v1"
)

func GetMetaRegistry(ctx context.Context, imageInput string) (*ImageMeta, error) {
	tag, err := name.ParseReference(imageInput)
	if err != nil {
		return nil, fmt.Errorf("parse image error: %v", err)
	}
	img, err := remote.Image(tag, remote.WithContext(ctx), remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return nil, fmt.Errorf("pull image from registry error: %v", err)
	}

	cfgFile, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("get config file error: %v", err)
	}
	digest, err := img.Digest()
	if err != nil {
		return nil, fmt.Errorf("get image digest error: %v", err)
	}

	size, err := img.Size()
	if err != nil {
		return nil, fmt.Errorf("get image size error: %v", err)
	}

	// (ggcr) v1.Config -> (moby) v1.DockerOCIImageConfig
	configBytes, err := json.Marshal(cfgFile.Config)
	if err != nil {
		return nil, fmt.Errorf("marshal config error: %v", err)
	}

	var dockerConfig v1.DockerOCIImageConfig
	if err := json.Unmarshal(configBytes, &dockerConfig); err != nil {
		return nil, fmt.Errorf("unmarshal config error: %v", err)
	}

	imageMeta := &ImageMeta{
		ID:           digest.String(),
		RepoTags:     []string{tag.Identifier()},
		OS:           cfgFile.OS,
		Architecture: cfgFile.Architecture,
		Size:         size,
		Created:      cfgFile.Created.Format(time.RFC3339),
		Config:       &dockerConfig,
	}

	fmt.Fprintf(os.Stderr, "This image is from the remote registry.\n")
	return imageMeta, nil
}
