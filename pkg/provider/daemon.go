package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/moby/moby/client"
)

type DaemonProvider struct {
	cli *client.Client
}

func NewDaemonProvider() (*DaemonProvider, error) {
	apiClient, err := client.New(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("init docker client error: %v", err)
	}
	return &DaemonProvider{cli: apiClient}, nil
}

func (p *DaemonProvider) GetMetaDaemon(ctx context.Context, imageInput string) (*ImageMeta, error) {
	imageInput = strings.TrimSpace(imageInput)

	imgRes, err := p.cli.ImageInspect(ctx, imageInput)
	if err != nil {
		return nil, fmt.Errorf("daemon image inspect error: %v", err)
	}

	fmt.Fprintf(os.Stderr, "This image is local.\n")
	return &ImageMeta{
		ID:           imgRes.ID,
		RepoTags:     imgRes.RepoTags,
		OS:           imgRes.Os,
		Architecture: imgRes.Architecture,
		Size:         imgRes.Size,
		Created:      imgRes.Created,
		Config:       imgRes.Config,
	}, nil

}

func (p *DaemonProvider) GetMirrorFromDocker(ctx context.Context) ([]string, error) {
	info, err := p.cli.Info(ctx, client.InfoOptions{})
	if err != nil {
		return nil, fmt.Errorf("daemon info error: %v", err)
	}

	return info.Info.RegistryConfig.Mirrors, nil
}
