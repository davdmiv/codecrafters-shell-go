package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"
)

func newTestShell(input string) (*shell, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	sh, err := newShell(strings.NewReader(input), stdout, stderr)
	if err != nil {
		panic(err)
	}
	return sh, stdout, stderr
}

func TestProcessInputReportsBuiltinType(t *testing.T) {
	sh, stdout, stderr := newTestShell("")

	shouldExit := sh.processInput("type echo")

	if shouldExit {
		t.Fatal("processInput(type echo) requested exit")
	}
	if got, want := stdout.String(), "echo is a shell builtin\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestProcessInputEchoesArguments(t *testing.T) {
	sh, stdout, stderr := newTestShell("")

	shouldExit := sh.processInput("echo hello shell")

	if shouldExit {
		t.Fatal("processInput(echo) requested exit")
	}
	if got, want := stdout.String(), "hello shell\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestProcessInputEchoesSingleQuotedArguments(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "preserves whitespace and special characters",
			input: "echo 'shell $*~\\    hello\tworld'",
			want:  "shell $*~\\    hello\tworld\n",
		},
		{
			name:  "concatenates adjacent quoted strings",
			input: "echo 'hello''world'",
			want:  "helloworld\n",
		},
		{
			name:  "concatenates empty quotes with unquoted text",
			input: "echo hello''world",
			want:  "helloworld\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			sh, stdout, stderr := newTestShell("")

			shouldExit := sh.processInput(testCase.input)

			if shouldExit {
				t.Fatal("processInput(echo) requested exit")
			}
			if got := stdout.String(); got != testCase.want {
				t.Errorf("stdout = %q, want %q", got, testCase.want)
			}
			if got := stderr.String(); got != "" {
				t.Errorf("stderr = %q, want empty", got)
			}
		})
	}
}

func TestProcessInputEchoesDoubleQuotedArguments(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "preserves whitespace",
			input: "echo \"quz  hello\"  \"bar\"",
			want:  "quz  hello bar\n",
		},
		{
			name:  "treats single quotes literally",
			input: "echo \"bar\"  \"shell's\"  \"foo\"",
			want:  "bar shell's foo\n",
		},
		{
			name:  "concatenates adjacent double quoted strings",
			input: "echo \"hello\"\"world\"",
			want:  "helloworld\n",
		},
		{
			name:  "concatenates quoted and unquoted strings",
			input: "echo \"hello\"world",
			want:  "helloworld\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			sh, stdout, stderr := newTestShell("")

			shouldExit := sh.processInput(testCase.input)

			if shouldExit {
				t.Fatal("processInput(echo) requested exit")
			}
			if got := stdout.String(); got != testCase.want {
				t.Errorf("stdout = %q, want %q", got, testCase.want)
			}
			if got := stderr.String(); got != "" {
				t.Errorf("stderr = %q, want empty", got)
			}
		})
	}
}

func TestProcessInputCatsDoubleQuotedFileNames(t *testing.T) {
	directory := t.TempDir()
	firstFile := filepath.Join(directory, "file name")
	secondFile := filepath.Join(directory, "'file name' with spaces")
	if err := os.WriteFile(firstFile, []byte("content1 "), 0o644); err != nil {
		t.Fatalf("create first file: %v", err)
	}
	if err := os.WriteFile(secondFile, []byte("content2"), 0o644); err != nil {
		t.Fatalf("create second file: %v", err)
	}

	sh, stdout, stderr := newTestShell("")
	sh.processInput("cat \"" + firstFile + "\" \"" + secondFile + "\"")

	if got, want := stdout.String(), "content1 content2"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestProcessInputEchoesBackslashEscapedCharactersOutsideQuotes(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "preserves escaped spaces",
			input: "echo multiple\\ \\ \\ \\ spaces",
			want:  "multiple    spaces\n",
		},
		{
			name:  "collapses unescaped spaces after an escaped space",
			input: "echo before\\     after",
			want:  "before  after\n",
		},
		{
			name:  "removes the escape before an ordinary character",
			input: "echo test\\nexample",
			want:  "testnexample\n",
		},
		{
			name:  "escapes a backslash",
			input: "echo hello\\\\world",
			want:  "hello\\world\n",
		},
		{
			name:  "escapes quote characters",
			input: "echo \\'\\\"literal quotes\\\"\\'",
			want:  "'\"literal quotes\"'\n",
		},
		{
			name:  "escapes an otherwise ordinary character",
			input: "echo ignore\\_backslash",
			want:  "ignore_backslash\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			sh, stdout, stderr := newTestShell("")
			sh.processInput(testCase.input)

			if got := stdout.String(); got != testCase.want {
				t.Errorf("stdout = %q, want %q", got, testCase.want)
			}
			if got := stderr.String(); got != "" {
				t.Errorf("stderr = %q, want empty", got)
			}
		})
	}
}

