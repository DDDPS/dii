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
	Version     = "dev"
)

var rootCmd = &cobra.Command{
	Use:     "dii [flags] [IMAGE]",
	Long:    "Quickly retrieve key information about the image.",
	Version: Version,

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
			cmd.PrintErrln("Error: Cannot specify more than one data source flag: choose either --daemon (-d), --registry (-r), or --tar (-t)")
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
				cmd.PrintErrln("\n✕ Operation cancelled by user")
			} else {
				cmd.PrintErrln(err)
			}
		} else {
			if write != "" {
				err := os.WriteFile(write, []byte(info), 0644)
				cmd.PrintErrln("written to file:", write)
				if err != nil {
					cmd.PrintErrln(err)
					os.Exit(1)
				}
			} else {
				if _, err := fmt.Fprint(cmd.OutOrStdout(), info); err != nil {
					cmd.PrintErrln("failed to write output:", err)
				}
			}
		}
	},
}

var versionCmd = &cobra.Command{
	Use:    "version",
	Short:  "Print the version number of dii",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "dii version %s\n", Version)
		return err
	},
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.AddCommand(versionCmd)

	rootCmd.Flags().SortFlags = false
	rootCmd.Flags().StringVarP(&format, "output", "o", "md", "Output format: markdown,json,yaml,compose")
	rootCmd.Flags().BoolVarP(&useDaemon, "daemon", "d", false, "Retrieve image data via the local Docker daemon")
	rootCmd.Flags().BoolVarP(&useRegistry, "registry", "r", false, "Retrieve image data via a remote registry")
	rootCmd.Flags().BoolVarP(&useTarball, "tar", "t", false, "Retrieve image data from a local image(support .tar, .tar.gz, .tgz)")
	rootCmd.Flags().StringVarP(&write, "write", "w", "", "Write the image data to a file(0644): dii ubuntu -w <filename>")
}

func Execute() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
