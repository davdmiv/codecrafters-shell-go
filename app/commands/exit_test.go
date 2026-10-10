package commands

import (
	"bytes"
	"errors"
	"testing"
)

func TestExit_Name(t *testing.T) {
	if got, want := NewExit().Name(), "exit"; got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestExit_Execute(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantErr  string
	}{
		{
			name:     "default code",
			args:     nil,
			wantCode: 0,
		},
		{
			name:     "numeric code",
			args:      []string{"2"},
			wantCode: 2,
		},
		{
			name:    "non numeric argument",
			args:    []string{"abc"},
			wantErr: "exit: abc: numeric argument required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			var errBuf bytes.Buffer
			ctx := &Context{Stdout: &out, Stderr: &errBuf}

			err := NewExit().Execute(ctx, tt.args)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Execute: error = nil, want %q", tt.wantErr)
				}
				if got := err.Error(); got != tt.wantErr {
					t.Fatalf("Execute: error = %q, want %q", got, tt.wantErr)
				}
				var exitErr *ExitError
				if errors.As(err, &exitErr) {
					t.Fatalf("Execute: got ExitError, want validation error")
				}
			} else {
				var exitErr *ExitError
				if !errors.As(err, &exitErr) {
					t.Fatalf("Execute: want ExitError, got %v", err)
				}
				if exitErr.Code != tt.wantCode {
					t.Errorf("ExitError.Code = %d, want %d", exitErr.Code, tt.wantCode)
				}
			}

			if got := out.String(); got != "" {
				t.Errorf("stdout = %q, want empty", got)
			}
			if got := errBuf.String(); got != "" {
				t.Errorf("stderr = %q, want empty", got)
			}
		})
	}
}
