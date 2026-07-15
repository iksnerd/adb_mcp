package android

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ListPackages lists installed package names, optionally filtered by substring.
func ListPackages(ctx context.Context, serial, filter string) ([]string, error) {
	args := []string{"shell", "pm", "list", "packages"}
	out, err := runAdb(ctx, serial, args...)
	if err != nil {
		return nil, err
	}
	needle := strings.ToLower(strings.TrimSpace(filter))
	var pkgs []string
	for _, line := range strings.Split(out, "\n") {
		pkg := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "package:"))
		if pkg == "" {
			continue
		}
		if needle != "" && !strings.Contains(strings.ToLower(pkg), needle) {
			continue
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

// InstallApp installs (or reinstalls, -r) an APK from a local path.
func InstallApp(ctx context.Context, serial, apkPath string) (string, error) {
	if _, err := os.Stat(apkPath); err != nil {
		return "", fmt.Errorf("apk not found: %s", apkPath)
	}
	return runAdb(ctx, serial, "install", "-r", apkPath)
}

// UninstallApp removes an app by package name.
func UninstallApp(ctx context.Context, serial, pkg string) (string, error) {
	return runAdb(ctx, serial, "uninstall", pkg)
}

// LaunchApp starts an app's launcher activity via monkey.
func LaunchApp(ctx context.Context, serial, pkg string) error {
	_, err := runAdb(ctx, serial, "shell", "monkey", "-p", pkg,
		"-c", "android.intent.category.LAUNCHER", "1")
	return err
}

// StopApp force-stops an app.
func StopApp(ctx context.Context, serial, pkg string) error {
	_, err := runAdb(ctx, serial, "shell", "am", "force-stop", pkg)
	return err
}

// ReloadApp attempts to trigger a Metro/JS reload via the classic React
// Native dev-support broadcast receiver (<pkg>.RELOAD_APP_ACTION). This is
// best-effort: the receiver is only registered in debug builds of classic
// (non-bridgeless) RN architectures, so on newer RN/Expo dev clients this
// broadcast may be silently ignored with no error. When it doesn't visibly
// reload the app, fall back to OpenDevMenu + tapping "Reload".
func ReloadApp(ctx context.Context, serial, pkg string) error {
	_, err := runAdb(ctx, serial, "shell", "am", "broadcast", "-a", pkg+".RELOAD_APP_ACTION")
	return err
}

// OpenDevMenu opens the React Native dev menu (KEYCODE_MENU) on the
// foreground app — the standard adb way to reach a dev build's Reload/Debug
// JS Remotely/etc. options. From here, tap_on_text("Reload") (or another menu
// item) drives it.
func OpenDevMenu(ctx context.Context, serial string) error {
	return PressKey(ctx, serial, 82)
}

// ClearAppData wipes an app's data/cache, returning it to a first-launch state.
func ClearAppData(ctx context.Context, serial, pkg string) (string, error) {
	return runAdb(ctx, serial, "shell", "pm", "clear", pkg)
}

// OpenURL opens a URL or deep link via an ACTION_VIEW intent. When pkg is set
// the intent is targeted at that package.
func OpenURL(ctx context.Context, serial, url, pkg string) (string, error) {
	args := []string{"shell", "am", "start", "-a", "android.intent.action.VIEW", "-d", url}
	if strings.TrimSpace(pkg) != "" {
		// Restrict the intent to a package with the -p option. A bare positional
		// argument would be parsed by `am` as the intent DATA URI, clobbering the
		// -d url above and silently opening the wrong thing.
		args = append(args, "-p", pkg)
	}
	return runAdb(ctx, serial, args...)
}

// AppDetails is a compact summary of an installed package.
type AppDetails struct {
	Package          string `json:"package"`
	Installed        bool   `json:"installed"`
	VersionName      string `json:"version_name,omitempty"`
	VersionCode      string `json:"version_code,omitempty"`
	LauncherActivity string `json:"launcher_activity,omitempty"`
}

var (
	versionNameRe = regexp.MustCompile(`versionName=(\S+)`)
	versionCodeRe = regexp.MustCompile(`versionCode=(\d+)`)
)

// GetAppDetails reports an app's version and launchable activity via
// `dumpsys package` + `cmd package resolve-activity`.
func GetAppDetails(ctx context.Context, serial, pkg string) (AppDetails, error) {
	d := AppDetails{Package: pkg}
	dump, err := runAdb(ctx, serial, "shell", "dumpsys", "package", pkg)
	if err != nil {
		return d, err
	}
	if !strings.Contains(dump, "Unable to find package") && strings.Contains(dump, "Package [") {
		d.Installed = true
	}
	if m := versionNameRe.FindStringSubmatch(dump); m != nil {
		d.VersionName = m[1]
	}
	if m := versionCodeRe.FindStringSubmatch(dump); m != nil {
		d.VersionCode = m[1]
	}
	// Best-effort launcher activity.
	if res, err := runAdb(ctx, serial, "shell", "cmd", "package", "resolve-activity",
		"--brief", "-c", "android.intent.category.LAUNCHER", pkg); err == nil {
		for _, line := range strings.Split(res, "\n") {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "/") && !strings.Contains(line, " ") {
				d.LauncherActivity = line
			}
		}
	}
	return d, nil
}
