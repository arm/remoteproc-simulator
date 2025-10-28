package simulator

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileSystemManager struct {
	instanceDir         string
	firmwareDirPathFile string
	firmwareDir         string
	createdDirs         []string
}

func NewFileSystemManager(rootDir string, index uint) *FileSystemManager {
	instanceName := fmt.Sprintf("remoteproc%d", index)
	return &FileSystemManager{
		instanceDir:         filepath.Join(rootDir, "sys", "class", "remoteproc", instanceName),
		firmwareDirPathFile: filepath.Join(rootDir, "sys", "module", "firmware_class", "parameters", "path"),
		firmwareDir:         filepath.Join(rootDir, "firmware"),
		createdDirs:         []string{},
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

	createdFirmwareDirPathFileFolder, err := mkfile(fs.firmwareDirPathFile, 0755)
	if err != nil {
		fs.Cleanup()
		return fmt.Errorf("failed to create firmware path directory: %w", err)
	}
	fs.createdDirs = append(fs.createdDirs, createdFirmwareDirPathFileFolder)

	if err := os.WriteFile(fs.firmwareDirPathFile, []byte(fs.firmwareDir), 0644); err != nil {
		fs.Cleanup()
		return fmt.Errorf("failed to write firmware path file: %w", err)
	}

	createdFirmwareDir, err := mkdirAll(fs.firmwareDir, 0755)
	if err != nil {
		fs.Cleanup()
		return fmt.Errorf("failed to create firmware directory: %w", err)
	}
	fs.createdDirs = append(fs.createdDirs, createdFirmwareDir)

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
	path := filepath.Join(fs.firmwareDir, firmwareName)
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (fs *FileSystemManager) InstanceDir() string {
	return fs.instanceDir
}

func (fs *FileSystemManager) FirmwareDir() string {
	return fs.firmwareDir
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

func mkfile(path string, perm os.FileMode) (string, error) {
	folder, err := mkdirAll(filepath.Dir(path), perm)
	if err != nil {
		return "", err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, perm)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return folder, nil
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
