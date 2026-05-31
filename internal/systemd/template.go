package systemd

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"seman/internal"

	"github.com/coreos/go-systemd/v22/unit"
	log "github.com/sirupsen/logrus"
)

type templateOptions struct {
	internal.Options
	MountUnitName string
}

func generateMountFile(options internal.Options) error {
	tmpl, err := template.ParseFS(templateFS, "templates/seman.mount.tmpl")
	if err != nil {
		log.Fatalf("Failed to parse embedded template: %v", err)
	}

	mountUnitName := fmt.Sprintf("%s.mount", unit.UnitNamePathEscape(options.MountPoint))
	targetFile := filepath.Join(internal.SystemdDataDir(), options.Name, mountUnitName)
	file, err := os.Create(targetFile)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, options)
	if err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	err = os.Symlink(targetFile, filepath.Join(ServicePath, filepath.Base(targetFile)))
	if err != nil {
		return fmt.Errorf("Error creating symlink: %v\n", err)
	}

	return nil
}

func generateServiceFile(options internal.Options) error {
	tmpl, err := template.ParseFS(templateFS, "templates/seman.service.tmpl")
	if err != nil {
		log.Fatalf("Failed to parse embedded template: %v", err)
	}

	targetFile := filepath.Join(internal.SystemdDataDir(), options.Name, fmt.Sprintf("seman-vault-%s.service", options.Name))
	file, err := os.Create(targetFile)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, options)
	if err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	err = os.Symlink(targetFile, filepath.Join(ServicePath, filepath.Base(targetFile)))
	if err != nil {
		return fmt.Errorf("Error creating symlink: %v\n", err)
	}

	return nil
}

func generateStowFile(options internal.Options) error {
	tmpl, err := template.ParseFS(templateFS, "templates/seman-stow.service.tmpl")
	if err != nil {
		log.Fatalf("Failed to parse embedded template: %v", err)
	}

	targetFile := filepath.Join(internal.SystemdDataDir(), options.Name, fmt.Sprintf("seman-vault-%s-stow.service", options.Name))
	file, err := os.Create(targetFile)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	templateOpts := templateOptions{Options: options, MountUnitName: fmt.Sprintf("%s.mount", unit.UnitNamePathEscape(options.MountPoint))}
	err = tmpl.Execute(file, templateOpts)
	if err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	err = os.Symlink(targetFile, filepath.Join(ServicePath, filepath.Base(targetFile)))
	if err != nil {
		return fmt.Errorf("Error creating symlink: %v\n", err)
	}

	return nil
}
