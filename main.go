package main

import (
	"context"
	"fmt"

	"DII/pkg/exporter"
)

func main() {
	ctx := context.Background()

	info, err := exporter.ExportInfo(ctx, "ghcr.io/gohugoio/hugo:v0.159.2 ", "json")
	if err != nil {
		panic(err)
	}
	fmt.Println(info)

}
