package internal

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/coreos/go-systemd/v22/dbus"
	log "github.com/sirupsen/logrus"
)

type Systemd struct{}

var _ Backend = Systemd{}

var SystemdCryptsetupFIFOAuth = "/dev/shm/vault-fifo"

func (s Systemd) UnlockVault(ctx context.Context, password string) error {
	return unlockVaultWithFIFO(ctx, password)
}

func unlockVaultWithFIFO(ctx context.Context, password string) error {
	fifoPath := SystemdCryptsetupFIFOAuth

	err := os.Remove(fifoPath)
	if err != nil {
		return err
	}

	err = syscall.Mkfifo(fifoPath, 0600)
	if err != nil {
		return fmt.Errorf("failed to create FIFO: %w", err)
	}

	defer func() {
		err := os.Remove(fifoPath)
		if err != nil {
			log.Errorf("failed to remove FIFO: %s", err)
		}
	}()

	errChan := make(chan error, 1)

	go func() {
		file, err := os.OpenFile(fifoPath, os.O_WRONLY, os.ModeNamedPipe)
		if err != nil {
			errChan <- fmt.Errorf("failed to open FIFO for writing: %w", err)
			return
		}
		defer func() {
			err := file.Close()
			if err != nil {
				log.Errorf("failed to close FIFO: %s", err)
			}
		}()

		_, err = file.WriteString(password)
		if err != nil {
			errChan <- fmt.Errorf("failed to write to FIFO: %w", err)
			return
		}
		errChan <- nil
	}()

	conn, err := dbus.NewSystemConnectionContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to system bus: %w", err)
	}
	defer conn.Close()

	resChan := make(chan string)
	jobID, err := conn.StartUnitContext(ctx, "seman-vault.mount", "replace", resChan)
	if err != nil {
		return fmt.Errorf("failed to queue start job: %w", err)
	}

	log.Debugf("Triggered systemd job #%d. Awaiting FIFO consumption...\n", jobID)

	if writerErr := <-errChan; writerErr != nil {
		return writerErr
	}

	select {
	case result := <-resChan:
		if result != "done" {
			return fmt.Errorf("systemd job aborted prematurely with status: %s", result)
		}
	case writerErr := <-errChan:
		if writerErr != nil {
			return writerErr
		}

		result := <-resChan
		if result != "done" {
			return fmt.Errorf("decryption failed or systemd job exited with status: %s", result)
		}
	}

	log.Debug("Success: Vault unlocked via memory-only stream.")
	return nil
}

func (s Systemd) LockVault(ctx context.Context) error {
	conn, err := dbus.NewSystemConnectionContext(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to systemd D-Bus: %v", err)
	}
	defer conn.Close()

	resChan := make(chan string)

	jobID, err := conn.StopUnitContext(ctx, "seman-vault.mount", "replace", resChan)
	if err != nil {
		log.Fatalf("failed to request unit stop: %v", err)
	}

	log.Debugf("successfully queued stop job #%d. Waiting for teardown...\n", jobID)

	result := <-resChan
	if result == "done" {
		log.Debugf("success: Vault is unstowed and locked.")
	} else {
		return fmt.Errorf("decryption failed or systemd job exited with status: %s", result)
	}
	return nil
}
