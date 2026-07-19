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
	sb.WriteString(fmt.Sprintf("- **ID**: `%s`\n", imageMeta.ID))

	if len(imageMeta.RepoTags) > 0 {
		sb.WriteString("- **RepoTags**:\n")
		for _, tag := range imageMeta.RepoTags {
			sb.WriteString(fmt.Sprintf("  - `%s`\n", tag))
		}
	} else {
		sb.WriteString("- **RepoTags**: none\n")
	}

	sb.WriteString(fmt.Sprintf("- **OS**: `%s`\n", imageMeta.OS))
	sb.WriteString(fmt.Sprintf("- **Architecture**: `%s`\n", imageMeta.Architecture))
	sizeInMB := float64(imageMeta.Size) / 1000 / 1000
	sb.WriteString(fmt.Sprintf("- **Size**: `%.1f MB`\n", sizeInMB))
	sb.WriteString(fmt.Sprintf("- **Created**: `%s`\n", imageMeta.Created))
	sb.WriteString("\n")

	if imageMeta.Config != nil {
		sb.WriteString("## Config\n")
		if imageMeta.Config.User != "" {
			sb.WriteString(fmt.Sprintf("- **User**: `%s`\n", imageMeta.Config.User))
		}
		if imageMeta.Config.WorkingDir != "" {
			sb.WriteString(fmt.Sprintf("- **WorkingDir**: `%s`\n", imageMeta.Config.WorkingDir))
		}
		if imageMeta.Config.StopSignal != "" {
			sb.WriteString(fmt.Sprintf("- **StopSignal**: `%s`\n", imageMeta.Config.StopSignal))
		}

		// ExposedPorts
		if len(imageMeta.Config.ExposedPorts) > 0 {
			sb.WriteString("- **ExposedPorts**:\n")
			for port := range imageMeta.Config.ExposedPorts {
				sb.WriteString(fmt.Sprintf("  - `%s`\n", port))
			}
		}

		// Env
		if len(imageMeta.Config.Env) > 0 {
			sb.WriteString("- **Env**:\n")
			for _, env := range imageMeta.Config.Env {
				sb.WriteString(fmt.Sprintf("  - `%s`\n", env))
			}
		}

		// Entrypoint
		if len(imageMeta.Config.Entrypoint) > 0 {
			sb.WriteString("- **Entrypoint**:\n")
			for _, ep := range imageMeta.Config.Entrypoint {
				sb.WriteString(fmt.Sprintf("  - `%s`\n", ep))
			}
		}

		// Cmd
		if len(imageMeta.Config.Cmd) > 0 {
			sb.WriteString("- **Cmd**:\n")
			for _, cmd := range imageMeta.Config.Cmd {
				sb.WriteString(fmt.Sprintf("  - `%s`\n", cmd))
			}
		}

		// Volumes
		if len(imageMeta.Config.Volumes) > 0 {
			sb.WriteString("- **Volumes**:\n")
			for vol := range imageMeta.Config.Volumes {
				sb.WriteString(fmt.Sprintf("  - `%s`\n", vol))
			}
		}

		// Healthcheck
		if imageMeta.Config.Healthcheck != nil {
			sb.WriteString("- **Healthcheck**:\n")
			if len(imageMeta.Config.Healthcheck.Test) > 0 {
				sb.WriteString(fmt.Sprintf("  - **Test**: `%s`\n", strings.Join(imageMeta.Config.Healthcheck.Test, " ")))
			}
			sb.WriteString(fmt.Sprintf("  - **Interval**: `%s`\n", imageMeta.Config.Healthcheck.Interval))
			sb.WriteString(fmt.Sprintf("  - **Timeout**: `%s`\n", imageMeta.Config.Healthcheck.Timeout))
			sb.WriteString(fmt.Sprintf("  - **StartPeriod**: `%s`\n", imageMeta.Config.Healthcheck.StartPeriod))
			sb.WriteString(fmt.Sprintf("  - **Retries**: `%d`\n", imageMeta.Config.Healthcheck.Retries))
		}

		// Shell
		if len(imageMeta.Config.Shell) > 0 {
			sb.WriteString("- **Shell**:\n")
			for _, sh := range imageMeta.Config.Shell {
				sb.WriteString(fmt.Sprintf("  - `%s`\n", sh))
			}
		}

		// OnBuild
		if len(imageMeta.Config.OnBuild) > 0 {
			sb.WriteString("- **OnBuild**:\n")
			for _, ob := range imageMeta.Config.OnBuild {
				sb.WriteString(fmt.Sprintf("  - `%s`\n", ob))
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
