package cli_test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"lesson1/internal/cli"
	cli_test "lesson1/internal/cli/mocks"
	"lesson1/internal/lib/logger/slogdiscard"
)

//nolint:paralleltest // uses global stdio redirection
func TestCliStart(t *testing.T) {

	tests := []struct {
		name           string
		input          string
		setupMock      func(h *cli_test.MockCommandHandler)
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:  "handler ok",
			input: "SET key value\nexit\n",
			setupMock: func(h *cli_test.MockCommandHandler) {
				h.EXPECT().ComputeHandler(mock.Anything, "SET key value").Return("OK", nil)
			},
			wantContains:   []string{"Input command (exit for exit)", "> ", "OK"},
			wantNotContain: []string{"set failed"},
		},
		{
			name:  "handler error",
			input: "SET key value\nexit\n",
			setupMock: func(h *cli_test.MockCommandHandler) {
				h.EXPECT().ComputeHandler(mock.Anything, "SET key value").Return("", assert.AnError)
			},
			wantContains:   []string{assert.AnError.Error()},
			wantNotContain: []string{"OK"},
		},
		{
			name:           "empty line skip",
			input:          "\nexit\n",
			wantContains:   []string{"Input command (exit for exit)"},
			wantNotContain: []string{"OK", assert.AnError.Error()},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := cli_test.NewMockCommandHandler(t)
			if tc.setupMock != nil {
				tc.setupMock(handler)
			}

			output, errs := runCli(t, handler, tc.input)

			require.Empty(t, errs)
			for _, want := range tc.wantContains {
				assert.Contains(t, output, want)
			}
			for _, notWant := range tc.wantNotContain {
				assert.NotContains(t, output, notWant)
			}
		})
	}
}

func runCli(t *testing.T, handler *cli_test.MockCommandHandler, input string) (string, []error) {
	t.Helper()

	stdinR, stdinW, err := os.Pipe()
	require.NoError(t, err)
	stdoutR, stdoutW, err := os.Pipe()
	require.NoError(t, err)

	origStdin, origStdout := os.Stdin, os.Stdout
	os.Stdin, os.Stdout = stdinR, stdoutW

	t.Cleanup(func() {
		_ = stdinR.Close()
		_ = stdinW.Close()
		_ = stdoutR.Close()
		_ = stdoutW.Close()
		os.Stdin, os.Stdout = origStdin, origStdout
	})

	logger := slogdiscard.NewDiscardLogger()
	c := cli.NewCli(logger, handler)

	ctx, errCh := c.Start(context.Background())

	_, _ = io.WriteString(stdinW, input)
	_ = stdinW.Close()

	var errs []error
	for err := range errCh {
		if err != nil {
			errs = append(errs, err)
		}
	}
	<-ctx.Done()

	_ = stdoutW.Close()
	outBytes, _ := io.ReadAll(stdoutR)

	return string(outBytes), errs
}
