package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/gographics/imagick.v3/imagick"
)

// Terminal preview helper for Kitty and iTerm2 inline-image protocols.
// Exported entry-point: PreviewWand(wand *imagick.MagickWand) error
//
// Usage:
//
//	if err := PreviewWand(wand); err != nil {
//	    // preview not available or failed
//	}
//
// Behavior:
//   - If kitty is detected (KITTY_WINDOW_ID or TERM contains "kitty"), the PNG is sent using
//     the kitty graphics protocol (chunked base64 inside ESC _G ... ESC \).
//   - Else if iTerm2 is detected (TERM_PROGRAM == "iTerm.app" || ITERM_SESSION_ID present),
//     the PNG is sent using the iTerm2 OSC 1337 inline file sequence.
//   - If neither is detected, PreviewWand returns an error indicating no supported terminal.
//
// Notes:
//   - The function clones the provided wand to set the image format to PNG without mutating
//     the caller's wand state.
//   - Sending binary escape sequences to stdout is expected in this terminal-only preview mode.
//
// Debugging helper controlled by PREVIEW_DEBUG=1
var previewDebug bool

func init() {
	err := godotenv.Load()
	if err != nil {
		// Ignore error if .env not present; it's optional
	}

	debug := os.Getenv("PREVIEW_DEBUG")
	if debug == "1" || debug == "true" {
		previewDebug = true
	}
}

func debugf(format string, args ...interface{}) {
	if previewDebug {
		fmt.Fprintf(os.Stderr, "termagick-preview: "+format+"\n", args...)
	}
}

func isKitty() bool {
	// Primary hint that the terminal is kitty or a kitty-compatible implementation
	// (e.g. ghostty exposes the kitty compatibility features).
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return true
	}
	// Inspect TERM for known kitty-compatible names.
	term := strings.ToLower(os.Getenv("TERM"))
	// Accept kitty and ghostty (and short 'ghost') as kitty-compatible terminals.
	if strings.Contains(term, "kitty") || strings.Contains(term, "ghostty") || strings.Contains(term, "ghost") {
		return true
	}
	// Konsole may implement parts of the protocol via an older kitty compatibility mode.
	if os.Getenv("KONSOLE_VERSION") != "" {
		return true
	}
	return false
}

// Detects terminals that implement the generic \"inline images\" OSC protocol
// (iTerm2 style) — many modern terminal emulators (WezTerm, Warp, Tabby, VSCode's terminal,
// Rio, Hyper, Bobcat and others) implement that or compatible behavior.
// We use a heuristic based on TERM_PROGRAM and common TERM substrings.
func isInlineImageCapable() bool {
	debugf("checking inline-image capability via TERM_PROGRAM/TERM")
	switch os.Getenv("TERM_PROGRAM") {
	case "iTerm.app", "WezTerm", "Warp", "Hyper", "vscode", "VSCode", "Tabby", "Bobcat":
		debugf("TERM_PROGRAM indicates inline-capable: %s", os.Getenv("TERM_PROGRAM"))
		return true
	}
	// Some terminals expose recognizable TERM values
	term := strings.ToLower(os.Getenv("TERM"))
	if strings.Contains(term, "wezterm") || strings.Contains(term, "warp") || strings.Contains(term, "tabby") ||
		strings.Contains(term, "vscode") || strings.Contains(term, "wez") {
		debugf("TERM suggests inline-capable: %s", term)
		return true
	}
	// A direct iTerm2 hint
	if os.Getenv("ITERM_SESSION_ID") != "" || os.Getenv("TERM_PROGRAM") == "iTerm.app" {
		debugf("iTerm2 indicators present")
		return true
	}
	return false
}

// Detect terminals that likely support Sixel graphics (foot, Windows Terminal >= certain versions,
// st with sixel patch, Black Box, etc). This is heuristic — if you rely on Sixel in CI, add
// a user-configurable override environment variable SIXEL_PREVIEW=1 to force it.
func isSixelCapable() bool {
	if os.Getenv("SIXEL_PREVIEW") == "1" {
		return true
	}
	term := strings.ToLower(os.Getenv("TERM"))
	if strings.Contains(term, "foot") || strings.Contains(term, "st") || strings.Contains(term, "linux") {
		return true
	}
	if os.Getenv("WT_SESSION") != "" { // Windows Terminal newer versions support sixel
		return true
	}
	return false
}

// PreviewSupported returns true if the running environment likely supports a terminal inline preview.
func PreviewSupported() bool {
	supported := isKitty() || isInlineImageCapable() || isSixelCapable()
	debugf("PreviewSupported -> %v (kitty=%v inline=%v sixel=%v)", supported, isKitty(), isInlineImageCapable(), isSixelCapable())
	return supported
}

