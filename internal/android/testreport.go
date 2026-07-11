package android

import (
	"encoding/xml"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// TestSummary is the aggregated result of a Gradle test run, parsed from the
// JUnit XML that both JVM unit tests (build/test-results/**) and instrumented
// tests (build/outputs/androidTest-results/**) emit.
type TestSummary struct {
	Suites   int      `json:"suites"`
	Tests    int      `json:"tests"`
	Failures int      `json:"failures"`
	Errors   int      `json:"errors"`
	Skipped  int      `json:"skipped"`
	Passed   int      `json:"passed"`
	Failed   []string `json:"failed,omitempty"` // "Class.method: reason", capped
}

// junitSuite mirrors the <testsuite> element of a JUnit XML report.
type junitSuite struct {
	Tests    int             `xml:"tests,attr"`
	Failures int             `xml:"failures,attr"`
	Errors   int             `xml:"errors,attr"`
	Skipped  int             `xml:"skipped,attr"`
	Cases    []junitTestCase `xml:"testcase"`
}

type junitTestCase struct {
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	Failure   *junitFailure `xml:"failure"`
	Error     *junitFailure `xml:"error"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
}

// maxFailedListed caps how many failing-test lines a summary carries, so a
// wholesale failure doesn't produce an unreadable wall of text.
const maxFailedListed = 20

// ParseTestResults walks projectDir for JUnit XML reports and aggregates them.
// found reports whether any report files were located at all (so callers can
// distinguish "0 tests" from "no report written").
func ParseTestResults(projectDir string) (sum TestSummary, found bool) {
	_ = filepath.WalkDir(projectDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !isTestReport(path) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if s, ok := parseJUnitSuite(data); ok {
			found = true
			sum.merge(s)
		}
		return nil
	})
	sum.Passed = sum.Tests - sum.Failures - sum.Errors - sum.Skipped
	if sum.Passed < 0 {
		sum.Passed = 0
	}
	sort.Strings(sum.Failed)
	if len(sum.Failed) > maxFailedListed {
		extra := len(sum.Failed) - maxFailedListed
		sum.Failed = append(sum.Failed[:maxFailedListed:maxFailedListed],
			fmt.Sprintf("… and %d more", extra))
	}
	return sum, found
}

// isTestReport matches Gradle's JUnit XML output: a TEST-*.xml (or any .xml)
// file under a test-results or androidTest-results directory.
func isTestReport(path string) bool {
	slash := filepath.ToSlash(path)
	if !strings.HasSuffix(slash, ".xml") {
		return false
	}
	return strings.Contains(slash, "/test-results/") ||
		strings.Contains(slash, "/androidTest-results/") ||
		strings.Contains(slash, "/androidTest-results-connected/")
}

// parseJUnitSuite parses one report, which may be a single <testsuite> or a
// <testsuites> wrapper containing several.
func parseJUnitSuite(data []byte) (junitSuite, bool) {
	// Try the <testsuites> wrapper first. Such a wrapper can carry aggregate
	// counts on its root (tests="N" ...) while the per-<testcase> detail lives on
	// the child <testsuite> elements. Parsing it as a single suite would read the
	// root's counts but see no cases, dropping every failing-test name. This
	// unmarshal only matches when there are genuine child <testsuite> elements,
	// so a lone <testsuite> document falls through to the single-suite path.
	var multi struct {
		Suites []junitSuite `xml:"testsuite"`
	}
	if err := xml.Unmarshal(data, &multi); err == nil && len(multi.Suites) > 0 {
		var combined junitSuite
		for _, s := range multi.Suites {
			combined.Tests += s.Tests
			combined.Failures += s.Failures
			combined.Errors += s.Errors
			combined.Skipped += s.Skipped
			combined.Cases = append(combined.Cases, s.Cases...)
		}
		return combined, true
	}
	var single junitSuite
	if err := xml.Unmarshal(data, &single); err == nil && (single.Tests > 0 || len(single.Cases) > 0) {
		return single, true
	}
	return junitSuite{}, false
}

func (sum *TestSummary) merge(s junitSuite) {
	sum.Suites++
	sum.Tests += s.Tests
	sum.Failures += s.Failures
	sum.Errors += s.Errors
	sum.Skipped += s.Skipped
	for _, c := range s.Cases {
		fail := c.Failure
		if fail == nil {
			fail = c.Error
		}
		if fail == nil {
			continue
		}
		name := c.Name
		if c.Classname != "" {
			name = c.Classname + "." + c.Name
		}
		if msg := strings.TrimSpace(firstLine(fail.Message)); msg != "" {
			name += ": " + msg
		}
		sum.Failed = append(sum.Failed, name)
	}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// String renders a one-line-plus-detail human summary for a tool result.
func (sum TestSummary) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d tests, %d passed, %d failed, %d errors, %d skipped (%d suite(s))",
		sum.Tests, sum.Passed, sum.Failures, sum.Errors, sum.Skipped, sum.Suites)
	for _, f := range sum.Failed {
		b.WriteString("\n  ✗ " + f)
	}
	return b.String()
}
