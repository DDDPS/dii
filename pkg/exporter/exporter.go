package exporter

import (
	"context"
	"errors"

	"DII/pkg/provider"
)

func ExportInfo(ctx context.Context, imageInput string, opt string) (string, error) {
	switch opt {
	case "md", "yaml", "json", "":

	default:
		return "", errors.New("error format")
	}
	imageMeta, err := provider.GetImageMeta(ctx, imageInput)
	if err != nil {
		return "", err
	}
	switch opt {
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
