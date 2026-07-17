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
	useTarball  bool
	write       string
)

var rootCmd = &cobra.Command{
	Use:  "dii [flags] [IMAGE]",
	Long: "Quickly retrieve key information about the image.",

	Args: cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		count := 0
		if useDaemon {
			count++
		}
		if useRegistry {
			count++
		}
		if useTarball {
			count++
		}
		if count > 1 {
			fmt.Fprintln(cmd.ErrOrStderr(), "Error: Cannot specify more than one data source flag: choose either --daemon (-d), --registry (-r), or --tar (-t)")
			os.Exit(1)
		}
		imageInput := args[0]
		ctx := cmd.Context()
		o := exporter.WithFormat(format)
		ud := exporter.WithUseDaemon(useDaemon)
		ur := exporter.WithUseRegistry(useRegistry)
		ut := exporter.WithUseTarball(useTarball)

		info, err := exporter.ExportInfo(ctx, imageInput, o, ud, ur, ut)
		if err != nil {
			if ctx.Err() != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), "\n✕ Operation cancelled by user")
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
			}
		} else {
			if write != "" {
				err := os.WriteFile(write, []byte(info), 0644)
				fmt.Fprintf(cmd.ErrOrStderr(), "written to file: %s\n", write)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
					os.Exit(1)
				}
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "%s", info)
			}
		}
	},
}

func init() {
	rootCmd.Flags().SortFlags = false
	rootCmd.Flags().StringVarP(&format, "output", "o", "md", "Output format: markdown,json,yaml,compose")
	rootCmd.Flags().BoolVarP(&useDaemon, "daemon", "d", false, "Retrieve image data via the local Docker daemon")
	rootCmd.Flags().BoolVarP(&useRegistry, "registry", "r", false, "Retrieve image data via a remote registry")
	rootCmd.Flags().BoolVarP(&useTarball, "tar", "t", false, "Retrieve image data from a local image")
	rootCmd.Flags().StringVarP(&write, "write", "w", "", "Write the image data to a file(0644): dii ubuntu -w <filename>")
}

func Execute() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
