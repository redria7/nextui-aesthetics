package models

import (
	"go.uber.org/zap/zapcore"
)

type AppState struct {
	Config *Config

	MenuPositionList []MenuPositionPointer

	DecorationsAggregatedOnConsoles []ConsoleAggregation
	DecorationsAggregatedOnDirectories []DirectoryAggregation

	// GamePlayMap 	map[string][]PlayHistoryAggregate
	// ConsolePlayMap 	map[string]int
	// TotalPlay 		int

	// CollectionMap	map[string][]Collection
}

func (a AppState) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	_ = enc.AddObject("config", a.Config)

	return nil
}

type MenuPositionPointer struct {
	SelectedIndex		int
	SelectedPosition    int
}

type Decoration struct {
	DecorationName	string	// For displaying to the user
	DecorationPath	string	// For file magic + finding the decoration in either aggregation list
	ConsoleName 	string	// For finding the decoration in the ConsoleAggregation list
	DirectoryName 	string	// For finding the decoration in the DirectoryAggregation list
}

type ConsoleAggregation struct {
	ConsoleTag	string
	ConsoleName	string
	DecorationList	[]Decoration
}

type DirectoryAggregation struct {
	DirectoryName	string
	DecorationList 	[]Decoration
}
