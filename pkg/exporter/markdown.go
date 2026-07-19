package exporter

import (
	"DII/pkg/provider"
	"encoding/json"
	"fmt"
	"strings"
)

func ExportMarkdown(imageMeta *provider.ImageMeta) (string, error) {
	if imageMeta == nil {
		return "", fmt.Errorf("imageMeta is nil")
	}

	var sb strings.Builder

	sb.WriteString("# Image Metadata\n\n")

	sb.WriteString("## Basic Information\n")
	fmt.Fprintf(&sb, "- **ID**: `%s`\n", imageMeta.ID)

	if len(imageMeta.RepoTags) > 0 {
		sb.WriteString("- **RepoTags**:\n")
		for _, tag := range imageMeta.RepoTags {
			fmt.Fprintf(&sb, "  - `%s`\n", tag)
		}
	} else {
		sb.WriteString("- **RepoTags**: none\n")
	}

	fmt.Fprintf(&sb, "- **OS**: `%s`\n", imageMeta.OS)
	fmt.Fprintf(&sb, "- **Architecture**: `%s`\n", imageMeta.Architecture)
	sizeInMB := float64(imageMeta.Size) / 1000 / 1000
	fmt.Fprintf(&sb, "- **Size**: `%.1f MB`\n", sizeInMB)
	fmt.Fprintf(&sb, "- **Created**: `%s`\n", imageMeta.Created)
	sb.WriteString("\n")

	if imageMeta.Config != nil {
		sb.WriteString("## Config\n")
		if imageMeta.Config.User != "" {
			fmt.Fprintf(&sb, "- **User**: `%s`\n", imageMeta.Config.User)
		}
		if imageMeta.Config.WorkingDir != "" {
			fmt.Fprintf(&sb, "- **WorkingDir**: `%s`\n", imageMeta.Config.WorkingDir)
		}
		if imageMeta.Config.StopSignal != "" {
			fmt.Fprintf(&sb, "- **StopSignal**: `%s`\n", imageMeta.Config.StopSignal)
		}

		// ExposedPorts
		if len(imageMeta.Config.ExposedPorts) > 0 {
			sb.WriteString("- **ExposedPorts**:\n")
			for port := range imageMeta.Config.ExposedPorts {
				fmt.Fprintf(&sb, "  - `%s`\n", port)
			}
		}

		// Env
		if len(imageMeta.Config.Env) > 0 {
			sb.WriteString("- **Env**:\n")
			for _, env := range imageMeta.Config.Env {
				fmt.Fprintf(&sb, "  - `%s`\n", env)
			}
		}

		// Entrypoint
		if len(imageMeta.Config.Entrypoint) > 0 {
			sb.WriteString("- **Entrypoint**:\n")
			for _, ep := range imageMeta.Config.Entrypoint {
				fmt.Fprintf(&sb, "  - `%s`\n", ep)
			}
		}

		// Cmd
		if len(imageMeta.Config.Cmd) > 0 {
			sb.WriteString("- **Cmd**:\n")
			for _, cmd := range imageMeta.Config.Cmd {
				fmt.Fprintf(&sb, "  - `%s`\n", cmd)
			}
		}

		// Volumes
		if len(imageMeta.Config.Volumes) > 0 {
			sb.WriteString("- **Volumes**:\n")
			for vol := range imageMeta.Config.Volumes {
				fmt.Fprintf(&sb, "  - `%s`\n", vol)
			}
		}

		// Healthcheck
		if imageMeta.Config.Healthcheck != nil {
			sb.WriteString("- **Healthcheck**:\n")
			if len(imageMeta.Config.Healthcheck.Test) > 0 {
				fmt.Fprintf(&sb, "  - **Test**: `%s`\n", strings.Join(imageMeta.Config.Healthcheck.Test, " "))
			}
			fmt.Fprintf(&sb, "  - **Interval**: `%s`\n", imageMeta.Config.Healthcheck.Interval)
			fmt.Fprintf(&sb, "  - **Timeout**: `%s`\n", imageMeta.Config.Healthcheck.Timeout)
			fmt.Fprintf(&sb, "  - **StartPeriod**: `%s`\n", imageMeta.Config.Healthcheck.StartPeriod)
			fmt.Fprintf(&sb, "  - **Retries**: `%d`\n", imageMeta.Config.Healthcheck.Retries)
		}

		// Shell
		if len(imageMeta.Config.Shell) > 0 {
			sb.WriteString("- **Shell**:\n")
			for _, sh := range imageMeta.Config.Shell {
				fmt.Fprintf(&sb, "  - `%s`\n", sh)
			}
		}

		// OnBuild
		if len(imageMeta.Config.OnBuild) > 0 {
			sb.WriteString("- **OnBuild**:\n")
			for _, ob := range imageMeta.Config.OnBuild {
				fmt.Fprintf(&sb, "  - `%s`\n", ob)
			}
		}

		// Labels
		if len(imageMeta.Config.Labels) > 0 {
			sb.WriteString("\n## Labels\n")
			labelsJSON, err := json.MarshalIndent(imageMeta.Config.Labels, "", "  ")
			if err != nil {
				return "", err
			}
			sb.WriteString("```json\n")
			sb.WriteString(string(labelsJSON))
			sb.WriteString("\n```\n")
		}
	}

	return sb.String(), nil
}
