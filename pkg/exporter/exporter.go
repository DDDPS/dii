package exporter

import (
	"context"

	"DII/pkg/provider"
)

type exportConfig struct {
	format      string
	useDaemon   bool
	useRegistry bool
	useTarball  bool
}

type ExportOption func(*exportConfig)

func WithFormat(format string) ExportOption {
	return func(c *exportConfig) {
		c.format = format
	}
}

func WithUseDaemon(useDaemon bool) ExportOption {
	return func(c *exportConfig) {
		c.useDaemon = useDaemon
	}
}
func WithUseRegistry(useRegistry bool) ExportOption {
	return func(c *exportConfig) {
		c.useRegistry = useRegistry
	}
}
func WithUseTarball(useTarball bool) ExportOption {
	return func(c *exportConfig) {
		c.useTarball = useTarball
	}
}

func ExportInfo(ctx context.Context, imageInput string, opts ...ExportOption) (string, error) {

	cfg := &exportConfig{}
	for _, optFn := range opts {
		optFn(cfg)
	}

	var mode string
	if cfg.useDaemon {
		mode = "daemon"
	} else if cfg.useRegistry {
		mode = "registry"
	} else if cfg.useTarball {
		mode = "tarball"
	}

	imageMeta, err := provider.GetImageMeta(ctx, imageInput, mode)
	if err != nil {
		return "", err
	}
	switch cfg.format {
	case "json":
		return ExportJson(imageMeta)
	case "md":
		return ExportMarkdown(imageMeta)
	case "yaml":
		return ExportYaml(imageMeta)
	case "compose":
		return ExportCompose(imageMeta)
	default:
		return ExportMarkdown(imageMeta)
	}

}
