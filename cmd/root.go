package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"DII/pkg/exporter"

	"github.com/spf13/cobra"
)

var format string

var rootCmd = &cobra.Command{
	Use:  "dii [IMAGE]",
	Long: "Quickly retrieve key information about the image.",

	Args: cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		imageInput := args[0]
		ctx := cmd.Context()
		f := exporter.WithFormat(format)
		info, err := exporter.ExportInfo(ctx, imageInput, f)
		if err != nil {
			if ctx.Err() != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), "\n✕ Operation cancelled by user")
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
			}
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "%s", info)
		}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&format, "output", "o", "md", "Output format: markdown,json,yaml,compose")
}

func Execute() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
