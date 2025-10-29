package utils

import "nextui-aesthetics/models"

const (
	// gameTrackerDBPath  = "/mnt/SDCARD/.userdata/shared/game_logs.sqlite"
	// saveFileDirectory  = "/mnt/SDCARD/Saves/"
	// RecentlyPlayedFile = "/mnt/SDCARD/.userdata/shared/.minui/recent.txt"
	defaultDirPerm             = 0755
	defaultFilePerm            = 0644
	ExitCodeAction             = 4
	ExitCodeSelect             = 0
	ExitCodeCancel             = 2
	ExitCodeError              = -1
	RecentlyPlayedDirectory    = "/mnt/SDCARD/Recently Played"
	ToolsDirectory             = "/mnt/SDCARD/Tools/tg5040"
	AggregateByConsole         = 1
	AggregateByDirectory       = 0
	CollectionsDisplayName     = "Collections"
	CollectionsTag             = "Collections"
	RecentlyPlayedName         = "Recently Played"
	RecentlyPlayedTag          = "Recently Played"
	ToolsName                  = "Tools"
	ToolsTag                   = "tg5040"
	ComponentTypeIcon          = "Icon"
	ComponentTypeWallpaper     = "Wallpaper"
	ComponentTypeListWallpaper = "ListWallpaper"
	ThemesDirectory			   = "/mnt/SDCARD/.userdata/shared/Aesthetics/Themes"
)

var ComponentTypes = map[string]models.ComponentTypeDetails{
	"SystemIcons": models.ComponentTypeDetails{
		ComponentType: ComponentTypeIcon,
		ComponentHomeDirectory: GetRomDirectory(),
		ContainsMetaFiles: true,
	},
	"CollectionIcons": models.ComponentTypeDetails{
		ComponentType: ComponentTypeIcon,
		ComponentHomeDirectory: GetCollectionDirectory(),
		ContainsMetaFiles: false,
	},
	"ToolIcons": models.ComponentTypeDetails{
		ComponentType: ComponentTypeIcon,
		ComponentHomeDirectory: ToolsDirectory,
		ContainsMetaFiles: false,
	},
	"SystemWallpapers": models.ComponentTypeDetails{
		ComponentType: ComponentTypeWallpaper,
		ComponentHomeDirectory: GetRomDirectory(),
		ContainsMetaFiles: true,
	},
	"CollectionWallpapers": models.ComponentTypeDetails{
		ComponentType: ComponentTypeWallpaper,
		ComponentHomeDirectory: GetCollectionDirectory(),
		ContainsMetaFiles: false,
	},
	"ToolWallpapers": models.ComponentTypeDetails{
		ComponentType: ComponentTypeWallpaper,
		ComponentHomeDirectory: ToolsDirectory,
		ContainsMetaFiles: false,
	},
	"SystemListWallpapers": models.ComponentTypeDetails{
		ComponentType: ComponentTypeListWallpaper,
		ComponentHomeDirectory: GetRomDirectory(),
		ContainsMetaFiles: true,
	},
	"ListWallpapers": models.ComponentTypeDetails{
		ComponentType: ComponentTypeListWallpaper,
		ComponentHomeDirectory: GetRomDirectory(),
		ContainsMetaFiles: true,
		DuplicateType: true,
	},
	"CollectionListWallpapers": models.ComponentTypeDetails{
		ComponentType: ComponentTypeListWallpaper,
		ComponentHomeDirectory: GetCollectionDirectory(),
		ContainsMetaFiles: false,
	},
	"ToolListWallpapers": models.ComponentTypeDetails{
		ComponentType: ComponentTypeListWallpaper,
		ComponentHomeDirectory: ToolsDirectory,
		ContainsMetaFiles: false,
	},
}

var metaPaths = map[string]bool{
	"/mnt/SDCARD/bg.png": true,
	"/mnt/SDCARD/.media/Collections.png": true,
	"/mnt/SDCARD/Collections/.media/bg.png": true,
	"/mnt/SDCARD/Collections/.media/bglist.png": true,
	"/mnt/SDCARD/.media/Recently Played.png": true,
	"/mnt/SDCARD/Recently Played/.media/bg.png": true,
	"/mnt/SDCARD/Recently Played/.media/bglist.png": true,
	"/mnt/SDCARD/Tools/.media/tg5040.png": true,
	"/mnt/SDCARD/Tools/tg5040/.media/bg.png": true,
	"/mnt/SDCARD/Tools/tg5040/.media/bglist.png": true,
}

var DecorationSources = []directorySource{
	directorySource{DirectoryPath: ThemesDirectory, FilenamesTagFree: true},
	directorySource{DirectoryPath: "/mnt/SDCARD/Screenshots", FilenamesTagFree: false},
	directorySource{DirectoryPath: "/mnt/SDCARD/Roms", FilenamesTagFree: false},
}

type directorySource struct {
	DirectoryPath    string
	FilenamesTagFree bool
}