package android

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Capture sessions are stateful: start spawns a background process that keeps
// running across tool calls, stop tears it down and returns what it collected.
// Keyed by resolved device serial.

type logSession struct {
	cmd  *exec.Cmd
	file *os.File
	path string
}

var (
	logMu       sync.Mutex
	logSessions = map[string]*logSession{}
)

// StartLogcatCapture begins streaming logcat for serial into a temp file, so a
// later StopLogcatCapture returns everything logged during a flow (unlike the
// one-shot `logcat` dump). Optionally clears the buffer first.
func StartLogcatCapture(ctx context.Context, serial string, clear bool) error {
	logMu.Lock()
	defer logMu.Unlock()
	if _, ok := logSessions[serial]; ok {
		return fmt.Errorf("a logcat capture is already running for %s; stop it first", serial)
	}
	if clear {
		_, _ = runAdb(ctx, serial, "logcat", "-c")
	}
	f, err := os.CreateTemp("", "aemcp-logcat-*.txt")
	if err != nil {
		return err
	}
	args := []string{}
	if serial != "" {
		args = append(args, "-s", serial)
	}
	args = append(args, "logcat", "-v", "time")
	cmd := exec.Command(adbPath(), args...)
	cmd.Env = commandEnv()
	cmd.Stdout = f
	cmd.Stderr = f
	detach(cmd)
	if err := cmd.Start(); err != nil {
		f.Close()
		os.Remove(f.Name())
		return fmt.Errorf("start logcat capture: %w", err)
	}
	logSessions[serial] = &logSession{cmd: cmd, file: f, path: f.Name()}
	return nil
}

// StopLogcatCapture stops the running capture for serial and returns the
// collected lines, optionally filtered (case-insensitive) and with chatty spam
// removed.
func StopLogcatCapture(serial, filter string) (string, error) {
	logMu.Lock()
	s, ok := logSessions[serial]
	if ok {
		delete(logSessions, serial)
	}
	logMu.Unlock()
	if !ok {
		return "", fmt.Errorf("no logcat capture running for %s", serial)
	}
	_ = s.cmd.Process.Kill()
	_, _ = s.cmd.Process.Wait()
	s.file.Close()
	data, err := os.ReadFile(s.path)
	os.Remove(s.path)
	if err != nil {
		return "", err
	}
	needle := strings.ToLower(strings.TrimSpace(filter))
	var kept []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "chatty") {
			continue
		}
		if needle != "" && !strings.Contains(strings.ToLower(line), needle) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.TrimRight(strings.Join(kept, "\n"), "\n"), nil
}

// StopAllCaptures tears down every running logcat/screen-record session,
// killing the detached adb client processes and removing the logcat temp files.
// Call it on server shutdown so an interrupted session does not leak an adb
// process or a /tmp file. (On-device artifacts from an interrupted screen
// recording are left as-is; a clean stop_screen_record removes them.)
func StopAllCaptures() {
	logMu.Lock()
	for serial, s := range logSessions {
		_ = s.cmd.Process.Kill()
		_, _ = s.cmd.Process.Wait()
		s.file.Close()
		os.Remove(s.path)
		delete(logSessions, serial)
	}
	logMu.Unlock()

	recMu.Lock()
	for serial, s := range recSessions {
		_ = s.cmd.Process.Kill()
		_, _ = s.cmd.Process.Wait()
		delete(recSessions, serial)
	}
	recMu.Unlock()
}

type recordSession struct {
	cmd        *exec.Cmd
	devicePath string
}

var (
	recMu       sync.Mutex
	recSessions = map[string]*recordSession{}
)

// StartScreenRecord starts screenrecord on the device (max ~180s per Android's
// limit). StopScreenRecord finalizes and pulls the mp4 to localPath.
func StartScreenRecord(ctx context.Context, serial string) error {
	recMu.Lock()
	defer recMu.Unlock()
	if _, ok := recSessions[serial]; ok {
		return fmt.Errorf("a screen recording is already running for %s; stop it first", serial)
	}
	devicePath := "/sdcard/aemcp-record.mp4"
	args := []string{}
	if serial != "" {
		args = append(args, "-s", serial)
	}
	args = append(args, "shell", "screenrecord", devicePath)
	cmd := exec.Command(adbPath(), args...)
	cmd.Env = commandEnv()
	detach(cmd)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start screenrecord: %w", err)
	}
	recSessions[serial] = &recordSession{cmd: cmd, devicePath: devicePath}
	return nil
}

// StopScreenRecord stops the recording (SIGINT so the mp4 is finalized) and
// pulls it to localPath.
func StopScreenRecord(ctx context.Context, serial, localPath string) (string, error) {
	recMu.Lock()
	s, ok := recSessions[serial]
	if ok {
		delete(recSessions, serial)
	}
	recMu.Unlock()
	if !ok {
		return "", fmt.Errorf("no screen recording running for %s", serial)
	}
	// Interrupt screenrecord ON THE DEVICE so it flushes the moov atom, then
	// wait for the file to finalize before pulling.
	_, _ = runAdb(ctx, serial, "shell", "pkill", "-INT", "screenrecord")
	time.Sleep(1500 * time.Millisecond)
	_ = s.cmd.Process.Kill()
	_, _ = s.cmd.Process.Wait()
	if _, err := PullFile(ctx, serial, s.devicePath, localPath); err != nil {
		return "", err
	}
	_, _ = runAdb(ctx, serial, "shell", "rm", "-f", s.devicePath)
	return localPath, nil
}
