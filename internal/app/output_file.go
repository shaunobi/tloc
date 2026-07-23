package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type outputPlan struct {
	path     string
	force    bool
	existing *os.File
}

func (plan *outputPlan) close() error {
	if plan.existing == nil {
		return nil
	}
	err := plan.existing.Close()
	plan.existing = nil
	return err
}

func preflightOutput(name string, force bool) (outputPlan, error) {
	if name == "" {
		return outputPlan{}, nil
	}
	path, err := filepath.Abs(name)
	if err != nil {
		return outputPlan{}, fmt.Errorf("resolve output %q: %w", name, err)
	}
	path = filepath.Clean(path)

	info, err := os.Lstat(path)
	switch {
	case err == nil:
		if info.IsDir() {
			return outputPlan{}, fmt.Errorf("output %q is a directory", name)
		}
		if info.Mode()&os.ModeSymlink == 0 && !info.Mode().IsRegular() {
			return outputPlan{}, fmt.Errorf("output %q is not a regular file", name)
		}
		if !force {
			return outputPlan{}, fmt.Errorf("output %q already exists; pass --force to overwrite it", name)
		}
		file, openErr := os.OpenFile(path, os.O_WRONLY, 0)
		if openErr != nil {
			return outputPlan{}, fmt.Errorf("check output %q: %w", name, openErr)
		}
		openedInfo, statErr := file.Stat()
		if statErr != nil {
			_ = file.Close()
			return outputPlan{}, fmt.Errorf("check output %q: %w", name, statErr)
		}
		if !openedInfo.Mode().IsRegular() {
			_ = file.Close()
			return outputPlan{}, fmt.Errorf("output %q is not a regular file", name)
		}
		return outputPlan{path: path, force: force, existing: file}, nil
	case !errors.Is(err, os.ErrNotExist):
		return outputPlan{}, fmt.Errorf("inspect output %q: %w", name, err)
	}

	probe, err := os.CreateTemp(filepath.Dir(path), ".tloc-output-check-*")
	if err != nil {
		return outputPlan{}, fmt.Errorf("check output directory for %q: %w", name, err)
	}
	probeName := probe.Name()
	closeErr := probe.Close()
	removeErr := os.Remove(probeName)
	if closeErr != nil {
		return outputPlan{}, fmt.Errorf("check output directory for %q: %w", name, closeErr)
	}
	if removeErr != nil {
		return outputPlan{}, fmt.Errorf("clean output check for %q: %w", name, removeErr)
	}
	return outputPlan{path: path, force: force}, nil
}

func writeOutput(plan *outputPlan, content []byte) error {
	if plan.existing != nil {
		if err := ensureOutputIdentity(plan); err != nil {
			return err
		}
		if err := plan.existing.Truncate(0); err != nil {
			return err
		}
		if _, err := plan.existing.Seek(0, io.SeekStart); err != nil {
			return err
		}
		if _, err := plan.existing.Write(content); err != nil {
			return err
		}
		if err := ensureOutputIdentity(plan); err != nil {
			return err
		}
		return plan.close()
	}

	file, err := os.OpenFile(plan.path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			if plan.force {
				return fmt.Errorf("output appeared after preflight; refusing to overwrite it until it is preflighted as an existing destination")
			}
			return fmt.Errorf("output appeared after preflight; refusing to overwrite it (pass --force to overwrite)")
		}
		return err
	}
	if _, err := file.Write(content); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func ensureOutputIdentity(plan *outputPlan) error {
	openedInfo, err := plan.existing.Stat()
	if err != nil {
		return fmt.Errorf("inspect preflighted output: %w", err)
	}
	currentInfo, err := os.Stat(plan.path)
	if err != nil {
		return fmt.Errorf("output changed after preflight: %w", err)
	}
	if !os.SameFile(openedInfo, currentInfo) {
		return errors.New("output changed after preflight; refusing to overwrite the replacement")
	}
	return nil
}
