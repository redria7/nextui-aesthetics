package utils

const (
	// gameTrackerDBPath  = "/mnt/SDCARD/.userdata/shared/game_logs.sqlite"
	// saveFileDirectory  = "/mnt/SDCARD/Saves/"
	// RecentlyPlayedFile = "/mnt/SDCARD/.userdata/shared/.minui/recent.txt"
	defaultDirPerm     = 0755
	defaultFilePerm    = 0644
	ExitCodeAction       = 4
	ExitCodeSelect         = 0
	ExitCodeCancel           = 2
	ExitCodeError         = -1
	RecentlyPlayedDirectory = "/mnt/SDCARD/Recently Played"
	AggregateByConsole	= 0
	AggregateByDirectory	= 1
)

var DecorationSources = []directorySource{
	directorySource{DirectoryPath: "/mnt/SDCARD/Tools/tg5040/Theme-Manager.pak/Themes", FilenamesTagFree: true},
	directorySource{DirectoryPath: "/mnt/SDCARD/Screenshots", FilenamesTagFree: false},
	//directorySource{DirectoryPath: "/mnt/SDCARD/Roms", FilenamesTagFree: true},
}

type directorySource struct {
	DirectoryPath		string
	FilenamesTagFree	bool
}