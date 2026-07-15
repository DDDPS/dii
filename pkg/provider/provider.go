package provider

import (
	"context"
	"log"

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
	if err != nil {
		log.Printf("init docker client error: %v", err)
	}
	defer p.cli.Close()

	imageMate, err := p.GetMetaDaemon(ctx, imageInput)
	if err != nil {
		imageMate, err = GetMetaRegistry(ctx, imageInput)
		if err != nil {
			return nil, err
		}
		return imageMate, nil
	}
	return imageMate, nil
}
