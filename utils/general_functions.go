package utils

import (
	"os"
	"path/filepath"
	"strings"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
)

func IsDev() bool {
	return os.Getenv("ENVIRONMENT") == "DEV"
}

func GetCollectionDirectory() string {
	dir := common.CollectionDirectory
	if IsDev() {
		dir = os.Getenv("COLLECTION_DIRECTORY")
	}

	_ = EnsureDirectoryExists(dir)
	return dir
}

func GetRomDirectory() string {
	if IsDev() {
		return os.Getenv("ROM_DIRECTORY")
	}
	return common.RomDirectory
}

func CreateRomDirectoryFromItem(item shared.Item) shared.RomDirectory {
	return shared.RomDirectory{
		DisplayName: item.DisplayName,
		Tag:         item.Tag,
		Path:        item.Path,
	}
}

func CheckIfCollectionTxtChild(itemPath string) bool {
	return filepath.Ext(filepath.Base(itemPath)) == ".txt" && strings.HasPrefix(itemPath, GetCollectionDirectory())
}