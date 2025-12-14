package cli

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Cli struct {
	log        *slog.Logger
	cliHandler CommandHandler
}

type CommandHandler interface {
	ComputeHandler(ctx context.Context, raw string) (string, error)
}

func NewCli(log *slog.Logger, handler CommandHandler) *Cli {
	return &Cli{
		log:        log,
		cliHandler: handler,
	}
}

func (cli *Cli) Start(parent context.Context) (context.Context, <-chan error) {
	const op = "cli.Cli.Start"

	ctx, cancel := context.WithCancel(parent)
	errCh := make(chan error, 1)

	go func() {
		<-ctx.Done()
		_ = os.Stdin.Close()
	}()

	go func() {
		defer close(errCh)
		defer cancel()

		scanner := bufio.NewScanner(os.Stdin)
		fmt.Fprintln(os.Stdout, "\nInput command (exit for exit)")
		fmt.Fprint(os.Stdout, "> ")

		for {

			if !scanner.Scan() {
				err := scanner.Err()
				if err != nil {
					cli.log.Info("Cli input closed", slog.String("operation", op))
					errCh <- err
				}
				return
			}

			line := strings.TrimSpace(scanner.Text())

			if line == "" {
				fmt.Fprint(os.Stdout, "> ")
				continue
			}

			if strings.EqualFold(line, "exit") {
				cli.log.Info("Cli exiting", slog.String("operation", op))
				return
			}

			result, err := cli.cliHandler.ComputeHandler(ctx, line)
			if err != nil {
				fmt.Fprintln(os.Stdout, err)
				fmt.Fprint(os.Stdout, "> ")
				continue
			}

			fmt.Fprintln(os.Stdout, result)

			fmt.Fprint(os.Stdout, "> ")
		}
	}()

	return ctx, errCh
}
