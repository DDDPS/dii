package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

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

func GetImageMeta(ctx context.Context, imageInput string, mode string) (*ImageMeta, error) {
	switch mode {
	case "daemon":
		p, err := NewDaemonProvider()
		if p != nil {
			defer p.cli.Close()
		}
		if err != nil {
			return nil, err
		} else {
			imageMeta, err := p.GetMetaDaemon(ctx, imageInput)
			if err != nil {
				return nil, err
			}
			return imageMeta, nil
		}

	case "registry":
		fmt.Fprintf(os.Stderr, "Pulling image from remote registry...\n")
		return GetMetaRegistry(ctx, imageInput)

	case "tarball":
		return GetMetaTarball(ctx, imageInput)

	default:
		if info, err := os.Stat(imageInput); err == nil && !info.IsDir() && (strings.HasSuffix(imageInput, ".tar") || strings.HasSuffix(imageInput, ".tar.gz") || strings.HasSuffix(imageInput, ".tgz")) {
			return GetMetaTarball(ctx, imageInput)
		}

		p, err := NewDaemonProvider()
		if p != nil {
			defer p.cli.Close()
		}

		if err == nil && p != nil {
			imageMate, err := p.GetMetaDaemon(ctx, imageInput)
			if err == nil {
				return imageMate, nil
			}
			if strings.Contains(err.Error(), "No such image") {
				fmt.Fprintf(os.Stderr, "Image not found in local daemon, trying remote registry...\n")
			} else {
				fmt.Fprintf(os.Stderr, "Local daemon check failed (%v), trying remote registry...\n", err)
			}
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Could not connect to local Docker daemon, trying remote registry...\n")
		}

		imageMate, err := GetMetaRegistry(ctx, imageInput)
		if err != nil {
			return nil, err
		}
		return imageMate, nil
	}
}
