package exporter

import (
	"context"

	"DII/pkg/provider"
)

type exportConfig struct {
	format string
}

type ExportOption func(*exportConfig)

func WithFormat(format string) ExportOption {
	return func(c *exportConfig) {
		c.format = format
	}
}

func ExportInfo(ctx context.Context, imageInput string, opts ...ExportOption) (string, error) {

	cfg := &exportConfig{}
	for _, optFn := range opts {
		optFn(cfg)
	}

	imageMeta, err := provider.GetImageMeta(ctx, imageInput)
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
	default:
		return ExportMarkdown(imageMeta)
	}

}
