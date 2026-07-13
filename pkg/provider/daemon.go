package provider

import (
	"context"
	"strings"

	"github.com/moby/moby/client"
)

type DaemonProvider struct {
	cli *client.Client
}

func NewDaemonProvider() (*DaemonProvider, error) {
	apiClient, err := client.New(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &DaemonProvider{cli: apiClient}, nil
}

func (p *DaemonProvider) GetMeta(ctx context.Context, imageInput string) (*ImageMeta, error) {

	imageInput = strings.TrimSpace(imageInput)

	imgRes, err := p.cli.ImageInspect(ctx, imageInput)
	if err != nil {
		return nil, err
	}

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
