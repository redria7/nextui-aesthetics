package utils

import (
	"path/filepath"
	"regexp"
	"strings"
	"sort"
	"nextui-aesthetics/models"
)

const (
	softConsole	= "(misc)"
)

func GenerateDecorationAggregations() (consoleAggregationList []models.ConsoleAggregation, directoryAggregationList []models.DirectoryAggregation) {
	consoleAggregation := make(map[string]map[string][]models.Decoration)
	directoryAggregation := make(map[string][]models.Decoration)
	for _, decorationSource := range DecorationSources {
		if DoesFileExists(decorationSource.DirectoryPath) {
			consoleAggregation, directoryAggregation = collectNestedDecorations(
				consoleAggregation, 
				directoryAggregation,
				decorationSource.DirectoryPath,	// current path
				decorationSource,	// soft parent
				"",	// hard parent path
				"",	// hard console
			)
		}
	}
	// Sort aggregation map results into list structures for consistent display output
	// Sort directory aggregation
	directoryKeys := make([]string, len(directoryAggregation))
	keyIndex := 0
	for key, _ := range directoryAggregation {
		directoryKeys[keyIndex] = key
		keyIndex++
	}
	sort.Strings(directoryKeys)
	for _, key := range directoryKeys {
		newDirectoryAggregationStruct := models.DirectoryAggregation{
			DirectoryName: key,
			DecorationList: directoryAggregation[key],
		}
		directoryAggregationList = append(directoryAggregationList, newDirectoryAggregationStruct)
	}

	// Sort console aggregation
	consoleKeys := make([]string, len(consoleAggregation))
	keyIndex = 0
	for key, _ := range consoleAggregation {
		consoleKeys[keyIndex] = key
		keyIndex++
	}
	sort.Strings(consoleKeys)
	for _, key := range consoleKeys {
		// Sort console sub list
		consoleSubKeys := make([]string, len(consoleAggregation[key]))
		subKeyIndex := 0
		for subKey, _ := range consoleAggregation[key] {
			consoleSubKeys[subKeyIndex] = subKey
			subKeyIndex++
		}
		sort.Strings(consoleSubKeys)
		for _, subKey := range consoleSubKeys {
			newConsoleAggregationStruct := models.ConsoleAggregation{
				ConsoleTag: key,
				ConsoleName: subKey,
				DecorationList: consoleAggregation[key][subKey],
			}
			consoleAggregationList = append(consoleAggregationList, newConsoleAggregationStruct)
		}
	}

	return consoleAggregationList, directoryAggregationList
}

func collectNestedDecorations(
	consoleAggregation map[string]map[string][]models.Decoration, 
	directoryAggregation map[string][]models.Decoration, 
	currentPath string, 
	softParent directorySource, 
	hardParentPath string, 
	hardConsole string,
) (map[string]map[string][]models.Decoration, map[string][]models.Decoration) {
	// Collect current path files. If error, just return currently built aggregations
	files, err := GetFileList(currentPath)
	if err != nil {
		return consoleAggregation, directoryAggregation
	}
	
	// Determine console tag of current directory if possible
	if hardConsole == "" {
		hardConsole = findConsoleTag(currentPath)
	}

	// If no hard parent found yet, scan files for any valid decorations. If some are found, set the current path as the hard parent path
	if hardParentPath == "" {
		for _, file := range files {
			itemName := file.Name()
			itemExt := filepath.Ext(itemName)
			if itemExt == ".png" && itemName != "preview.png" {
				hardParentPath = currentPath
				break
			}
		}
	}

	// All preconditions are checked for the current directory: check each entry and drill down in any child directories
	for _, file := range files {
		if file.IsDir() {
			// Current file is a directory, pass current settings and drill down
			consoleAggregation, directoryAggregation = collectNestedDecorations(
				consoleAggregation, 
				directoryAggregation,
				filepath.Join(currentPath, file.Name()),	// current path
				softParent,	// soft parent path
				hardParentPath,	// hard parent path, default ""
				hardConsole,	// hard console, default ""
			)
		} else {
			// Current file is not a directory. Evaluate.
			itemName := file.Name()
			itemExt := filepath.Ext(itemName)
			if itemExt == ".png" {
				// Current file is a png. Valid decoration found. Create Decoration item and attach to maps
				// Generate formal path
				decorationPath := filepath.Join(currentPath, itemName)
				
				// Finalize console tag
				consoleTag := hardConsole
				if hardConsole == "" && softParent.FilenamesTagFree {
					tempTag := findConsoleTag(itemName)
					if tempTag != "" {
						consoleTag = tempTag
					}
				}
				if consoleTag == "" {
					consoleTag = softConsole
				}

				// Finalize console sub name
				consoleSubName := consoleTag
				if hardParentPath != "" {
					consoleSubName = consoleSubName + " " + filepath.Base(hardParentPath)
				}

				// Finalize directory name
				directoryName := filepath.Base(softParent.DirectoryPath)
				if hardParentPath != "" {
					directoryName = directoryName + "/" + filepath.Base(hardParentPath)
				}

				// Generate decoration names for each aggregation
				var directoryDecorationName string 
				if hardParentPath == "" {
					directoryDecorationName = strings.ReplaceAll(decorationPath, softParent.DirectoryPath, "")
				} else {
					directoryDecorationName = strings.ReplaceAll(decorationPath, hardParentPath, "")
				}
				directoryDecorationName = strings.TrimPrefix(directoryDecorationName, "/")

				consoleDecorationName := filepath.Base(softParent.DirectoryPath) + "/" + itemName


				// Generate decorations
				consoleDecoration := models.Decoration{
					DecorationName: consoleDecorationName, // For displaying to the user
					DecorationPath: decorationPath,	// For file magic + finding the decoration in either aggregation list
					ConsoleName: consoleSubName,	// For finding the decoration in the ConsoleAggregation list
					DirectoryName: directoryName, 	// For finding the decoration in the DirectoryAggregation list
				}
				directoryDecoration := models.Decoration{
					DecorationName: directoryDecorationName, // For displaying to the user
					DecorationPath: decorationPath,	// For file magic + finding the decoration in either aggregation list
					ConsoleName: consoleSubName,	// For finding the decoration in the ConsoleAggregation list
					DirectoryName: directoryName, 	// For finding the decoration in the DirectoryAggregation list
				}

				// Attach decorations to map
				if consoleAggregation[consoleTag] == nil {
					consoleAggregation[consoleTag] = make(map[string][]models.Decoration)
				}
				consoleAggregation[consoleTag][consoleSubName] = append(consoleAggregation[consoleTag][consoleSubName], consoleDecoration)
				directoryAggregation[directoryName] = append(directoryAggregation[directoryName], directoryDecoration)
			}
		}
	}

	// Return updated maps
	return consoleAggregation, directoryAggregation
}

func findConsoleTag(directoryPath string) string {
	re := regexp.MustCompile(`\(([^)]+)\)`)
	match := re.FindStringSubmatch(directoryPath)
	if len(match) > 1 {
		return match[0]
	}
	return ""
}