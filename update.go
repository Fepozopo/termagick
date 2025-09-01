package main

import (
	"fmt"
	"os"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

var Version = "0.1.0"

func checkForUpdates() error {
	const repo = "Fepozopo/termagick"
	latest, found, err := selfupdate.DetectLatest(repo)
	if err != nil {
		return fmt.Errorf("update check failed: %w", err)
	}

	currentVer, _ := semver.Parse(Version)
	if !found || latest.Version.Equals(currentVer) {
		fmt.Printf("You are already running the latest version: %s.\n", latest.Version)
		return nil
	}
	fmt.Printf("A new version (%s) is available.\n", latest.Version)

	fmt.Println("Updating...")
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not locate executable: %w", err)
	}
	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		return fmt.Errorf("could not locate executable: %w", err)
	}
	fmt.Printf("Successfully updated to version %s. Please restart the application.\n", latest.Version)
	return nil
}