// PreviewWand takes a MagickWand and tries to display it inline in the terminal.
// It prefers kitty unicode/graphics placement, then the inline images OSC, then Sixel.
// Returns error if unsupported or on failure.
func PreviewWand(wand *imagick.MagickWand) error {
	if wand == nil {
		return fmt.Errorf("nil wand")
	}

	// Log entry and detection state when debugging is enabled
	debugf("PreviewWand called (supported=%v, KITTY=%v, INLINE=%v, SIXEL=%v)", PreviewSupported(), isKitty(), isInlineImageCapable(), isSixelCapable())

	if !PreviewSupported() {
		return fmt.Errorf("no supported terminal preview protocol detected")
	}

	// Clone the wand to avoid mutating the caller's wand (format, etc).
	clone := wand.Clone()
	if clone == nil {
		debugf("failed to clone wand")
		return fmt.Errorf("failed to clone wand")
	}
	defer clone.Destroy()

	// Ensure PNG format for reliable transmission
	if err := clone.SetImageFormat("PNG"); err != nil {
		// SetImageFormat returns error on failure; attempt to continue but report if it fails
		return fmt.Errorf("failed to set PNG format: %w", err)
	}

	blob, err := clone.GetImageBlob()
	if err != nil {
		return fmt.Errorf("GetImageBlob failed: %w", err)
	}
	if len(blob) == 0 {
		return fmt.Errorf("empty image blob")
	}

	// Prefer kitty if available (unicode placeholders / placement)
	if isKitty() {
		debugf("attempting kitty protocol")
		if err := sendKittyPNG(blob); err != nil {
			debugf("kitty protocol failed: %v", err)
			// fallback attempt to inline images if available and kitty failed
			if isInlineImageCapable() {
				debugf("falling back to inline image OSC")
				if err2 := sendInlineImagePNG(blob); err2 == nil {
					debugf("inline image OSC succeeded after kitty failure")
					return nil
				} else {
					debugf("inline image OSC also failed: %v", err2)
				}
			}
			// fallback to Sixel if supported
			if isSixelCapable() {
				debugf("falling back to Sixel rendering")
				if err3 := sendSixelPNG(blob); err3 == nil {
					debugf("sixel succeeded after kitty failure")
					return nil
				} else {
					debugf("sixel also failed: %v", err3)
				}
			}
			return fmt.Errorf("kitty preview failed: %w", err)
		}
		debugf("kitty protocol succeeded")
		return nil
	}

	// If terminal supports inline images OSC (iTerm2-style) prefer that.
	if isInlineImageCapable() {
		debugf("attempting inline image OSC protocol")
		if err := sendInlineImagePNG(blob); err != nil {
			debugf("inline image OSC failed: %v", err)
			// fallback to Sixel if available
			if isSixelCapable() {
				debugf("falling back to Sixel rendering from inline image failure")
				if err2 := sendSixelPNG(blob); err2 == nil {
					debugf("sixel succeeded after inline image failure")
					return nil
				} else {
					debugf("sixel also failed: %v", err2)
				}
			}
			return fmt.Errorf("inline image preview failed: %w", err)
		}
		debugf("inline image OSC succeeded")
		return nil
	}

	// Fallback: try Sixel-capable terminals
	if isSixelCapable() {
		if err := sendSixelPNG(blob); err != nil {
			return fmt.Errorf("sixel preview failed: %w", err)
		}
		return nil
	}

	return fmt.Errorf("no preview protocol matched")
}

