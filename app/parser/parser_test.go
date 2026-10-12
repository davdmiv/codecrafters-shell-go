package parser

import (
	"reflect"
	"slices"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr string
	}{
		{
			name:  "simple words",
			input: "echo hello world",
			want:  []string{"echo", "hello", "world"},
		},
		{
			name:  "multiple spaces",
			input: "echo   hello    world",
			want:  []string{"echo", "hello", "world"},
		},
		{
			name:  "leading and trailing spaces",
			input: "  echo hello  ",
			want:  []string{"echo", "hello"},
		},
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "whitespace only",
			input: "   \t  ",
			want:  nil,
		},
		{
			name:  "double quotes preserve spaces",
			input: `echo "hello world"`,
			want:  []string{"echo", "hello world"},
		},
		{
			name:  "single quotes preserve spaces",
			input: "echo 'hello world'",
			want:  []string{"echo", "hello world"},
		},
		{
			name:  "empty double quotes",
			input: `echo ""`,
			want:  []string{"echo", ""},
		},
		{
			name:  "empty single quotes",
			input: "echo ''",
			want:  []string{"echo", ""},
		},
		{
			name:  "adjacent quoted and unquoted",
			input: `echo "hello"world`,
			want:  []string{"echo", "helloworld"},
		},
		{
			name:  "backslash escapes space",
			input: `echo hello\ world`,
			want:  []string{"echo", "hello world"},
		},
		{
			name:  "backslash escapes quote",
			input: `echo \"hello\"`,
			want:  []string{"echo", `"hello"`},
		},
		{
			name:  "backslash inside double quotes",
			input: `echo "hello\"world"`,
			want:  []string{"echo", `hello"world`},
		},
		{
			name:  "backslash inside single quotes is literal",
			input: `echo 'hello\world'`,
			want:  []string{"echo", `hello\world`},
		},
		{
			name:  "single quotes inside double quotes",
			input: `echo "it's fine"`,
			want:  []string{"echo", "it's fine"},
		},
		{
			name:  "double quotes inside single quotes",
			input: `echo 'say "hi"'`,
			want:  []string{"echo", `say "hi"`},
		},
		{
			name:    "unterminated double quote",
			input:   `echo "hello`,
			wantErr: "unterminated double quote",
		},
		{
			name:    "unterminated single quote",
			input:   "echo 'hello",
			wantErr: "unterminated single quote",
		},
		{
			name:    "unterminated escape",
			input:   `echo hello\`,
			wantErr: "unterminated escape",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := tokenize(tt.input)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("tokenize(%q) error = nil, want %q", tt.input, tt.wantErr)
				}
				if got := err.Error(); got != tt.wantErr {
					t.Fatalf("tokenize(%q) error = %q, want %q", tt.input, got, tt.wantErr)
				}
				if tokens != nil {
					t.Fatalf("tokenize(%q) tokens = %v, want nil on error", tt.input, tokens)
				}
				return
			}

			if err != nil {
				t.Fatalf("tokenize(%q) unexpected error: %v", tt.input, err)
			}

			got := make([]string, len(tokens))
			for i, token := range tokens {
				if token.Type != TokenArg {
					t.Fatalf("tokenize(%q) token %d type = %v, want %v", tt.input, i, token.Type, TokenArg)
				}
				got[i] = token.Value
			}
			if !slices.Equal(got, tt.want) {
				t.Fatalf("tokenize(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Command
		wantErr string
	}{
		{
			name:  "command and arguments",
			input: "echo hello world",
			want:  &Command{Name: "echo", Args: []string{"hello", "world"}},
		},
		{
			name:  "command without arguments",
			input: "pwd",
			want:  &Command{Name: "pwd"},
		},
		{
			name:  "output redirection",
			input: "echo hello > out.txt",
			want: &Command{
				Name: "echo",
				Args: []string{"hello"},
				Redirections: []Redirection{{
					Type: TokenRedirectOut, FilePath: "out.txt",
				}},
			},
		},
		{
			name:  "redirection adjacent to argument",
			input: "echo hello>out.txt",
			want: &Command{
				Name: "echo",
				Args: []string{"hello"},
				Redirections: []Redirection{{
					Type: TokenRedirectOut, FilePath: "out.txt",
				}},
			},
		},
		{
			name:  "quoted redirection-like argument",
			input: `echo ">"`,
			want:  &Command{Name: "echo", Args: []string{">"}},
		},
		{
			name:  "escaped redirection-like argument",
			input: `echo \>`,
			want:  &Command{Name: "echo", Args: []string{">"}},
		},
		{
			name:  "quoted redirection filename",
			input: `echo hello>"out file.txt"`,
			want: &Command{
				Name: "echo",
				Args: []string{"hello"},
				Redirections: []Redirection{{
					Type: TokenRedirectOut, FilePath: "out file.txt",
				}},
			},
		},
		{
			name:  "explicit stdout file descriptor",
			input: "echo 1>out.txt",
			want: &Command{
				Name: "echo",
				Redirections: []Redirection{{
					Type: TokenRedirectOut, FilePath: "out.txt",
				}},
			},
		},
		{
			name:  "stderr redirection",
			input: "cmd 2>errors.log",
			want: &Command{
				Name: "cmd",
				Redirections: []Redirection{{
					Type: TokenRedirectErr, FilePath: "errors.log",
				}},
			},
		},
		{
			name:  "append stderr redirection",
			input: "cmd 2>>errors.log",
			want: &Command{
				Name: "cmd",
				Redirections: []Redirection{{
					Type: TokenRedirectAppendErr, FilePath: "errors.log",
				}},
			},
		},
		{
			name:  "append and input redirections",
			input: "cat < in.txt >> out.txt",
			want: &Command{
				Name: "cat",
				Redirections: []Redirection{
					{Type: TokenRedirectIn, FilePath: "in.txt"},
					{Type: TokenRedirectAppend, FilePath: "out.txt"},
				},
			},
		},
		{
			name:    "empty command",
			input:   "  \t ",
			wantErr: "empty command",
		},
		{
			name:    "missing output redirection file",
			input:   "echo hello >",
			wantErr: "missing file for redirection >",
		},
		{
			name:    "missing append redirection file",
			input:   "echo hello >>",
			wantErr: "missing file for redirection >>",
		},
		{
			name:    "missing input redirection file",
			input:   "cat <",
			wantErr: "missing file for redirection <",
		},
		{
			name:    "missing stderr redirection file",
			input:   "cmd 2>",
			wantErr: "missing file for redirection 2>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)

			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("Parse(%q) error = %v, want %q", tt.input, err, tt.wantErr)
				}
				if got != nil {
					t.Fatalf("Parse(%q) command = %#v, want nil on error", tt.input, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tt.input, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Parse(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}
