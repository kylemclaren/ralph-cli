package pidfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

const DefaultPIDFileName = ".ralph.pid"

var (
	ErrNotRunning     = errors.New("ralph is not running")
	ErrAlreadyRunning = errors.New("ralph is already running")
)

// PIDFile manages the Ralph process ID file
type PIDFile struct {
	path string
}

// New creates a new PIDFile manager
// If dir is empty, uses the current working directory
func New(dir string) *PIDFile {
	if dir == "" {
		dir, _ = os.Getwd()
	}
	return &PIDFile{
		path: filepath.Join(dir, DefaultPIDFileName),
	}
}

// Path returns the path to the PID file
func (p *PIDFile) Path() string {
	return p.path
}

// Write writes the current process ID to the PID file
func (p *PIDFile) Write() error {
	// Check if already running
	if pid, err := p.Read(); err == nil {
		if isProcessRunning(pid) {
			return fmt.Errorf("%w: PID %d", ErrAlreadyRunning, pid)
		}
		// Stale PID file, remove it
		_ = p.Remove()
	}

	pid := os.Getpid()
	return os.WriteFile(p.path, []byte(strconv.Itoa(pid)), 0644)
}

// Read reads the PID from the file
func (p *PIDFile) Read() (int, error) {
	data, err := os.ReadFile(p.path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, ErrNotRunning
		}
		return 0, err
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("invalid PID file: %w", err)
	}

	return pid, nil
}

// Remove removes the PID file
func (p *PIDFile) Remove() error {
	err := os.Remove(p.path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// IsRunning checks if Ralph is currently running
func (p *PIDFile) IsRunning() (bool, int) {
	pid, err := p.Read()
	if err != nil {
		return false, 0
	}
	return isProcessRunning(pid), pid
}

// Signal sends a signal to the running Ralph process
func (p *PIDFile) Signal(sig syscall.Signal) error {
	pid, err := p.Read()
	if err != nil {
		return err
	}

	if !isProcessRunning(pid) {
		// Clean up stale PID file
		_ = p.Remove()
		return ErrNotRunning
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Signal(sig)
}

// Stop sends SIGTERM to gracefully stop the running Ralph process
func (p *PIDFile) Stop() error {
	return p.Signal(syscall.SIGTERM)
}

// Kill sends SIGKILL to forcefully stop the running Ralph process
func (p *PIDFile) Kill() error {
	return p.Signal(syscall.SIGKILL)
}

// isProcessRunning checks if a process with the given PID is running
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Sending signal 0 checks if process exists without actually sending a signal
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
