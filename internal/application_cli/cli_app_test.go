package applicationcli

import (
	"bytes"
	"flag"
	"io"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateLoggerLevels(t *testing.T) {
	t.Parallel()

	for _, level := range []string{LoggerLevelInfo, LoggerLevelDev, LoggerLevelProd, "unknown"} {
		log, err := CreateLogger(level)
		require.NoError(t, err)
		require.NotNil(t, log)
	}
}

func TestAppCliRun_ExitAndCommand(t *testing.T) {
	// глобальные флаги/stdio — без параллельности

	// стартуем простой TCP сервер-эхо
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buf := make([]byte, 32)
		n, _ := conn.Read(buf)
		_, _ = conn.Write([]byte("RESP:" + string(bytes.TrimSpace(buf[:n]))))
	}()

	// подготавливаем stdin/stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	// stdin через pipe: пишем команды и закрываем writer
	stdinR, stdinW, _ := os.Pipe()
	_, _ = stdinW.Write([]byte("PING\nexit\n"))
	_ = stdinW.Close()

	stdoutR, stdoutW, _ := os.Pipe()

	os.Stdin = stdinR
	os.Stdout = stdoutW

	t.Cleanup(func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	})

	// отдельный FlagSet для теста
	oldArgs := os.Args
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{
		"cli",
		"-address", ln.Addr().String(),
		"-idle", "100ms",
		"-mes_size", "16",
		"-debug=false",
	}
	t.Cleanup(func() { os.Args = oldArgs })

	app := NewAppCli()
	app.Run()

	// закрываем stdout, читаем вывод
	_ = stdoutW.Close()
	outBytes, _ := io.ReadAll(stdoutR)

	<-serverDone

	out := string(outBytes)
	require.Contains(t, out, "RESP:PING")
	require.Contains(t, out, "Input command")
}
