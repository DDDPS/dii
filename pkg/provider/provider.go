package provider

import (
	"context"
	"fmt"
	"os"

	v1 "github.com/moby/docker-image-spec/specs-go/v1"
)

type ImageMeta struct {
	ID           string                   `json:"Id"`           //ID
	RepoTags     []string                 `json:"RepoTags"`     //标签
	OS           string                   `json:"Os"`           //操作系统
	Architecture string                   `json:"Architecture"` //架构
	Size         int64                    `json:"Size"`         //大小
	Created      string                   `json:"Created"`      //创建时间
	Config       *v1.DockerOCIImageConfig `json:"Config"`       //配置-包含env port volumes等
}

func GetImageMeta(ctx context.Context, imageInput string) (*ImageMeta, error) {
	p, err := NewDaemonProvider()
	if p != nil {
		defer p.cli.Close()
	}

	if err == nil && p != nil {
		imageMate, err := p.GetMetaDaemon(ctx, imageInput)
		if err == nil {
			return imageMate, nil
		}

		fmt.Fprintf(os.Stderr, "local image inspect failed: %v, trying remote registry...\n", err)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "init docker client error: %v, trying remote registry...\n", err)
	}

	imageMate, err := GetMetaRegistry(ctx, imageInput)
	if err != nil {
		return nil, err
	}
	return imageMate, nil
}
