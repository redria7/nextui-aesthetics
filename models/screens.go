package models

import "qlova.tech/sum"

type ScreenName struct {
	DirectoryBrowser,
	DecorationOptions,

	MainMenu sum.Int[ScreenName]
	// Settings,
	// Tools,

	// GamesList,
	// SearchBox,
	// Actions,
	// BulkActions,
	// AddToCollection,
	// Confirm,
	// DownloadArt,

	// AddToArchive,
	// ArchiveCreate,
	// ArchiveList,
	// ArchiveManagement,
	// ArchiveOptions,
	// ArchiveGamesList,

	// CollectionsList,
	// CollectionOptions,
	// CollectionManagement,
	// CollectionCreate,

	// PlayHistoryActions,
	// PlayHistoryGameDetails,
	// PlayHistoryGameHistory,
	// PlayHistoryGameList,
	// PlayHistoryList,
	// PlayHistoryFilter,

	// GlobalActions sum.Int[ScreenName]
}

var ScreenNames = sum.Int[ScreenName]{}.Sum()

type Screen interface {
	Name() sum.Int[ScreenName]
	Draw() (value interface{}, exitCode int, e error)
}
