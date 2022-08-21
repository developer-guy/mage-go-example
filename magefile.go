//go:build mage
// +build mage

package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/carolynvs/magex/pkg"
	"github.com/carolynvs/magex/pkg/archive"
	"github.com/carolynvs/magex/pkg/downloads"
	"github.com/magefile/mage/sh"

	"sigs.k8s.io/release-utils/mage"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
var Default = BuildImagesLocal

// BuildImages build bom image using ko
func BuildImages() error {
	fmt.Println("Building images with ko...")
	if err := EnsureBinary("ko", "version", "0.11.2"); err != nil {
		return err
	}

	ldFlags, _ := mage.GenerateLDFlags()
	os.Setenv("LDFLAGS", ldFlags)
	os.Setenv("KOCACHE", "/tmp/ko")

	if os.Getenv("KO_DOCKER_REPO") == "" {
		return errors.New("missing KO_DOCKER_REPO environment variable")
	}
	// echo "${{ github.token }}" | ./ko login ghcr.io --username "${{ github.actor }}" --password-stdin
	_ = sh.RunV("ko", "login", os.Getenv("ghcr.io"), "-u", os.Getenv("GITHUB_ACTOR"), "-p", os.Getenv("GITHUB_TOKEN"))

	return sh.RunV("ko", "build", "--bare",
		"--platform=linux/amd64", "-t", "latest",
		"-t", os.Getenv("GITHUB_REF_NAME"),
		".")
}

// BuildImagesLocal build images locally and not push
func BuildImagesLocal() error {
	fmt.Println("Building images with ko for local...")
	if err := EnsureBinary("ko", "version", "0.11.2"); err != nil {
		return err
	}

	ldFlags, _ := mage.GenerateLDFlags()
	os.Setenv("LDFLAGS", ldFlags)
	os.Setenv("KOCACHE", "/tmp/ko")

	return sh.RunV("ko", "build", "--bare",
		"--local", "--platform=linux/amd64",
		".")
}

func Release() error {
	fmt.Println("Releasing greeting with goreleaser...")
	if err := EnsureBinary("goreleaser", "-v", "1.10.3"); err != nil {
		return err
	}

	ldFlags, _ := mage.GenerateLDFlags()
	os.Setenv("LDFLAGS", ldFlags)

	args := []string{"release", "--rm-dist"}

	return sh.RunV("goreleaser", args...)
}

func EnsureBinary(binary, cmd, version string) error {
	fmt.Printf("Checking if `%s` version %s is installed\n", binary, version)
	found, err := pkg.IsCommandAvailable(binary, cmd, "")
	if err != nil {
		return err
	}

	if !found {
		fmt.Printf("`%s` not found\n", binary)
		switch binary {
		case "goreleaser":
			return InstallGoReleaser(version)
		case "ko":
			return InstallKO(version)
		}
	}

	fmt.Printf("`%s` is installed!\n", binary)
	return nil
}

func InstallKO(version string) error {
	fmt.Println("Will install `ko`")
	target := "ko"
	if runtime.GOOS == "windows" {
		target = "ko.exe"
	}

	opts := archive.DownloadArchiveOptions{
		DownloadOptions: downloads.DownloadOptions{
			UrlTemplate: "https://github.com/google/ko/releases/download/v{{.VERSION}}/ko_{{.VERSION}}_{{.GOOS}}_{{.GOARCH}}{{.EXT}}",
			Name:        "ko",
			Version:     version,
			OsReplacement: map[string]string{
				"darwin":  "Darwin",
				"linux":   "Linux",
				"windows": "Windows",
			},
			ArchReplacement: map[string]string{
				"amd64": "x86_64",
			},
		},
		ArchiveExtensions: map[string]string{
			"linux":   ".tar.gz",
			"darwin":  ".tar.gz",
			"windows": ".tar.gz",
		},
		TargetFileTemplate: target,
	}

	return archive.DownloadToGopathBin(opts)
}

func InstallGoReleaser(version string) error {
	fmt.Println("Will install `goreleaser` version `%s`", version)
	target := "goreleaser"
	opts := archive.DownloadArchiveOptions{
		DownloadOptions: downloads.DownloadOptions{
			// https://github.com/goreleaser/goreleaser/releases/download/v1.10.3/goreleaser_Linux_arm64.tar.gz.sbom
			UrlTemplate: "https://github.com/goreleaser/goreleaser/releases/download/v{{.VERSION}}/goreleaser_{{.GOOS}}_{{.GOARCH}}{{.EXT}}",
			Name:        "goreleaser",
			Version:     version,
			OsReplacement: map[string]string{
				"darwin":  "Darwin",
				"linux":   "Linux",
				"windows": "Windows",
			},
			ArchReplacement: map[string]string{
				"amd64": "x86_64",
			},
		},
		ArchiveExtensions: map[string]string{
			"linux":   ".tar.gz",
			"darwin":  ".tar.gz",
			"windows": ".tar.gz",
		},
		TargetFileTemplate: target,
	}

	return archive.DownloadToGopathBin(opts)
}