func TestProcessInputRejectsTrailingBackslash(t *testing.T) {
	sh, stdout, stderr := newTestShell("")

	shouldExit := sh.processInput("echo trailing\\")

	if shouldExit {
		t.Fatal("processInput(echo) requested exit")
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got, want := stderr.String(), "unterminated escape\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestProcessInputTreatsBackslashesInsideSingleQuotesLiterally(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "preserves multiple backslashes",
			input: "echo 'multiple\\\\slashes'",
			want:  "multiple\\\\slashes\n",
		},
		{
			name:  "preserves backslashes before double quotes",
			input: "echo 'every\\\"thing_is\\\"literal'",
			want:  "every\\\"thing_is\\\"literal\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			sh, stdout, stderr := newTestShell("")
			sh.processInput(testCase.input)

			if got := stdout.String(); got != testCase.want {
				t.Errorf("stdout = %q, want %q", got, testCase.want)
			}
			if got := stderr.String(); got != "" {
				t.Errorf("stderr = %q, want empty", got)
			}
		})
	}
}

func TestProcessInputEscapesQuotesAndBackslashesInsideDoubleQuotes(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "backslash escapes itself",
			input: "echo \"A \\\\ escapes itself\"",
			want:  "A \\ escapes itself\n",
		},
		{
			name:  "backslash escapes a double quote",
			input: "echo \"A \\\" inside double quotes\"",
			want:  "A \" inside double quotes\n",
		},
		{
			name:  "preserves single quotes and reduces an escaped backslash",
			input: "echo \"just'one'\\\\n'backslash\"",
			want:  "just'one'\\n'backslash\n",
		},
		{
			name:  "concatenates text following a closed double quote",
			input: "echo \"inside\\\"literal_quote.\"outside\\\"",
			want:  "inside\"literal_quote.outside\"\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			sh, stdout, stderr := newTestShell("")
			sh.processInput(testCase.input)

			if got := stdout.String(); got != testCase.want {
				t.Errorf("stdout = %q, want %q", got, testCase.want)
			}
			if got := stderr.String(); got != "" {
				t.Errorf("stderr = %q, want empty", got)
			}
		})
	}
}

