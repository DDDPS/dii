package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	ggcrv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	v1 "github.com/moby/docker-image-spec/specs-go/v1"
)

func GetMetaRegistry(ctx context.Context, imageInput string) (*ImageMeta, error) {
	tag, err := name.ParseReference(imageInput)
	if err != nil {
		return nil, fmt.Errorf("parse image error: %v", err)
	}
	regStr := tag.Context().RegistryStr()
	repoStr := tag.Context().RepositoryStr()
	tagStr := tag.Identifier()

	customTransport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,  //TCP建连超时
			KeepAlive: 15 * time.Second, //TCP保活探测
		}).DialContext,
		ForceAttemptHTTP2:     true,             //强制优先使用HTTP2
		MaxIdleConns:          100,              //最大空闲连接数
		IdleConnTimeout:       90 * time.Second, //空闲连接超时
		TLSHandshakeTimeout:   10 * time.Second, //TLS握手超时
		ResponseHeaderTimeout: 10 * time.Second, //等待响应头超时
		ExpectContinueTimeout: 1 * time.Second,  //ExpectContinue握手超时
	}

	var img ggcrv1.Image
	var pullErr error

	if regStr == "index.docker.io" || regStr == "docker.io" {
		var mirrors []string

		if envMirror := os.Getenv("DII_MIRROR"); envMirror != "" {
			for _, m := range strings.FieldsFunc(envMirror, func(r rune) bool {
				return r == ',' || r == ';'
			}) {
				if cleaned := strings.TrimSpace(m); cleaned != "" {
					mirrors = append(mirrors, cleaned)
				}
			}
		}

		if len(mirrors) == 0 {
			if p, err := NewDaemonProvider(); err == nil {
				if daemonMirrors, err := p.GetMirrorFromDocker(ctx); err == nil && len(daemonMirrors) > 0 {
					mirrors = append(mirrors, daemonMirrors...)
				}
			}
		}

		if len(mirrors) > 0 {
			img, pullErr = tryMirrors(ctx, mirrors, repoStr, tagStr, customTransport)
		}
	}

	if img == nil {
		fmt.Fprintf(os.Stderr, "Mirror failed or unavailable, falling back to original registry...\n")
		img, pullErr = remote.Image(tag,
			remote.WithContext(ctx),
			remote.WithAuthFromKeychain(authn.DefaultKeychain),
			remote.WithTransport(customTransport),
		)
	}

	if pullErr != nil {
		return nil, fmt.Errorf("pull image error: %w", pullErr)
	}

	cfgFile, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("get config file error: %v", err)
	}
	digest, err := img.Digest()
	if err != nil {
		return nil, fmt.Errorf("get image digest error: %v", err)
	}

	manifest, err := img.Manifest()
	if err != nil {
		return nil, fmt.Errorf("get image manifest error: %v", err)
	}

	var size int64
	for _, layer := range manifest.Layers {
		size += layer.Size
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
		RepoTags:     []string{tag.String()},
		OS:           cfgFile.OS,
		Architecture: cfgFile.Architecture,
		Size:         size,
		Created:      cfgFile.Created.Format(time.RFC3339),
		Config:       &dockerConfig,
	}

	fmt.Fprintf(os.Stderr, "This image is from the remote registry.\n")
	return imageMeta, nil
}

func tryMirrors(ctx context.Context, mirrors []string, repo, tag string, transport http.RoundTripper) (ggcrv1.Image, error) {
	var errs []error
	for _, mirror := range mirrors {
		cleanMirror := strings.TrimPrefix(mirror, "https://")
		cleanMirror = strings.TrimPrefix(cleanMirror, "http://")
		cleanMirror = strings.TrimRight(cleanMirror, "/")

		ref := fmt.Sprintf("%s/%s:%s", cleanMirror, repo, tag)
		parsed, err := name.ParseReference(ref)
		if err != nil {
			errs = append(errs, fmt.Errorf("mirror %s parse error: %w", mirror, err))
			continue
		}
		img, err := remote.Image(parsed,
			remote.WithContext(ctx),
			remote.WithAuthFromKeychain(authn.DefaultKeychain),
			remote.WithTransport(transport),
		)
		if err == nil {
			return img, nil
		}
		errs = append(errs, fmt.Errorf("mirror %s error: %w", mirror, err))
	}

	return nil, errors.Join(errs...)
}
