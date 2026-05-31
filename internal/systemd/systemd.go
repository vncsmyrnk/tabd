package systemd

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"seman/internal"

	"github.com/coreos/go-systemd/v22/dbus"
	log "github.com/sirupsen/logrus"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

type Systemd struct{}

var _ internal.Backend = Systemd{}

var (
	CryptsetupFIFOAuth = "/dev/shm/vault-fifo"
	ServicePath        = "/etc/systemd/system"
)

func (s Systemd) UnlockVault(ctx context.Context, vault, password string) error {
	return unlockVaultWithFIFO(ctx, vault, password)
}

func unlockVaultWithFIFO(ctx context.Context, vault, password string) error {
	fifoPath := CryptsetupFIFOAuth
	err := os.RemoveAll(fifoPath)
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

	mountServiceName, err := vaultMountServiceFileBase(vault)
	if err != nil {
		return err
	}

	resChan := make(chan string)
	jobID, err := conn.StartUnitContext(ctx, mountServiceName, "replace", resChan)
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

func vaultMountServiceFileBase(vault string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(internal.SystemdDataDir(), vault, "*.mount"))
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("failed to file mount service file")
	}

	return filepath.Base(matches[0]), nil
}

func (s Systemd) LockVault(ctx context.Context, vault string) error {
	conn, err := dbus.NewSystemConnectionContext(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to systemd D-Bus: %v", err)
	}
	defer conn.Close()

	mountServiceName, err := vaultMountServiceFileBase(vault)
	if err != nil {
		return err
	}

	resChan := make(chan string)
	jobID, err := conn.StopUnitContext(ctx, mountServiceName, "replace", resChan)
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

func (s Systemd) GenerateService(_ context.Context, options internal.Options) error {
	servicePath := filepath.Join(internal.SystemdDataDir(), options.Name)
	err := os.MkdirAll(servicePath, 0755)
	if err != nil {
		return err
	}

	if options.MountPoint == "" {
		options.MountPoint = filepath.Join(servicePath, "mnt")
	}

	err = generateServiceFile(options)
	if err != nil {
		return err
	}

	err = generateMountFile(options)
	if err != nil {
		return err
	}

	if options.StowPath != "" {
		err = generateStowFile(options)
		if err != nil {
			return err
		}
	}

	return nil
}
