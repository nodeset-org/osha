package filesystem

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FilesystemManager manages folders used by OSHA tests
type FilesystemManager struct {
	name        string
	logger      *slog.Logger
	testDir     string
	snapshotDir string
}

// Creates a new FilesystemManager instance
func NewFilesystemManager(logger *slog.Logger) (*FilesystemManager, error) {
	// Create a temp folder
	testDir, err := os.MkdirTemp("", "osha-*")
	if err != nil {
		return nil, fmt.Errorf("error creating test dir: %v", err)
	}
	logger.Info("Created test dir", "dir", testDir)

	// Create a snapshot folder
	snapshotDir, err := os.MkdirTemp("", "osha-snapshots-*")
	if err != nil {
		return nil, fmt.Errorf("error creating snapshot dir: %v", err)
	}
	logger.Info("Created snapshot dir", "dir", snapshotDir)

	return &FilesystemManager{
		logger:      logger,
		testDir:     testDir,
		snapshotDir: snapshotDir,
	}, nil
}

func (m *FilesystemManager) GetName() string {
	return m.name
}

func (m *FilesystemManager) GetRequirements() {
}

// Get the test dir
func (m *FilesystemManager) GetTestDir() string {
	return m.testDir
}

// Delete the test dir and snapshot dir
func (m *FilesystemManager) Close() error {
	// Remove the test dir
	if m.testDir != "" {
		err := os.RemoveAll(m.testDir)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("error removing test dir [%s]: %v", m.testDir, err)
		}
		m.testDir = ""
	}

	// Remove the snapshot dir
	if m.snapshotDir != "" {
		err := os.RemoveAll(m.snapshotDir)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("error removing test dir [%s]: %v", m.snapshotDir, err)
		}
		m.snapshotDir = ""
	}
	return nil
}

// Take a snapshot of the current test dir
func (m *FilesystemManager) TakeSnapshot() error {
	timestamp := time.Now().Format("20060102_150405")
	snapshotName := fmt.Sprintf("%s_%s", m.name, timestamp)

	// Error out if the snapshot already exists
	snapshotPath := filepath.Join(m.snapshotDir, snapshotName)
	_, err := os.ReadFile(snapshotPath)
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("snapshot with name [%s] already exists", snapshotName)
	}

	// Make the snapshot folder
	err = os.Mkdir(snapshotPath, 0755)
	if err != nil {
		return fmt.Errorf("error creating snapshot dir [%s]: %v", snapshotPath, err)
	}

	// Copy the test dir to the snapshot dir
	err = copyDirectory(m.testDir, snapshotPath)
	if err != nil {
		return fmt.Errorf("error copying test dir to snapshot dir: %v", err)
	}

	m.logger.Info("Took snapshot", "name", snapshotName, "path", snapshotPath)
	return nil
}

// Revert to a snapshot of the test dir
func (m *FilesystemManager) RevertToSnapshot(name string) error {
	// Error out if the snapshot already exists
	snapshotPath := filepath.Join(m.snapshotDir, name)
	_, err := os.ReadFile(snapshotPath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("snapshot with name [%s] doesn't exist", name)
	}

	// Delete everything in the test dir
	err = os.RemoveAll(m.testDir)
	if err != nil {
		return fmt.Errorf("error removing test dir [%s]: %v", m.testDir, err)
	}

	// Recreate the test dir
	err = os.Mkdir(m.testDir, 0755)
	if err != nil {
		return fmt.Errorf("error recreating test dir [%s]: %v", m.testDir, err)
	}

	// Copy the snapshot dir to the test dir
	err = copyDirectory(snapshotPath, m.testDir)
	if err != nil {
		return fmt.Errorf("error copying snapshot dir to test dir: %v", err)
	}

	m.logger.Info("Reverted to snapshot", "name", name, "path", snapshotPath)
	return nil
}

// Recursively copies an entire directory for snapshotting. Irregular files like symlinks aren't supported.
// source should be a full path.
func copyDirectory(source string, target string) error {
	// Derived from Gregory Vincic's code: https://stackoverflow.com/a/72246196
	walker := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		sourceRelative := strings.TrimPrefix(path, source)
		targetPath := filepath.Join(target, sourceRelative)

		// Directories will be traversed via Walk() later so just make the dir here
		if info.IsDir() {
			err = os.MkdirAll(targetPath, info.Mode())
			return err
		}

		// Irregular files aren't supported yet
		if !info.Mode().IsRegular() {
			return fmt.Errorf("file [%s] is irregular, copying is not supported", path)
		}

		// Open the source
		sourceFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		// Create the target
		targetFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer targetFile.Close()

		// Set the permissions
		err = targetFile.Chmod(info.Mode())
		if err != nil {
			return fmt.Errorf("error setting permissions on file [%s]: %w", targetPath, err)
		}

		// Copy the file
		_, err = io.Copy(targetFile, sourceFile)
		return err
	}

	return filepath.Walk(source, walker)
}
