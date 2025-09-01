package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

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

	currentVer, parseErr := semver.Parse(Version)
	if parseErr != nil {
		// If the built Version isn't valid semver, continue but warn.
		fmt.Printf("warning: could not parse current version %q: %v\n", Version, parseErr)
	}

	// No release found or nil result -> nothing to do.
	if !found || latest == nil {
		fmt.Printf("No releases found for %s.\n", repo)
		return nil
	}

	// If same version -> up-to-date.
	if latest.Version.Equals(currentVer) {
		fmt.Printf("You are already running the latest version: %s.\n", currentVer)
		return nil
	}

	// If we don't have an asset URL, cannot update automatically.
	if latest.AssetURL == "" {
		fmt.Printf("A new version (%s) is available but there is no downloadable asset.\n", latest.Version)
		fmt.Println("Please visit the project releases page to download the new version.")
		return nil
	}

	// Prompt the user to confirm updating.
	answer, perr := promptLine(fmt.Sprintf("A new version (%s) is available. Update now? (y/N): ", latest.Version))
	if perr != nil {
		return fmt.Errorf("failed reading input: %w", perr)
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		fmt.Println("Update cancelled.")
		return nil
	}

	fmt.Println("Updating...")
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not locate executable: %w", err)
	}

	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	// Attempt to restart the process by replacing the current process image.
	argv := append([]string{exe}, os.Args[1:]...)
	if err := syscall.Exec(exe, argv, os.Environ()); err != nil {
		// Exec only returns on error. Try a fallback of starting the new binary as a child process.
		cmd := exec.Command(exe, os.Args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if startErr := cmd.Start(); startErr != nil {
			// If fallback also fails, report success but instruct user to restart manually.
			fmt.Printf("Updated to version %s, but failed to restart automatically: %v; fallback start error: %v\n", latest.Version, err, startErr)
			fmt.Println("Please restart the application manually.")
			return nil
		}
		// Successfully started the new process; exit the current one.
		os.Exit(0)
	}

	// If Exec succeeds, this process is replaced and the following lines won't run.
	return nil
}
