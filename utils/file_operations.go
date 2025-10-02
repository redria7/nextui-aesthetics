package utils

import (
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	// shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"fmt"
	"os"
	"io"
	"path/filepath"
)

func GetFileList(dirPath string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}
	return entries, nil
}

func DoesFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func EnsureDirectoryExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, defaultDirPerm)
	}
	return nil
}

func MoveFile(sourcePath, destinationPath string) error {
	logger := common.GetLoggerInstance()

	if err := EnsureDirectoryExists(filepath.Dir(destinationPath)); err != nil {
		logger.Error("Failed to create destination directory", zap.Error(err))
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := os.Rename(sourcePath, destinationPath); err != nil {
		logger.Error("Failed to move file", zap.String("from", sourcePath), zap.String("to", destinationPath), zap.Error(err))
		return fmt.Errorf("failed to move file from %s to %s: %w", sourcePath, destinationPath, err)
	}

	return nil
}

func getDefaultWallpaperPath() string {
	if DoesFileExists(common.SDCardRoot + "/bg.png") {
		return common.SDCardRoot + "/bg.png"
	}
	return ""
}

func GetWallpaperPath(itemPath string, parentPath string) string {
	actualWallpaperPath := genWallpaperPath(itemPath)
	if DoesFileExists(actualWallpaperPath) {
		return actualWallpaperPath
	}
	return GetListWallpaperPath(parentPath)
}

func GetListWallpaperPath(itemPath string) string {
	if itemPath == GetRomDirectory() {
		return getDefaultWallpaperPath()
	}
	actualListWallpaperPath := genListWallpaperPath(itemPath)
	if DoesFileExists(actualListWallpaperPath) {
		return actualListWallpaperPath
	}
	return getDefaultWallpaperPath()
}

func GetIconPath(parentPath string, itemName string) string {
	actualIconPath := genIconPath(parentPath, itemName)
	if DoesFileExists(actualIconPath) {
		return actualIconPath
	}
	return ""
}

func genWallpaperPath(itemPath string) string {
	return itemPath + "/.media/bg.png"
}

func genListWallpaperPath(itemPath string) string {
	return itemPath + "/.media/bglist.png"
}

func genIconPath(parentPath string, itemName string) string {
	return parentPath + "/.media/" + itemName + ".png"
}

// copyFile copies a file from src to dst.
func copyFile(sourcePath, destinationPath string) error {
	EnsureDirectoryExists(filepath.Dir(destinationPath))

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	sourceFileInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}
	err = os.Chmod(destinationPath, sourceFileInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to set destination file permissions: %w", err)
	}

	return nil
}

// func DeleteRom(game shared.Item, romDirectory shared.RomDirectory) {
// 	romPath := filepath.Join(romDirectory.Path, game.Filename)
// 	if common.DeleteFile(romPath) {
// 		DeleteArt(game.Filename, romDirectory)
// 	}
// }

// func Nuke(game shared.Item, romDirectory shared.RomDirectory) {
// 	ClearGameTracker(game.Filename, romDirectory)
// 	DeleteRom(game, romDirectory)
// }