func TestProcessInputCatsDoubleQuotedBackslashFileNames(t *testing.T) {
	directory := t.TempDir()
	firstFile := filepath.Join(directory, "number 1")
	secondFile := filepath.Join(directory, "doublequote \" 2")
	thirdFile := filepath.Join(directory, "backslash \\ 3")
	for path, content := range map[string]string{
		firstFile:  "content1 ",
		secondFile: "content2 ",
		thirdFile:  "content3",
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("create file %q: %v", path, err)
		}
	}

	sh, stdout, stderr := newTestShell("")
	sh.processInput("cat " + directory + "/\"number 1\" " + directory + "/\"doublequote \\\" 2\" " + directory + "/\"backslash \\\\ 3\"")

	if got, want := stdout.String(), "content1 content2 content3"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestProcessInputCatsSingleQuotedBackslashFileNames(t *testing.T) {
	directory := t.TempDir()
	firstFile := filepath.Join(directory, "no slash 1")
	secondFile := filepath.Join(directory, "one slash \\2")
	thirdFile := filepath.Join(directory, "two slashes \\3\\")
	for path, content := range map[string]string{
		firstFile:  "content1 ",
		secondFile: "content2 ",
		thirdFile:  "content3",
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("create file %q: %v", path, err)
		}
	}

	sh, stdout, stderr := newTestShell("")
	sh.processInput("cat " + directory + "/'no slash 1' " + directory + "/'one slash \\2' " + directory + "/'two slashes \\3\\'")

	if got, want := stdout.String(), "content1 content2 content3"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestProcessInputCatsBackslashEscapedFileNames(t *testing.T) {
	directory := t.TempDir()
	firstFile := filepath.Join(directory, "_ignored_1")
	secondFile := filepath.Join(directory, "ignore_2")
	thirdFile := filepath.Join(directory, "just_one_\\_3")
	for path, content := range map[string]string{
		firstFile:  "content1 ",
		secondFile: "content2 ",
		thirdFile:  "content3",
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("create file %q: %v", path, err)
		}
	}

	sh, stdout, stderr := newTestShell("")
	sh.processInput("cat " + directory + "/\\_ignored_1 " + directory + "/ignore_\\2 " + directory + "/just_one_\\\\_3")

	if got, want := stdout.String(), "content1 content2 content3"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestProcessInputPrintsWorkingDirectory(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	sh := &shell{
		cwd:    "/test/working-directory",
		stdout: stdout,
		stderr: stderr,
	}

	shouldExit := sh.processInput("pwd")

	if shouldExit {
		t.Fatal("processInput(pwd) requested exit")
	}
	if got, want := stdout.String(), "/test/working-directory\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestProcessInputChangesWorkingDirectory(t *testing.T) {
	originalDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	destination := t.TempDir()
	sh, stdout, stderr := newTestShell("")

	shouldExit := sh.processInput("cd " + destination)

	if shouldExit {
		t.Fatal("processInput(cd) requested exit")
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
	if got, want := sh.cwd, destination; got != want {
		t.Errorf("shell cwd = %q, want %q", got, want)
	}

	stdout.Reset()
	sh.processInput("pwd")
	if got, want := stdout.String(), destination+"\n"; got != want {
		t.Errorf("pwd stdout = %q, want %q", got, want)
	}
}

func TestProcessInputChangesToHomeDirectoryWithoutArguments(t *testing.T) {
	originalDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	homeDirectory := t.TempDir()
	t.Setenv("HOME", homeDirectory)
	sh, stdout, stderr := newTestShell("")

	shouldExit := sh.processInput("cd")

	if shouldExit {
		t.Fatal("processInput(cd) requested exit")
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
	if got, want := sh.cwd, homeDirectory; got != want {
		t.Errorf("shell cwd = %q, want %q", got, want)
	}
}

func TestProcessInputReportsMissingCdDirectory(t *testing.T) {
	sh, stdout, stderr := newTestShell("")
	destination := filepath.Join(t.TempDir(), "missing")

	shouldExit := sh.processInput("cd " + destination)

	if shouldExit {
		t.Fatal("processInput(cd missing directory) requested exit")
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got, want := stderr.String(), "cd: "+destination+": No such file or directory\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestProcessInputRejectsCdWithTooManyArguments(t *testing.T) {
	sh, stdout, stderr := newTestShell("")

	shouldExit := sh.processInput("cd first second")

	if shouldExit {
		t.Fatal("processInput(cd with too many arguments) requested exit")
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got, want := stderr.String(), "cd: too many arguments\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestProcessInputReportsUnknownType(t *testing.T) {
	sh, stdout, stderr := newTestShell("")

	shouldExit := sh.processInput("type missing")

	if shouldExit {
		t.Fatal("processInput(type missing) requested exit")
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got, want := stderr.String(), "missing: not found\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestProcessInputFindsExecutableTypeInPath(t *testing.T) {
	firstDir := t.TempDir()
	secondDir := t.TempDir()
	command := "path-command"
	firstPath := filepath.Join(firstDir, command)

	if err := os.WriteFile(firstPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("create first executable: %v", err)
	}
	if err := os.WriteFile(filepath.Join(secondDir, command), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("create second executable: %v", err)
	}
	t.Setenv("PATH", strings.Join([]string{firstDir, secondDir}, string(os.PathListSeparator)))

	sh, stdout, stderr := newTestShell("")
	shouldExit := sh.processInput("type " + command)

	if shouldExit {
		t.Fatal("processInput(type executable) requested exit")
	}
	if got, want := stdout.String(), command+" is "+firstPath+"\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestProcessInputTypeSkipsNonExecutablePathEntry(t *testing.T) {
	nonExecutableDir := t.TempDir()
	executableDir := t.TempDir()
	command := "later-command"

	if err := os.WriteFile(filepath.Join(nonExecutableDir, command), []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatalf("create non-executable file: %v", err)
	}
	executablePath := filepath.Join(executableDir, command)
	if err := os.WriteFile(executablePath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("create executable file: %v", err)
	}
	t.Setenv("PATH", strings.Join([]string{nonExecutableDir, executableDir}, string(os.PathListSeparator)))

	sh, stdout, stderr := newTestShell("")
	shouldExit := sh.processInput("type " + command)

	if shouldExit {
		t.Fatal("processInput(type executable) requested exit")
	}
	if got, want := stdout.String(), command+" is "+executablePath+"\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestProcessInputReportsTypeNotFoundWhenPathHasNoExecutable(t *testing.T) {
	dir := t.TempDir()
	command := "not-an-executable"
	if err := os.WriteFile(filepath.Join(dir, command), []byte("not executable\n"), 0o644); err != nil {
		t.Fatalf("create non-executable file: %v", err)
	}
	t.Setenv("PATH", dir)

	sh, stdout, stderr := newTestShell("")
	shouldExit := sh.processInput("type " + command)

	if shouldExit {
		t.Fatal("processInput(type missing executable) requested exit")
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got, want := stderr.String(), command+": not found\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestProcessInputRequiresTypeArgument(t *testing.T) {
	sh, stdout, stderr := newTestShell("")

	shouldExit := sh.processInput("type")

	if shouldExit {
		t.Fatal("processInput(type) requested exit")
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got, want := stderr.String(), "Usage: type [command]\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestProcessInputReportsUnknownCommand(t *testing.T) {
	sh, stdout, stderr := newTestShell("")

	shouldExit := sh.processInput("missing")

	if shouldExit {
		t.Fatal("processInput(missing) requested exit")
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got, want := stderr.String(), "missing: command not found\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestProcessInputExitRequestsExit(t *testing.T) {
	sh, stdout, stderr := newTestShell("")

	shouldExit := sh.processInput("exit")

	if !shouldExit {
		t.Fatal("processInput(exit) did not request exit")
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestProcessInputIgnoresWhitespace(t *testing.T) {
	sh, stdout, stderr := newTestShell("")

	shouldExit := sh.processInput("   \t  ")

	if shouldExit {
		t.Fatal("processInput(whitespace) requested exit")
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestRunExecutesCommandsUntilExit(t *testing.T) {
	input := "echo first\ntype echo\necho last\nexit\necho ignored\n"
	sh, stdout, stderr := newTestShell(input)

	err := sh.run()

	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	wantStdout := "$ first\n$ echo is a shell builtin\n$ last\n$ "
	if got := stdout.String(); got != wantStdout {
		t.Errorf("stdout = %q, want %q", got, wantStdout)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestRunProcessesFinalLineAtEOF(t *testing.T) {
	sh, stdout, stderr := newTestShell("echo final")

	err := sh.run()

	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if got, want := stdout.String(), "$ final\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestRunReturnsReadError(t *testing.T) {
	wantErr := errors.New("read failed")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	sh, err := newShell(iotest.ErrReader(wantErr), stdout, stderr)
	if err != nil {
		t.Fatalf("newShell() error = %v", err)
	}

	err = sh.run()

	if !errors.Is(err, wantErr) {
		t.Fatalf("run() error = %v, want error wrapping %v", err, wantErr)
	}
	if got, want := stdout.String(), "$ "; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}
