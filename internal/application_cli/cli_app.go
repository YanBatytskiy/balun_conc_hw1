package applicationcli

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"flag"
	"spyder/internal/lib/logger/slogdiscard"
	"spyder/internal/lib/logger/slogpretty"
	"spyder/internal/network"
	"time"
)

type AppCli struct{}

func NewAppCli() *AppCli {
	return &AppCli{}
}

func (appCli *AppCli) Run() {
	const op = "AppCli.Run"

	address := flag.String("address", "127.0.0.1:3323", "Address of server")
	idleTimeout := flag.Duration("idle", time.Minute*5, "Idle timeout of connections")
	maxMessageSize := flag.Int("mes_size", 4096, "Max message size")
	debug := flag.Bool("debug", true, "Debug enviroment")

	flag.Parse()

	env := ""
	if *debug {
		env = "dev"
	} else {
		env = "info"
	}
	log, err := CreateLogger(env)
	if err != nil {
		_ = log
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	var options []network.TCPClientOption

	options = append(options, network.WithClientTCPAddress(*address))
	options = append(options, network.WithClientTCPMaxMessageSize(*maxMessageSize))
	options = append(options, network.WithClientTCPIdleTimeout(*idleTimeout))

	reader := bufio.NewReader(os.Stdin)
	// log.Debug("connecting to server", slog.String("address", *address))
	tcpClient, err := network.NewTCPClient(log, options...)
	if err != nil {
		log.Info("cannot connect to server", slog.String("address", *address), slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("connected to server", slog.String("address", *address))

	for {
		fmt.Fprintln(os.Stdout, "\nInput command (exit for exit)")
		fmt.Fprint(os.Stdout, "> ")

		request, err := reader.ReadString('\n')
		if err != nil {
			log.Info("input closed, exiting")
			return
		}

		request = strings.TrimSpace(request)

		if request == "" {
			fmt.Fprint(os.Stdout, "> ")
			continue
		}

		if strings.EqualFold(request, "exit") {
			log.Info("Cli exit and finished", slog.String("operation", op))
			return
		}

		response, err := tcpClient.SendAndReceive([]byte(request))

		if err != nil {
			log.Info("failed to send command", slog.String("error", err.Error()))
			fmt.Fprint(os.Stdout, "> ")
			continue
		}

		fmt.Fprintln(os.Stdout, string(response))
		fmt.Fprint(os.Stdout, "> ")
	}

}

const (
	LoggerLevelInfo = "info"
	LoggerLevelDev  = "dev"
	LoggerLevelProd = "prod"
)

func CreateLogger(env string) (*slog.Logger, error) {

	var log *slog.Logger

	switch env {
	case LoggerLevelInfo:
		opts := slogpretty.PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{Level: slog.LevelInfo},
		}
		handler := opts.NewPrettyHandler(os.Stdout)

		log = slog.New(handler)

	case LoggerLevelDev:
		opts := slogpretty.PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{Level: slog.LevelDebug},
		}
		handler := opts.NewPrettyHandler(os.Stdout)

		log = slog.New(handler)
	case LoggerLevelProd:
		log = slogdiscard.NewDiscardLogger()
	default:
		opts := slogpretty.PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{Level: slog.LevelInfo},
		}
		handler := opts.NewPrettyHandler(os.Stdout)

		log = slog.New(handler)

	}
	log.Info("starting service", slog.String("logger level", env))
	return log, nil
}
