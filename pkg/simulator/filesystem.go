package simulator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileSystemManager struct {
	instanceDir            string
	firmwareSearchPathFile string
	defaultFirmwareDir     string
	createdDirs            []string
}

func NewFileSystemManager(rootDir string, index uint) *FileSystemManager {
	instanceName := fmt.Sprintf("remoteproc%d", index)
	return &FileSystemManager{
		instanceDir:            filepath.Join(rootDir, "sys", "class", "remoteproc", instanceName),
		firmwareSearchPathFile: filepath.Join(rootDir, "sys", "module", "firmware_class", "parameters", "path"),
		defaultFirmwareDir:     filepath.Join(rootDir, "lib", "firmware"),
		createdDirs:            []string{},
	}
}

func (fs *FileSystemManager) BootstrapDirectories() error {
	createdInstancePath, err := mkdirAll(fs.instanceDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create instance directory: %w", err)
	}
	if createdInstancePath != "" {
		fs.createdDirs = append(fs.createdDirs, createdInstancePath)
	}

	createdParametersDir, err := mkdirAll(filepath.Dir(fs.firmwareSearchPathFile), 0755)
	if err != nil {
		fs.Cleanup()
		return fmt.Errorf("failed to create parameters directory: %w", err)
	}
	if createdParametersDir != "" {
		fs.createdDirs = append(fs.createdDirs, createdParametersDir)
	}
	if err := os.WriteFile(fs.firmwareSearchPathFile, []byte(""), 0644); err != nil {
		fs.Cleanup()
		return fmt.Errorf("failed to create empty fimware search path file; %w", err)
	}

	createdDefaultFirmwareDir, err := mkdirAll(fs.defaultFirmwareDir, 0755)
	if err != nil {
		fs.Cleanup()
		return fmt.Errorf("failed to create firmware directory: %w", err)
	}
	if createdDefaultFirmwareDir != "" {
		fs.createdDirs = append(fs.createdDirs, createdDefaultFirmwareDir)
	}

	return nil
}

func (fs *FileSystemManager) WriteInstanceFile(filename, content string) error {
	path := filepath.Join(fs.instanceDir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write %s: %w", filename, err)
	}
	return nil
}

func (fs *FileSystemManager) FirmwareExists(firmwareName string) bool {
	return fs.firmwareExistsInCustomPath(firmwareName) || fs.firmwareExistsInDefaultPath(firmwareName)
}

func (fs *FileSystemManager) firmwareExistsInCustomPath(firmwareName string) bool {
	customFirmwarePath := fs.customFirmwarePath()
	if customFirmwarePath == "" {
		return false
	}
	path := filepath.Join(customFirmwarePath, firmwareName)
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (fs *FileSystemManager) customFirmwarePath() string {
	customFirmwareLoadPath, err := os.ReadFile(fs.firmwareSearchPathFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(customFirmwareLoadPath))
}

func (fs *FileSystemManager) firmwareExistsInDefaultPath(firmwareName string) bool {
	path := filepath.Join(fs.defaultFirmwareDir, firmwareName)
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (fs *FileSystemManager) InstanceDir() string {
	return fs.instanceDir
}

func (fs *FileSystemManager) FirmwareDir() string {
	return fs.defaultFirmwareDir
}

func (fs *FileSystemManager) Cleanup() error {
	for _, dir := range fs.createdDirs {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("failed to remove directory %s: %w", dir, err)
		}
	}
	fs.createdDirs = []string{}
	return nil
}

func mkdirAll(path string, perm os.FileMode) (string, error) {
	current := path
	var topmostMissing string

	for current != "/" && current != "." {
		_, err := os.Stat(current)
		dirExists := err == nil
		if dirExists {
			break
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		topmostMissing = current
		current = filepath.Dir(current)
	}

	if topmostMissing == "" {
		return "", nil
	}

	if err := os.MkdirAll(path, perm); err != nil {
		return "", err
	}

	return topmostMissing, nil
}
