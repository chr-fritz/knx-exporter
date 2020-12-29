package cmd

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunCompletion(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantOut string
		wantErr bool
	}{
		{"no shell", []string{}, "", true},
		{"too many args", []string{"bash", "test"}, "", true},
		{"invalid shell", []string{"invalid_shell"}, "", true},
		{"zsh", []string{"zsh"}, "#compdef knx-exporter\n", false},
		{"bash", []string{"bash"}, "# bash completion for knx-exporter", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			rCmd := &cobra.Command{Use: "knx-exporter"}
			cmd := NewCompletionCmd()
			rCmd.AddCommand(cmd)
			if err := RunCompletion(out, cmd, tt.args); (err != nil) != tt.wantErr {
				t.Errorf("RunCompletion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotOut := out.String(); !strings.HasPrefix(gotOut, tt.wantOut) {
				t.Errorf("RunCompletion() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

func Test_runCompletionZsh(t *testing.T) {
	tests := []struct {
		name    string
		writer  testWriter
		wantOut string
		wantErr bool
	}{
		{"ok", &bytes.Buffer{}, "#compdef knx-exporter\n", false},
		{"write fails", &errorWriter{}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rCmd := &cobra.Command{Use: "knx-exporter"}
			cmd := NewCompletionCmd()
			rCmd.AddCommand(cmd)

			if err := runCompletionZsh(tt.writer, cmd); (err != nil) != tt.wantErr {
				t.Errorf("runCompletionZsh() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOut := tt.writer.String(); !strings.HasPrefix(gotOut, tt.wantOut) {
				t.Errorf("runCompletionZsh() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

type testWriter interface {
	io.Writer
	fmt.Stringer
}

type errorWriter struct {
}

func (errorWriter) String() string {
	return ""
}

func (errorWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("just a dummy error")
}
