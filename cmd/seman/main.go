package main

import (
	"context"
	"fmt"
	"os"
	"seman/internal"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"golang.org/x/term"
)

func main() {
	ctx := context.Background()

	backend := flag.StringP("backend", "b", "systemd", "Service manager backend")
	debug := flag.BoolP("debug", "d", false, "Prints useful debug info")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: seman <subcommand> [FLAGS]",
			"\n\nCommands:\n", " service")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	serviceCmd := flag.NewFlagSet("service", flag.ExitOnError)
	serviceCmd.AddFlagSet(flag.CommandLine)
	serviceCmd.Usage = func() {
		fmt.Fprint(os.Stderr, "")
	}
	flag.CommandLine.AddFlagSet(serviceCmd)

	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	serviceManager := func() internal.Backend {
		switch *backend {
		case "systemd":
			return internal.Systemd{}
		}
		os.Exit(1)
		return nil
	}()

	log.SetLevel(log.ErrorLevel)
	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	switch os.Args[1] {
	case "service":
		if len(os.Args) < 3 {
			os.Exit(1)
		}
		switch os.Args[2] {
		case "up":
			fmt.Print("Vault password: ")
			passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				fmt.Fprintln(os.Stderr, "\nError reading password:", err)
				return
			}

			fmt.Println()
			err = serviceManager.UnlockVault(ctx, string(passwordBytes))
			if err != nil {
				log.Fatal(err)
			}
		case "down":
			err := serviceManager.LockVault(ctx)
			if err != nil {
				log.Fatal(err)
			}
		default:
			os.Exit(1)
		}
	default:
		flag.Usage()
		os.Exit(1)
	}
}
