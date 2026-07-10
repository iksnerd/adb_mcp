package tools

import (
	"context"
	"fmt"
	"strings"

	"AndroidEmulatorMCP/internal/android"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---- Arguments ----

type gradleArgs struct {
	ProjectDir string   `json:"project_dir" jsonschema:"Path to the Android project root containing the Gradle wrapper (gradlew)."`
	Task       string   `json:"task,omitempty" jsonschema:"Gradle task to run. Defaults to the tool's standard task."`
	Args       []string `json:"args,omitempty" jsonschema:"Extra arguments passed to Gradle (e.g. --stacktrace, -Pflavor=free)."`
}

// ---- Handlers ----

func gradleBuild(ctx context.Context, in gradleArgs) (*mcp.CallToolResult, error) {
	task := in.Task
	if task == "" {
		task = "assembleDebug"
	}
	out, err := android.Gradle(ctx, in.ProjectDir, append([]string{task}, in.Args...)...)
	if err != nil {
		return nil, fmt.Errorf("%v\n%s", err, tailLines(out, 40))
	}
	apks := android.FindAPKs(in.ProjectDir)
	msg := "Build succeeded (" + task + ")."
	if len(apks) > 0 {
		msg += "\nAPK(s):\n" + strings.Join(apks, "\n")
	}
	return text("%s\n\n%s", msg, tailLines(out, 20)), nil
}

func runUnitTests(ctx context.Context, in gradleArgs) (*mcp.CallToolResult, error) {
	return runGradleReporting(ctx, in, "test")
}

func runInstrumentedTests(ctx context.Context, in gradleArgs) (*mcp.CallToolResult, error) {
	return runGradleReporting(ctx, in, "connectedAndroidTest")
}

func runGradleReporting(ctx context.Context, in gradleArgs, defaultTask string) (*mcp.CallToolResult, error) {
	task := in.Task
	if task == "" {
		task = defaultTask
	}
	out, err := android.Gradle(ctx, in.ProjectDir, append([]string{task}, in.Args...)...)
	// Parse the JUnit XML regardless of exit code: a non-zero Gradle exit is
	// exactly when the per-test breakdown (which tests failed and why) is most
	// useful, so surface it in both the success and failure paths.
	summary, found := android.ParseTestResults(in.ProjectDir)
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		if found {
			msg += "\n\n" + summary.String()
		}
		return nil, fmt.Errorf("%s\n\n%s", msg, tailLines(out, 60))
	}
	if found {
		return text("Tests passed (%s).\n\n%s\n\n%s", task, summary.String(), tailLines(out, 20)), nil
	}
	return text("Tests passed (%s).\n\n%s", task, tailLines(out, 30)), nil
}

func listGradleTasks(ctx context.Context, in gradleArgs) (*mcp.CallToolResult, error) {
	out, err := android.Gradle(ctx, in.ProjectDir, "tasks")
	if err != nil {
		return nil, fmt.Errorf("%v\n%s", err, tailLines(out, 40))
	}
	return text("%s", tailLines(out, 120)), nil
}

// tailLines keeps the last n non-trivial lines of possibly-huge tool output
// (Gradle logs) so results stay readable.
func tailLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) <= n {
		return strings.Join(lines, "\n")
	}
	return fmt.Sprintf("… (%d earlier lines omitted)\n%s", len(lines)-n, strings.Join(lines[len(lines)-n:], "\n"))
}
