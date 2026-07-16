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

var (
	format      string
	useDaemon   bool
	useRegistry bool
)

var rootCmd = &cobra.Command{
	Use:  "dii [flags] [IMAGE]",
	Long: "Quickly retrieve key information about the image.",

	Args: cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		if useDaemon && useRegistry {
			fmt.Fprintln(cmd.ErrOrStderr(), "Error: Cannot specify both --daemon (-d) and --registry (-r)")
			os.Exit(1)
		}
		imageInput := args[0]
		ctx := cmd.Context()
		o := exporter.WithFormat(format)
		ud := exporter.WithUseDaemon(useDaemon)
		ur := exporter.WithUseRegistry(useRegistry)

		info, err := exporter.ExportInfo(ctx, imageInput, o, ud, ur)
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
	rootCmd.Flags().SortFlags = false
	rootCmd.Flags().StringVarP(&format, "output", "o", "md", "Output format: markdown,json,yaml,compose")
	rootCmd.Flags().BoolVarP(&useDaemon, "daemon", "d", false, "Retrieve image data via the local Docker daemon")
	rootCmd.Flags().BoolVarP(&useRegistry, "registry", "r", false, "Retrieve image data via a remote registry")
}

func Execute() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
