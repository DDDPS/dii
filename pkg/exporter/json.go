package exporter

import (
	"DII/pkg/provider"
	"encoding/json"
)

func ExportJson(imageMeta *provider.ImageMeta) (string, error) {
	jsonData, err := json.MarshalIndent(imageMeta, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
