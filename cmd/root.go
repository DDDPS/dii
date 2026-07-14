package cmd

import (
	"context"
	"fmt"
	"os"

	"DII/pkg/exporter"

	"github.com/spf13/cobra"
)

var format string

var rootCmd = &cobra.Command{
	Use:  "dii [IMAGE]",
	Long: "this long msg",

	Args: cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		imageInput := args[0]
		ctx := cmd.Context()
		f := exporter.WithFormat(format)
		info, err := exporter.ExportInfo(ctx, imageInput, f)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(info)
		}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&format, "ouptut", "o", "md", "markdown,json,yaml")
}

func Execute() {
	ctx := context.Background()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