// sendKittyPNG pushes PNG bytes to the terminal using the kitty graphics protocol.
// It chunks base64 payload into <=4096-byte chunks per spec. The first chunk includes
// placement parameters to force the image to render into a fixed area (columns x rows).
//
// Placement sizing is controlled by environment variables (optional):
//
//	KITTY_PREVIEW_COLS and KITTY_PREVIEW_ROWS
//
// If those are not present, sensible defaults are used.
//
// Note: we still transmit PNG data (f=100) and a=T to transmit+display. The keys `c` and `r`
// request the image be displayed over the specified number of columns and rows respectively.
// We suppress terminal responses with q=2.
func sendKittyPNG(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("no data")
	}

	debugf("sendKittyPNG preparing to send %d bytes (raw PNG)", len(data))

	enc := base64.StdEncoding.EncodeToString(data)
	const chunkSize = 4096

	// Determine preview placement size from environment (defaults).
	cols := 40
	rows := 20
	if v := os.Getenv("KITTY_PREVIEW_COLS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cols = n
		}
	}
	if v := os.Getenv("KITTY_PREVIEW_ROWS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			rows = n
		}
	}

	debugf("kitty placement: cols=%d rows=%d (requested)", cols, rows)

	stdout := os.Stdout

	// Helper to write a raw sequence to stdout.
	writeSeq := func(s string) error {
		_, err := stdout.Write([]byte(s))
		return err
	}

	total := len(enc)
	first := true
	for pos := 0; pos < total; pos += chunkSize {
		end := pos + chunkSize
		if end > total {
			end = total
		}
		chunk := enc[pos:end]
		last := end == total

		mVal := "0"
		if !last {
			mVal = "1"
		}

		if first {
			// First chunk includes full control keys and placement (c,r).
			// a=T transmit+display, f=100 PNG, t=d direct payload,
			// q=2 suppress responses, c=<cols>, r=<rows> request rendering area.
			header := fmt.Sprintf("\x1b_Ga=T,f=100,t=d,q=2,c=%d,r=%d,m=%s;", cols, rows, mVal)
			header += chunk + "\x1b\\"
			if err := writeSeq(header); err != nil {
				return err
			}
			first = false
			continue
		}

		// Subsequent chunks must contain only m=1/m=0 and the payload chunk.
		header := "\x1b_G" + "m=" + mVal + ";" + chunk + "\x1b\\"
		if err := writeSeq(header); err != nil {
			return err
		}
	}

	// After the image is transmitted, we must print a newline to ensure the cursor
	// is advanced past the image area. Otherwise, subsequent text may be obscured.
	fmt.Println()

	// Done
	return nil
}

// sendInlineImagePNG emits the generic iTerm2-style inline image OSC (1337) sequence.
// Many terminals implement a compatible inline-image OSC (iTerm2, WezTerm, Warp, Tabby, VSCode, etc).
// Format: ESC ] 1337 ; File=inline=1;size=<n> : <base64> BEL
func sendInlineImagePNG(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("no data")
	}
	debugf("sendInlineImagePNG preparing to send %d bytes", len(data))
	enc := base64.StdEncoding.EncodeToString(data)
	seq := "\x1b]1337;File=inline=1;size=" + fmt.Sprintf("%d", len(data)) + ":" + enc + "\a"
	n, err := os.Stdout.Write([]byte(seq))
	debugf("wrote %d bytes to stdout for inline image (err=%v)", n, err)

	// After the image is transmitted, we must print a newline to ensure the cursor
	// is advanced past the image area. Otherwise, subsequent text may be obscured.
	// We print multiple newlines to account for the height of the image.
	for i := 0; i < 20; i++ {
		fmt.Println()
	}

	return err
}

// sendSixelPNG attempts to render PNG data using an external sixel renderer (img2sixel).
// It pipes the PNG bytes to the external tool which is expected to emit sixel to stdout.
// This is a pragmatic approach because implementing a sixel encoder here is beyond scope.
func sendSixelPNG(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("no data")
	}

	debugf("sendSixelPNG attempting img2sixel (or chafa) for %d bytes", len(data))

	// Try to locate a suitable external sixel tool.
	// Common tool: img2sixel (part of libsixel or some distributions).
	// We call it with '-' to accept stdin.
	cmd := exec.Command("img2sixel", "-")
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err == nil {
		debugf("img2sixel succeeded")
		// Ensure the cursor moves to the next line after the image.
		for i := 0; i < 20; i++ {
			fmt.Println()
		}
		return nil
	} else {
		debugf("img2sixel failed: %v", err)
	}

	// If img2sixel isn't available, try chafa as a fallback (chafa supports multiple terminals).
	cmd = exec.Command("chafa", "--fill=block", "--symbols=block", "-s", "auto", "-")
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err == nil {
		debugf("chafa succeeded")
		// Ensure the cursor moves to the next line after the image.
		for i := 0; i < 20; i++ {
			fmt.Println()
		}
		return nil
	} else {
		debugf("chafa failed: %v", err)
	}

	// As a last resort, write a small inline PNG with base64 to the terminal (rarely supported).
	debugf("falling back to inline PNG base64 sequence as last resort")
	enc := base64.StdEncoding.EncodeToString(data)
	seq := "\x1b]1337;File=name=preview.png;inline=1;size=" + fmt.Sprintf("%d", len(data)) + ":" + enc + "\a"
	n, err := os.Stdout.Write([]byte(seq))
	debugf("wrote %d bytes for inline PNG fallback (err=%v)", n, err)

	// Ensure the cursor moves to the next line after the image.
	for i := 0; i < 20; i++ {
		fmt.Println()
	}

	return err
}
