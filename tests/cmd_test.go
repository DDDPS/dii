package tests

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "dii-test-*")
	if err != nil {
		fmt.Printf("failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	binPath = filepath.Join(tmpDir, "dii")
	buildCmd := exec.Command("go", "build", "-o", binPath, "../main.go")
	buildCmd.Stderr = os.Stderr
	buildCmd.Stdout = os.Stdout
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("failed to build binary: %v\n", err)
		if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
			log.Printf("warning: failed to remove temp dir: %v", removeErr)
		}
		os.Exit(1)
	}

	code := m.Run()
	if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
		log.Printf("warning: failed to remove temp dir: %v", removeErr)
	}
	os.Exit(code)
}

func TestCLINoArgs(t *testing.T) {
	cmd := exec.Command(binPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected command to fail with no arguments, but it succeeded")
	}

	errMsg := stderr.String()
	if !strings.Contains(errMsg, "accepts 1 arg") && !strings.Contains(errMsg, "Error:") {
		t.Errorf("Expected error message about args, got: %q", errMsg)
	}
}

func TestCLIMultipleSources(t *testing.T) {
	cmd := exec.Command(binPath, "-d", "-r", "ubuntu")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected command to fail with multiple source flags, but it succeeded")
	}

	errMsg := stderr.String()
	expectedErr := "Error: Cannot specify more than one data source flag: choose either --daemon (-d), --registry (-r), or --tar (-t)"
	if !strings.Contains(errMsg, expectedErr) {
		t.Errorf("Expected error %q, got: %q", expectedErr, errMsg)
	}
}

func TestCLITarballExporters(t *testing.T) {
	tarPath := createMockTarball(t)

	tests := []struct {
		format         string
		expectedSubs   []string
		unexpectedSubs []string
	}{
		{
			format: "md",
			expectedSubs: []string{
				"# Image Metadata",
				"my-tarball-image:v1.0",
				"TEST_ENV=true",
			},
		},
		{
			format: "json",
			expectedSubs: []string{
				`"Id":`,
				`"RepoTags":`,
				`"my-tarball-image:v1.0"`,
			},
		},
		{
			format: "yaml",
			expectedSubs: []string{
				"Id:",
				"RepoTags:",
				"my-tarball-image:v1.0",
			},
		},
		{
			format: "compose",
			expectedSubs: []string{
				"version: \"3.8\"",
				"services:",
				"my-tarball-image:",
				"image: my-tarball-image:v1.0",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.format, func(t *testing.T) {
			cmd := exec.Command(binPath, "-t", tarPath, "-o", tc.format)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err != nil {
				t.Fatalf("CLI command failed: %v, stderr: %s", err, stderr.String())
			}

			output := stdout.String()
			for _, sub := range tc.expectedSubs {
				if !strings.Contains(output, sub) {
					t.Errorf("Expected output to contain %q, but got:\n%s", sub, output)
				}
			}
		})
	}
}

func TestCLIWriteToFile(t *testing.T) {
	tarPath := createMockTarball(t)
	outPath := filepath.Join(t.TempDir(), "output.json")

	cmd := exec.Command(binPath, "-t", tarPath, "-o", "json", "-w", outPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("CLI command failed: %v, stderr: %s", err, stderr.String())
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, `"my-tarball-image:v1.0"`) {
		t.Errorf("Expected output file to contain repo tag, but got:\n%s", content)
	}

	stderrMsg := stderr.String()
	if !strings.Contains(stderrMsg, "written to file") {
		t.Errorf("Expected stderr to confirm writing to file, got: %q", stderrMsg)
	}
}

func TestCLIVersion(t *testing.T) {
	cmd := exec.Command(binPath, "version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		t.Fatalf("CLI command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "dii version") {
		t.Errorf("Expected output to contain 'dii version', got: %q", output)
	}
}
