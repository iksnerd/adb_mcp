package android

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTestResults(t *testing.T) {
	dir := t.TempDir()
	// A passing+failing JVM unit-test report.
	unit := filepath.Join(dir, "build", "test-results", "testDebugUnitTest")
	if err := os.MkdirAll(unit, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(unit, "TEST-com.example.MathTest.xml"), `<?xml version="1.0"?>
<testsuite name="com.example.MathTest" tests="3" failures="1" errors="0" skipped="1">
  <testcase name="adds" classname="com.example.MathTest"/>
  <testcase name="divides" classname="com.example.MathTest">
    <failure message="expected 2 but was 3&#10;stacktrace...">assert</failure>
  </testcase>
  <testcase name="skipMe" classname="com.example.MathTest"><skipped/></testcase>
</testsuite>`)
	// A second suite via the <testsuites> wrapper form.
	inst := filepath.Join(dir, "build", "outputs", "androidTest-results", "connected")
	if err := os.MkdirAll(inst, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(inst, "TEST-emulator.xml"), `<?xml version="1.0"?>
<testsuites>
  <testsuite name="UiTest" tests="2" failures="0" errors="1" skipped="0">
    <testcase name="loads" classname="com.example.UiTest"/>
    <testcase name="crashes" classname="com.example.UiTest">
      <error message="NullPointerException">boom</error>
    </testcase>
  </testsuite>
</testsuites>`)
	// A non-report xml that must be ignored.
	writeFile(t, filepath.Join(dir, "AndroidManifest.xml"), `<manifest/>`)

	sum, found := ParseTestResults(dir)
	if !found {
		t.Fatal("expected reports to be found")
	}
	if sum.Suites != 2 {
		t.Errorf("Suites = %d, want 2", sum.Suites)
	}
	if sum.Tests != 5 {
		t.Errorf("Tests = %d, want 5", sum.Tests)
	}
	if sum.Failures != 1 {
		t.Errorf("Failures = %d, want 1", sum.Failures)
	}
	if sum.Errors != 1 {
		t.Errorf("Errors = %d, want 1", sum.Errors)
	}
	if sum.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", sum.Skipped)
	}
	if sum.Passed != 2 { // 5 - 1 failure - 1 error - 1 skipped
		t.Errorf("Passed = %d, want 2", sum.Passed)
	}
	if len(sum.Failed) != 2 {
		t.Fatalf("Failed count = %d, want 2 (%v)", len(sum.Failed), sum.Failed)
	}
	// Failure message should be truncated to its first line and prefixed with class.method.
	wantContains := "com.example.MathTest.divides: expected 2 but was 3"
	found2 := false
	for _, f := range sum.Failed {
		if f == wantContains {
			found2 = true
		}
	}
	if !found2 {
		t.Errorf("Failed = %v, want an entry %q", sum.Failed, wantContains)
	}
}

func TestParseTestResultsNone(t *testing.T) {
	if _, found := ParseTestResults(t.TempDir()); found {
		t.Error("expected found=false for a dir with no reports")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
