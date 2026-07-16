package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
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

	img, err := remote.Image(tag,
		remote.WithContext(ctx),
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
		remote.WithTransport(customTransport),
	)
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
