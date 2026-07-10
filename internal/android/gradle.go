package android

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Gradle runs one or more Gradle tasks in projectDir using its wrapper
// (./gradlew) and returns combined stdout+stderr. Gradle runs on the host, not
// on a device, so no serial is involved.
func Gradle(ctx context.Context, projectDir string, args ...string) (string, error) {
	if strings.TrimSpace(projectDir) == "" {
		return "", fmt.Errorf("project_dir is required (path to the Android project root containing gradlew)")
	}
	gradlew := filepath.Join(projectDir, wrapperName())
	if !fileExists(gradlew) {
		return "", fmt.Errorf("%s not found in %s — point project_dir at the module/root that has the Gradle wrapper", wrapperName(), projectDir)
	}
	cmd := exec.CommandContext(ctx, gradlew, args...)
	cmd.Dir = projectDir
	cmd.Env = commandEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("gradle %s failed: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}

// FindAPKs returns any .apk files under projectDir's build outputs, newest-path
// first, so a build tool can report where the artifact landed.
func FindAPKs(projectDir string) []string {
	var apks []string
	_ = filepath.WalkDir(projectDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".apk") && strings.Contains(filepath.ToSlash(path), "/build/outputs/") {
			apks = append(apks, path)
		}
		return nil
	})
	sort.Strings(apks)
	return apks
}

func wrapperName() string {
	if runtime.GOOS == "windows" {
		return "gradlew.bat"
	}
	return "gradlew"
}
