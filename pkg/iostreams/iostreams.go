package iostreams

import (
	"bytes"
	"io"
	"os"
)

// IOStreams provides access to standard I/O streams. Commands should use
// these instead of directly referencing os.Stdin/Stdout/Stderr.
type IOStreams struct {
	In     io.ReadCloser
	Out    io.Writer
	ErrOut io.Writer

	inTTY  bool
	outTTY bool
	errTTY bool
}

// System creates IOStreams using real os.Stdin, os.Stdout, and os.Stderr.
func System() *IOStreams {
	return &IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
		inTTY:  isTerminal(os.Stdin),
		outTTY: isTerminal(os.Stdout),
		errTTY: isTerminal(os.Stderr),
	}
}

// Test creates IOStreams backed by bytes.Buffer for testing. Returns the
// IOStreams and the underlying buffers for in, out, and err.
func Test() (*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	return &IOStreams{
		In:     io.NopCloser(in),
		Out:    out,
		ErrOut: errBuf,
	}, in, out, errBuf
}

// IsStdinTTY returns true if stdin is a terminal.
func (s *IOStreams) IsStdinTTY() bool {
	return s.inTTY
}

// IsStdoutTTY returns true if stdout is a terminal.
func (s *IOStreams) IsStdoutTTY() bool {
	return s.outTTY
}

// IsStderrTTY returns true if stderr is a terminal.
func (s *IOStreams) IsStderrTTY() bool {
	return s.errTTY
}

// SetStdinTTY overrides the stdin TTY detection (for testing).
func (s *IOStreams) SetStdinTTY(isTTY bool) {
	s.inTTY = isTTY
}

// SetStdoutTTY overrides the stdout TTY detection (for testing).
func (s *IOStreams) SetStdoutTTY(isTTY bool) {
	s.outTTY = isTTY
}

// SetStderrTTY overrides the stderr TTY detection (for testing).
func (s *IOStreams) SetStderrTTY(isTTY bool) {
	s.errTTY = isTTY
}

// ColorEnabled returns true if color output should be used.
// Color is disabled when NO_COLOR env var is set or stdout is not a TTY.
func (s *IOStreams) ColorEnabled() bool {
	return os.Getenv("NO_COLOR") == "" && s.outTTY
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
