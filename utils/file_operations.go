package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
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

func CheckWallpaperPath(itemPath string) bool {
	actualWallpaperPath := genWallpaperPath(itemPath)
	return DoesFileExists(actualWallpaperPath)
}

func GetTrueWallpaperPath(itemPath string) string {
	return genWallpaperPath(itemPath)
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

func CheckListWallpaperPath(itemPath string) bool {
	if itemPath == GetRomDirectory() {
		return DoesFileExists(common.SDCardRoot + "/bg.png")
	}
	actualListWallpaperPath := genListWallpaperPath(itemPath)
	return DoesFileExists(actualListWallpaperPath)
}

func GetTrueListWallpaperPath(itemPath string) string {
	if itemPath == GetRomDirectory() {
		return common.SDCardRoot + "/bg.png"
	}
	return genListWallpaperPath(itemPath)
}

func GetIconPath(parentPath string, itemPath string) string {
	actualIconPath := genIconPath(parentPath, itemPath)
	if DoesFileExists(actualIconPath) {
		return actualIconPath
	}
	return ""
}

func CheckIconPath(parentPath string, itemPath string) bool {
	actualIconPath := genIconPath(parentPath, itemPath)
	return DoesFileExists(actualIconPath)
}

func GetTrueIconPath(parentPath string, itemPath string) string {
	return genIconPath(parentPath, itemPath)
}

func genWallpaperPath(itemPath string) string {
	return itemPath + "/.media/bg.png"
}

func genListWallpaperPath(itemPath string) string {
	return itemPath + "/.media/bglist.png"
}

func genIconPath(parentPath string, itemPath string) string {
	itemName := GetSimpleFileName(itemPath)
	if itemName == ToolsName || itemName == ToolsTag {
		parentPath = "/mnt/SDCARD/Tools"
		itemName = ToolsTag
	}
	if itemName == RecentlyPlayedName || itemName == RecentlyPlayedTag {
		parentPath = "/mnt/SDCARD"
		itemName = RecentlyPlayedTag
	}
	if itemName == CollectionsDisplayName || itemName == CollectionsTag {
		parentPath = "/mnt/SDCARD"
		itemName = CollectionsTag
	}
	return parentPath + "/.media/" + itemName + ".png"
}

func GetCurrentDecorationDetails(directoryList []shared.RomDirectory) (topLevel bool, currentPath string, parentPath string) {
	topLevel = true
	currentDirectory := directoryList[len(directoryList) - 1]
	currentPath = currentDirectory.Path
	parentPath = GetRomDirectory()

	if len(directoryList) > 1 {
		topLevel = false
		parentDirectory := directoryList[len(directoryList) - 2]
		parentPath = parentDirectory.Path
	}
	return topLevel, currentPath, parentPath
}

// copyFile copies a file from src to dst.
func CopyFile(sourcePath, destinationPath string) error {
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

func GetSimpleFileName(fullPath string) string {
	itemWithExt := filepath.Base(fullPath)
	return strings.TrimSuffix(itemWithExt, filepath.Ext(itemWithExt))
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
