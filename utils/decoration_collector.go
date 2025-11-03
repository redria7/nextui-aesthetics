package utils

import (
	"errors"
	"nextui-aesthetics/models"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"strconv"

	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	gaba "github.com/redria7/gabagool/pkg/gabagool"
)

const (
	softConsole	= "(misc)"
	folderDelimiter = "[~]"
	consoleDelimiter = "[-]"
	consoleDelimitedCountDefault = -1
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
				decorationSource,	// original parent
				"", // soft parent path
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
	originalParent directorySource, 
	softParentPath string,
	hardParentPath string, 
	hardConsole string,
) (map[string]map[string][]models.Decoration, map[string][]models.Decoration) {
	// Collect current path files. If error, just return currently built aggregations
	files, err := GetFileList(currentPath)
	if err != nil {
		return consoleAggregation, directoryAggregation
	}
	
	// Determine console tag of current directory if possible
	portsFolder := false
	if hardConsole == "" {
		hardConsole = FindConsoleTag(currentPath)
		portsFolder = (hardConsole == "(PORTS)" && originalParent.DirectoryPath == "/mnt/SDCARD/Roms")
	}

	// If no hard parent found yet, scan files for any valid decorations. If some are found, set the current path as the hard parent path
	if hardParentPath == "" {
		for _, file := range files {
			itemName := file.Name()
			itemExt := filepath.Ext(itemName)
			if itemExt == ".png" && itemName != previewStandardName && itemName != previewHiddenName{
				hardParentPath = currentPath
				break
			}
		}
	}

	if softParentPath == "" && currentPath != originalParent.DirectoryPath {
		softParentPath = currentPath
	}

	// All preconditions are checked for the current directory: check each entry and drill down in any child directories
	for _, file := range files {
		if file.IsDir() && (!portsFolder || file.Name() == ".media") {
			// Current file is a directory, pass current settings and drill down
			consoleAggregation, directoryAggregation = collectNestedDecorations(
				consoleAggregation, 
				directoryAggregation,
				filepath.Join(currentPath, file.Name()),	// current path
				originalParent,	// original parent
				softParentPath,	// soft parent path
				hardParentPath,	// hard parent path, default ""
				hardConsole,	// hard console, default ""
			)
		} else {
			// Current file is not a directory. Evaluate.
			itemName := file.Name()
			itemExt := filepath.Ext(itemName)
			// Build conditions
			isDecoration := itemExt == ".png"
			isPreview := itemName == previewStandardName || itemName == previewHiddenName
			isMedia := filepath.Base(currentPath) == ".media"
			isMediaBg := isMedia && itemName == "bg.png"
			isMediaBgList := isMedia && itemName == "bglist.png"
			isFolderIcon := false
			// If a .media non-wallpaper image is found, check to see if the icon target is probably self contained
			if isDecoration && isMedia && !isMediaBg && !isMediaBgList {
				isFolderIcon = checkIfFolderIcon(filepath.Dir(currentPath), itemName, itemExt)
			}
			if isDecoration && !isPreview && !isMediaBg && !isMediaBgList && !isFolderIcon {
				// Current file is a png. Valid decoration found. Create Decoration item and attach to maps
				// Finalize soft parent
				softParent := softParentPath
				if softParent == "" {
					softParent = originalParent.DirectoryPath
				}

				// Generate formal path
				decorationPath := filepath.Join(currentPath, itemName)
				
				// Finalize console tag
				consoleTag := hardConsole
				if hardConsole == "" && originalParent.FilenamesTagFree {
					tempTag := FindConsoleTag(itemName)
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
				directoryName := filepath.Base(softParent)
				if hardParentPath != "" {
					directoryName = directoryName + "/" + filepath.Base(hardParentPath)
				}

				// Generate decoration names for each aggregation
				var directoryDecorationName string 
				if hardParentPath == "" {
					directoryDecorationName = strings.ReplaceAll(decorationPath, softParent, "")
				} else {
					directoryDecorationName = strings.ReplaceAll(decorationPath, hardParentPath, "")
				}
				directoryDecorationName = strings.TrimPrefix(directoryDecorationName, "/")

				consoleDecorationName := filepath.Base(softParent) + "/" + itemName


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

func checkIfFolderIcon(mediaParent string, itemName string, itemExt string) bool {
	// For tools folder, always treat as icon. we do not dive deeper
	if strings.HasPrefix(mediaParent, ToolsDirectory) {
		return true
	}
	// For collections folder, always treat as icon. no need for further checks
	if strings.HasPrefix(mediaParent, GetCollectionDirectory()) {
		return true
	}
	// If a .media non-wallpaper image is found, check to see if the icon target is probably self contained
	itemBase := strings.TrimSuffix(itemName, itemExt)
	iconTarget := filepath.Join(mediaParent, itemBase)
	stats, err := os.Stat(iconTarget)
	if err == nil && stats.IsDir() {
		iconChildPattern := filepath.Join(iconTarget, itemBase + ".*")
		matches, err := filepath.Glob(iconChildPattern)
		if err == nil && len(matches) == 0 {
			return true
		}
	}
	return false
}

func collectComponentsByNestedDirectoryForCurrentTheme(componentList []models.Component, componentHomeDirectory string, currentPath string) []models.Component {
	//logger := common.GetLoggerInstance()

	// Check if current directory meets component checks for missing components
	mediaDirectory := filepath.Join(currentPath, ".media")
	if DoesFileExists(mediaDirectory) {
		files, err := GetFileList(mediaDirectory)
		if err == nil {
			// Generate shortened list of components matching the home directory, that are not yet supported, and that are not duplicate types
			componentsInReview := map[int]models.Component{}
			for componentIndex, component := range componentList {
				if component.ComponentType.ComponentHomeDirectory == componentHomeDirectory && !component.IsSupported && !component.ComponentType.DuplicateType {
					componentsInReview[componentIndex] = component
				}
			}

			// Check for each content type in each file
			bgFound := false
			bgListFound := false
			iconFound := false
			for _, file := range files {
				itemName := file.Name()
				itemExt := filepath.Ext(itemName)
				// Skip meta files as they are already reviewed
				if !isMetaFile(filepath.Join(mediaDirectory, itemName)) {
					// Build conditions
					isDecoration := itemExt == ".png"
					isMediaBg := itemName == "bg.png"
					isMediaBgList := itemName == "bglist.png"
					isFolderIcon := false
					// If a .media non-wallpaper image is found, check to see if the icon target is probably self contained
					if isDecoration && !isMediaBg && !isMediaBgList {
						isFolderIcon = checkIfFolderIcon(filepath.Dir(mediaDirectory), itemName, itemExt)
					}
					// Update loop status
					if isMediaBg {
						bgFound = true
					}
					if isMediaBgList {
						bgListFound = true
					}
					if isFolderIcon {
						iconFound = true
					}
				}
			}

			// With content types checked, update component list and components in review
			recurseDeeper := false
			for componentListIndex, component := range componentsInReview {
				switch component.ComponentType.ComponentType {
					case ComponentTypeIcon:
						if iconFound {
							component.IsSupported = true
							componentList[componentListIndex] = component
						} else {
							recurseDeeper = true
						}
					case ComponentTypeWallpaper:
						if bgFound {
							component.IsSupported = true
							componentList[componentListIndex] = component
						} else {
							recurseDeeper = true
						}
					case ComponentTypeListWallpaper:
						if bgListFound {
							component.IsSupported = true
							componentList[componentListIndex] = component
						} else {
							recurseDeeper = true
						}
				}
			}
			if !recurseDeeper {
				return componentList
			}
		}
	}

	// Recurse through child directories
	files, err := GetFileList(currentPath)
	if err != nil {
		return componentList
	}
	portsFolder := (FindConsoleTag(currentPath) == "(PORTS)" && componentHomeDirectory == "/mnt/SDCARD/Roms")
	toolsFolder := (componentHomeDirectory == ToolsDirectory && currentPath != componentHomeDirectory)
	for _, file := range files {
		if file.IsDir() && file.Name() != ".media" && !portsFolder && !toolsFolder {
			componentList = collectComponentsByNestedDirectoryForCurrentTheme(componentList, componentHomeDirectory, filepath.Join(currentPath, file.Name()))
		}
	}
	return componentList
}

func collectComponentsByDirectoryForCurrentTheme(componentList []models.Component) []models.Component {
	//logger := common.GetLoggerInstance()

	// Note component success for system hard coded files
	systemIcons := false
	systemWallpapers := false
	systemListWallpapers := false
	if DoesFileExists("/mnt/SDCARD/.media/Collections.png") || DoesFileExists("/mnt/SDCARD/.media/Recently Played.png") || DoesFileExists("/mnt/SDCARD/Tools/.media/tg5040.png") {
		systemIcons = true
	}
	if DoesFileExists("/mnt/SDCARD/bg.png") || DoesFileExists("/mnt/SDCARD/Collections/.media/bg.png") || DoesFileExists("/mnt/SDCARD/Recently Played/.media/bg.png") || DoesFileExists("/mnt/SDCARD/Tools/tg5040/.media/bg.png") {
		systemWallpapers = true
	}
	if DoesFileExists("/mnt/SDCARD/Collections/.media/bglist.png") || DoesFileExists("/mnt/SDCARD/Recently Played/.media/bglist.png") || DoesFileExists("/mnt/SDCARD/Tools/tg5040/.media/bglist.png") {
		systemIcons = true
	}
	if systemIcons || systemWallpapers || systemListWallpapers {
		for index, component := range componentList {
			switch component.ComponentName {
				case "SystemIcons":
					if systemIcons {
						component.IsSupported = true
						componentList[index] = component
					}
				case "SystemWallpapers":
					if systemWallpapers {
						component.IsSupported = true
						componentList[index] = component
					}
				case "SystemListWallpapers":
					if systemListWallpapers {
						component.IsSupported = true
						componentList[index] = component
					}
			}
		}
	}

	// Collect list of directories to search
	unProvenDirectories := map[string]bool{}
	for _, component := range componentList {
		if !component.IsSupported && !component.ComponentType.DuplicateType {
			unProvenDirectories[component.ComponentType.ComponentHomeDirectory] = true
		}
	}

	// Visit each directory and update the relevant components 
	for componentHomeDirectory, _ := range unProvenDirectories {
		componentList = collectComponentsByNestedDirectoryForCurrentTheme(componentList, componentHomeDirectory, componentHomeDirectory)
	}

	return componentList
}

func isMetaFile(directoryPath string) bool {
	return metaPaths[directoryPath]
}

func FindConsoleTag(directoryPath string) string {
	re := regexp.MustCompile(`\(([^)]+)\)`)
	match := re.FindStringSubmatch(directoryPath)
	if len(match) > 1 {
		return match[0]
	}
	return ""
}

func GetThemeComponents(theme models.Theme) []models.Component {
	isCurrentTheme := IsCurrentTheme(theme)
	
	// Prep sorted components for each supported theme
	componentTypeKeys := make([]string, len(ComponentTypes))
	keyIndex := 0
	for key, _ := range ComponentTypes {
		componentTypeKeys[keyIndex] = key
		keyIndex++
	}
	sort.Strings(componentTypeKeys)
	var componentList []models.Component
	for _, componentName := range componentTypeKeys {
		componentList = append(componentList, models.Component{
			ComponentName: componentName,
			IsSupported: false,
			ComponentPaths: []string{},
			ComponentType: ComponentTypes[componentName],
		})
	}
	
	// Collect component status
	if isCurrentTheme {
		// Check for components from current save data
		componentList = collectComponentsByDirectoryForCurrentTheme(componentList)
	} else {
		// Check for components from theme folder
		componentList = collectComponentsByDirectoryForSavedTheme(theme.ThemePath, componentList)
	}

	return componentList
}

func collectComponentsByDirectoryForSavedTheme(currentPath string, componentList []models.Component) []models.Component {
	// Get files in working path
	files, err := GetFileList(currentPath)
	if err != nil {
		return componentList
	}
	currentDirectory := filepath.Base(currentPath)

	// Check for component file types in current directory
	pngFound := false
	for _, file := range files {
		itemExt := filepath.Ext(file.Name())
		if itemExt == ".png" {
			pngFound = true
		}
		if pngFound {
			break
		}
	}

	// If component file type found, try to match current directory to component type
	if pngFound {
		for index, component := range componentList {
			if component.ComponentName == currentDirectory {
				component.ComponentPaths = append(component.ComponentPaths, currentPath)
				component.IsSupported = true
				componentList[index] = component
				return componentList
			}
		}
	}

	// If current directory does not contain a component file type, or does not match a component type, recurse for all directories
	for _, file := range files {
		if file.IsDir() {
			componentList = collectComponentsByDirectoryForSavedTheme(filepath.Join(currentPath, file.Name()), componentList)
		}
	}
	return componentList
}

func ApplyThemeComponentUpdates(theme models.Theme, components []models.Component, options models.ComponentOptionSelections) (string, error, int) {
	isCurrentTheme := IsCurrentTheme(theme)
	modifyCount := 0

	if options.OptionClear {
		if ConfirmAction("Begin resetting requested components?", "") {
			// Clear requested stuff
			if !options.OptionConfirm {
				res, err := gaba.ProcessMessage("Resetting requested components to default", gaba.ProcessMessageOptions{}, func() (interface{}, error) {
					count, err := resetToDefaultRequestedComponents(components, options)
					if err != nil {
						return count, err
					}
					return count, nil
				})
				modifyCount = modifyCount + res.Result.(int)
				if err != nil {
					return "Reverting to Defaults", err, modifyCount
				}
			} else {
				count, err := resetToDefaultRequestedComponents(components, options)
				modifyCount = modifyCount + count
				if err != nil {
					return "Reverting to Defaults", err, modifyCount
				}
			}
		}
		
		// Current theme clears or saves. If clear is done for current theme, then return
		if isCurrentTheme {
			return "", nil, modifyCount
		}
	}

	// If current theme and not exited yet, then save and exit
	if isCurrentTheme {
		if ConfirmAction("Begin saving requested components?", "") {
			// Save current theme
			themeName, err := generateThemeName()
			if err != nil {
				return "Saving Current Theme", err, modifyCount
			}
			if !options.OptionConfirm {
				res, err := gaba.ProcessMessage("Saving requested components to new theme " + themeName, gaba.ProcessMessageOptions{}, func() (interface{}, error) {
					count, err := saveCurrentTheme(components, options, themeName)
					if err != nil {
						return count, err
					}
					return count, nil
				})
				modifyCount = modifyCount + res.Result.(int)
				if err != nil {
					return "Saving Current Theme", err, modifyCount
				}
			} else {
				count, err := saveCurrentTheme(components, options, themeName)
				modifyCount = modifyCount + count
				if err != nil {
					return "Saving Current Theme", err, modifyCount
				}
			}
		}
		return "", nil, modifyCount
	}

	// Apply selected theme
	if ConfirmAction("Begin applying requested components?", "") {
		if !options.OptionConfirm {
			res, err := gaba.ProcessMessage("Applying requested components from theme " + theme.ThemeName, gaba.ProcessMessageOptions{}, func() (interface{}, error) {
				count, err := applySelectedThemeComponents(theme, components, options)
				if err != nil {
					return count, err
				}
				return count, nil
			})
			modifyCount = modifyCount + res.Result.(int)
			if err != nil {
				return "Applying Components", err, modifyCount
			}
		} else {
			// When run with the confirm option, don't wrap in a gaba process
			count, err := applySelectedThemeComponents(theme, components, options)
			modifyCount = modifyCount + count
			if err != nil {
				return "Applying Components", err, modifyCount
			}
		}
	}

	return "", nil, modifyCount
}

func generateThemeName() (string, error) {
	themeName := "LocalTheme-" + time.Now().Format("2006.01.02-15.04.05")
	attemptNumber := 1
	for {
		var suffix string
		if attemptNumber == 1 {
			suffix = ""
		} else {
			suffix = " (" + strconv.Itoa(attemptNumber) + ")"
		}
		if !DoesFileExists(filepath.Join(ThemesDirectory, themeName + suffix)) {
			return themeName + suffix, nil
		}
		if attemptNumber > 10 {
			return "", errors.New("No valid theme names available")
		}
		attemptNumber++
	}
}

func saveCurrentTheme(components []models.Component, options models.ComponentOptionSelections, themeName string) (int, error) {
	modifyCount := 0

	// Save meta components and build component directory/type maps for recursion
	homeDirectories := make(map[string]map[string]bool)
	for _, component := range components {
		if component.ComponentType.ContainsMetaFiles {
			switch component.ComponentType.ComponentType {
				case ComponentTypeIcon:
					modifyCount = modifyCount + saveThemeDecorationSafely("/mnt/SDCARD/.media/Collections.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Collections.png"), options.OptionConfirm)
					modifyCount = modifyCount + saveThemeDecorationSafely("/mnt/SDCARD/.media/Recently Played.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Recently Played.png"), options.OptionConfirm)
					modifyCount = modifyCount + saveThemeDecorationSafely("/mnt/SDCARD/Tools/.media/tg5040.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Tools.png"), options.OptionConfirm)
				case ComponentTypeWallpaper:
					modifyCount = modifyCount + saveThemeDecorationSafely("/mnt/SDCARD/Collections/.media/bg.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Collections.png"), options.OptionConfirm)
					modifyCount = modifyCount + saveThemeDecorationSafely("/mnt/SDCARD/Recently Played/.media/bg.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Recently Played.png"), options.OptionConfirm)
					modifyCount = modifyCount + saveThemeDecorationSafely("/mnt/SDCARD/Tools/tg5040/.media/bg.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Tools.png"), options.OptionConfirm)
					modifyCount = modifyCount + saveThemeDecorationSafely("/mnt/SDCARD/bg.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Root.png"), options.OptionConfirm)
				case ComponentTypeListWallpaper:
					modifyCount = modifyCount + saveThemeDecorationSafely("/mnt/SDCARD/Collections/.media/bglist.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Collections.png"), options.OptionConfirm)
					modifyCount = modifyCount + saveThemeDecorationSafely("/mnt/SDCARD/Recently Played/.media/bglist.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Recently Played.png"), options.OptionConfirm)
					modifyCount = modifyCount + saveThemeDecorationSafely("/mnt/SDCARD/Tools/tg5040/.media/bglist.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Tools.png"), options.OptionConfirm)
			}
		}
		// Add component directories and types to map while looping
		if homeDirectories[component.ComponentType.ComponentHomeDirectory] == nil {
			homeDirectories[component.ComponentType.ComponentHomeDirectory] = make(map[string]bool)
		}
		homeDirectories[component.ComponentType.ComponentHomeDirectory][component.ComponentType.ComponentType] = true
	}
	
	// Collect valid parent directories for rom dependent directories
	validParents := make(map[string][]string)
	parentsList, err := getTopLevelRomsDirectories(options.OptionActive)
	if err != nil {
		return modifyCount, err
	}
	for _, parent := range parentsList {
		parentConsole := FindConsoleTag(parent.Filename)
		if options.OptionInactive {
			if parent.DirectoryFileCount == 0 {
				validParents[parentConsole] = append(validParents[parentConsole], parent.Filename)
			}
		} else {
			validParents[parentConsole] = append(validParents[parentConsole], parent.Filename)
		}
	}

	// Save non meta components
	for homeDirectory, homeDirectoryComponentTypes := range homeDirectories {
		isRomDependent := checkComponentForRomsDependency(homeDirectory)
		modifyCount = modifyCount + saveDecorations(homeDirectory, isRomDependent, validParents, false, homeDirectoryComponentTypes, themeName, homeDirectory, options.OptionConfirm)
	}

	return modifyCount, nil
}

func genNumberedConsoleTag(consoleDirectory string, validRomParents map[string][]string) string {
	consoleTag := FindConsoleTag(consoleDirectory)
	if consoleTag == "" {
		return ""
	}
	validParentList := validRomParents[consoleTag]
	listLength := len(validParentList)
	for index, validParentDirectory := range validParentList {
		if validParentDirectory == consoleDirectory {
			if listLength == 1 {
				return consoleTag
			}
			return consoleTag + consoleDelimiter + strconv.Itoa(index)
		}
	}
	return ""
}

func collectConsoleDelimitedNumber(delimitedName string) int {
	consoleParts := strings.Split(delimitedName, consoleDelimiter)
	if len(consoleParts) <= 1 {
		return consoleDelimitedCountDefault
	}
	delimitedNumber := consoleParts[1]
	realNumber, err := strconv.Atoi(delimitedNumber)
	if err != nil {
		return consoleDelimitedCountDefault
	}
	return realNumber
}

func saveDecorations(currentPath string, isRomDependent bool, validRomParents map[string][]string, romParentValidated bool, componentTypes map[string]bool, themeName string, componentHomeDirectory string, optionConfirm bool) int {
	modifyCount := 0

	currentDirectory := filepath.Base(currentPath)
	isMedia := false
	if currentDirectory == ".media" {
		isMedia = true
	}
	
	files, err := GetFileList(currentPath)
	if err != nil {
		return modifyCount
	}

	for _, file := range files {
		// For each file in path, make decision about how to proceed
		itemName := file.Name()
		itemExt := filepath.Ext(itemName)
		if isMedia {
			// reset png files
			if itemExt == ".png" && !isMetaFile(filepath.Join(currentPath, itemName)) {
				// Build conditions
				isMediaBg := itemName == "bg.png"
				isMediaBgList := itemName == "bglist.png"
				isFolderIcon := false
				// If a .media non-wallpaper image is found, check to see if the icon target is probably self contained
				if !isMediaBg && !isMediaBgList {
					isFolderIcon = checkIfFolderIcon(filepath.Dir(currentPath), itemName, itemExt)
				}

				// Generate component prefix according to media type
				componentNamePrefix := ""
				if isMediaBg || isMediaBgList || isFolderIcon {
					switch componentHomeDirectory {
						case GetRomDirectory():
							componentNamePrefix = "System"
						case GetCollectionDirectory():
							componentNamePrefix = "Collection"
						case ToolsDirectory:
							componentNamePrefix = "Tool"
					}
				}

				// Generate starting path list
				decorationPathList := strings.Split(strings.TrimPrefix(currentPath, componentHomeDirectory), string(filepath.Separator))
				decorationPathList = decorationPathList[1:len(decorationPathList) - 1] // remove .media and empty leading element from leading separator

				// console name adjustment?
				// Copy file to theme directory with appropriate name
				if isMediaBg && componentTypes[ComponentTypeWallpaper] && len(decorationPathList) > 0 {
					if isRomDependent {
						decorationPathList[0] = genNumberedConsoleTag(decorationPathList[0], validRomParents)
					}
					if decorationPathList[0] != "" {
						destinationName := strings.Join(decorationPathList, folderDelimiter)
						modifyCount = modifyCount + saveThemeDecorationSafely(filepath.Join(currentPath, itemName), filepath.Join(ThemesDirectory, themeName, componentNamePrefix + "Wallpapers", destinationName + ".png"), optionConfirm)
					}
				}
				if isMediaBgList && componentTypes[ComponentTypeListWallpaper] && len(decorationPathList) > 0 {
					if isRomDependent {
						decorationPathList[0] = genNumberedConsoleTag(decorationPathList[0], validRomParents)
					}

					if decorationPathList[0] != "" {
						destinationName := strings.Join(decorationPathList, folderDelimiter)
						modifyCount = modifyCount + saveThemeDecorationSafely(filepath.Join(currentPath, itemName), filepath.Join(ThemesDirectory, themeName, componentNamePrefix + "ListWallpapers", destinationName + ".png"), optionConfirm)
					}
				}
				if isFolderIcon && componentTypes[ComponentTypeIcon] {
					iconDecorationPathList := append(decorationPathList, strings.TrimSuffix(itemName, itemExt))
					if isRomDependent {
						iconDecorationPathList[0] = genNumberedConsoleTag(iconDecorationPathList[0], validRomParents)
					}

					if iconDecorationPathList[0] != "" {
						destinationName := strings.Join(iconDecorationPathList, folderDelimiter)
						modifyCount = modifyCount + saveThemeDecorationSafely(filepath.Join(currentPath, itemName), filepath.Join(ThemesDirectory, themeName, componentNamePrefix + "Icons", destinationName + ".png"), optionConfirm)
					}
				}
			}
		} else {
			// recurse
			portsFolder := (FindConsoleTag(currentPath) == "(PORTS)" && componentHomeDirectory == "/mnt/SDCARD/Roms" && !(filepath.Dir(currentPath) == componentHomeDirectory || filepath.Dir(filepath.Dir(currentPath)) == componentHomeDirectory))
			toolsFolder := (componentHomeDirectory == ToolsDirectory && !(currentPath == componentHomeDirectory || filepath.Dir(currentPath) == componentHomeDirectory))
			if file.IsDir() && !portsFolder && !toolsFolder {
				if isRomDependent && !romParentValidated {
					itemConsole := FindConsoleTag(itemName)
					validParentList := validRomParents[itemConsole]
					for _, parentDirectory := range validParentList {
						if parentDirectory == itemName {
							romParentValidated = true
							break
						}
					}
				}
				if !isRomDependent || romParentValidated || itemName == ".media" {
					modifyCount = modifyCount + saveDecorations(filepath.Join(currentPath, itemName), isRomDependent, validRomParents, true, componentTypes, themeName, componentHomeDirectory, optionConfirm)
				}
			}
		}
	}

	return modifyCount
}

func saveThemeDecorationSafely(sourcePath string, destinationPath string, confirmCopy bool) int {
	if confirmCopy {
		destinationPathList := strings.Split(destinationPath, string(filepath.Separator))
		message := "Save " + filepath.Join(destinationPathList[len(destinationPathList) - 2:]...)
		if ConfirmActionCustomBack(message, sourcePath, "Skip") {
			err := CopyFile(sourcePath, destinationPath)
			if err == nil {
				return 1
			}
		}
		return 0
	}
	err := CopyFile(sourcePath, destinationPath)
	if err == nil {
		return 1
	}
	return 0
}

func resetThemeDecorationSafely(sourcePath string, confirmCopy bool) int {
	if confirmCopy {
		message := "Clear " + sourcePath
		if ConfirmActionCustomBack(message, sourcePath, "Skip") {
			if common.DeleteFile(sourcePath) {
				return 1
			}
		}
		return 0
	}
	if common.DeleteFile(sourcePath) {
		return 1
	}
	return 0
}

func resetToDefaultRequestedComponents(components []models.Component, options models.ComponentOptionSelections) (int, error) {
	modifyCount := 0
	// Reset meta components and build component directory/type maps for recursion
	homeDirectories := make(map[string]map[string]bool)
	for _, component := range components {
		if component.ComponentType.ContainsMetaFiles && !options.OptionInactive {
			switch component.ComponentType.ComponentType {
				case ComponentTypeIcon:
					modifyCount = modifyCount + resetThemeDecorationSafely("/mnt/SDCARD/.media/Collections.png", options.OptionConfirm)
					modifyCount = modifyCount + resetThemeDecorationSafely("/mnt/SDCARD/.media/Recently Played.png", options.OptionConfirm)
					modifyCount = modifyCount + resetThemeDecorationSafely("/mnt/SDCARD/Tools/.media/tg5040.png", options.OptionConfirm)
				case ComponentTypeWallpaper:
					modifyCount = modifyCount + resetThemeDecorationSafely("/mnt/SDCARD/Collections/.media/bg.png", options.OptionConfirm)
					modifyCount = modifyCount + resetThemeDecorationSafely("/mnt/SDCARD/Recently Played/.media/bg.png", options.OptionConfirm)
					modifyCount = modifyCount + resetThemeDecorationSafely("/mnt/SDCARD/Tools/tg5040/.media/bg.png", options.OptionConfirm)
					modifyCount = modifyCount + resetThemeDecorationSafely("/mnt/SDCARD/bg.png", options.OptionConfirm)
				case ComponentTypeListWallpaper:
					modifyCount = modifyCount + resetThemeDecorationSafely("/mnt/SDCARD/Collections/.media/bglist.png", options.OptionConfirm)
					modifyCount = modifyCount + resetThemeDecorationSafely("/mnt/SDCARD/Recently Played/.media/bglist.png", options.OptionConfirm)
					modifyCount = modifyCount + resetThemeDecorationSafely("/mnt/SDCARD/Tools/tg5040/.media/bglist.png", options.OptionConfirm)
			}
		}
		// Add component directories and types to map while looping
		if homeDirectories[component.ComponentType.ComponentHomeDirectory] == nil {
			homeDirectories[component.ComponentType.ComponentHomeDirectory] = make(map[string]bool)
		}
		homeDirectories[component.ComponentType.ComponentHomeDirectory][component.ComponentType.ComponentType] = true
	}
	
	// Collect valid parent directories for rom dependent directories
	validParents := make(map[string][]string)
	parentsList, err := getTopLevelRomsDirectories(options.OptionActive)
	if err != nil {
		return modifyCount, err
	}
	for _, parent := range parentsList {
		parentConsole := FindConsoleTag(parent.Filename)
		if options.OptionInactive {
			if parent.DirectoryFileCount == 0 {
				validParents[parentConsole] = append(validParents[parentConsole], parent.Filename)
			}
		} else {
			validParents[parentConsole] = append(validParents[parentConsole], parent.Filename)
		}
	}

	// Reset non meta components
	for homeDirectory, homeDirectoryComponentTypes := range homeDirectories {
		isRomDependent := checkComponentForRomsDependency(homeDirectory)
		if !options.OptionInactive || isRomDependent {
			modifyCount = modifyCount + resetDecorations(homeDirectory, isRomDependent, validParents, false, homeDirectoryComponentTypes, homeDirectory, options.OptionConfirm)
		}
	}

	return modifyCount, nil
}

func resetDecorations(currentPath string, isRomDependent bool, validRomParents map[string][]string, romParentValidated bool, componentTypes map[string]bool, componentHomeDirectory string, optionConfirm bool) int {
	modifyCount := 0
	currentDirectory := filepath.Base(currentPath)
	isMedia := false
	if currentDirectory == ".media" {
		isMedia = true
	}
	
	files, err := GetFileList(currentPath)
	if err != nil {
		return modifyCount
	}

	for _, file := range files {
		// For each file in path, make decision about how to proceed
		itemName := file.Name()
		itemExt := filepath.Ext(itemName)
		if isMedia {
			// reset png files
			if itemExt == ".png" && !isMetaFile(filepath.Join(currentPath, itemName)) {
				// Build conditions
				isMediaBg := itemName == "bg.png"
				isMediaBgList := itemName == "bglist.png"
				isFolderIcon := false
				// If a .media non-wallpaper image is found, check to see if the icon target is probably self contained
				if !isMediaBg && !isMediaBgList {
					isFolderIcon = checkIfFolderIcon(filepath.Dir(currentPath), itemName, itemExt)
				}
				// Check for matches to components then remove if found
				if isMediaBg && componentTypes[ComponentTypeWallpaper] {
					modifyCount = modifyCount + resetThemeDecorationSafely(filepath.Join(currentPath, itemName), optionConfirm)
				}
				if isMediaBgList && componentTypes[ComponentTypeListWallpaper] {
					modifyCount = modifyCount + resetThemeDecorationSafely(filepath.Join(currentPath, itemName), optionConfirm)
				}
				if isFolderIcon && componentTypes[ComponentTypeIcon] {
					if !romParentValidated {
						itemBase := strings.TrimSuffix(itemName, itemExt)
						itemConsole := FindConsoleTag(itemBase)
						validParentList := validRomParents[itemConsole]
						for _, parentDirectory := range validParentList {
							if parentDirectory == itemBase {
								modifyCount = modifyCount + resetThemeDecorationSafely(filepath.Join(currentPath, itemName), optionConfirm)
								break
							}
						}
					} else {
						modifyCount = modifyCount + resetThemeDecorationSafely(filepath.Join(currentPath, itemName), optionConfirm)
					}
				}
			}
		} else {
			// recurse
			portsFolder := (FindConsoleTag(currentPath) == "(PORTS)" && componentHomeDirectory == "/mnt/SDCARD/Roms" && !(filepath.Dir(currentPath) == componentHomeDirectory || filepath.Dir(filepath.Dir(currentPath)) == componentHomeDirectory))
			toolsFolder := (componentHomeDirectory == ToolsDirectory && !(currentPath == componentHomeDirectory || filepath.Dir(currentPath) == componentHomeDirectory))
			if file.IsDir() && !portsFolder && !toolsFolder {
				if isRomDependent && !romParentValidated {
					itemConsole := FindConsoleTag(itemName)
					validParentList := validRomParents[itemConsole]
					for _, parentDirectory := range validParentList {
						if parentDirectory == itemName {
							romParentValidated = true
							break
						}
					}
				}
				if !isRomDependent || romParentValidated || itemName == ".media" {
					modifyCount = modifyCount + resetDecorations(filepath.Join(currentPath, itemName), isRomDependent, validRomParents, true, componentTypes, componentHomeDirectory, optionConfirm)
				}
			}
		}
	}

	return modifyCount
}

func applySelectedThemeComponents(theme models.Theme, components []models.Component, options models.ComponentOptionSelections) (int, error) {
	modifyCount := 0
	
	// Collect valid parent directories for non-meta components
	validParents := make(map[string][]string)
	parentsList, err := getTopLevelRomsDirectories(options.OptionActive)
	if err != nil {
		return modifyCount, err
	}
	for _, parent := range parentsList {
		parentConsole := FindConsoleTag(parent.Filename)
		if options.OptionInactive {
			if parent.DirectoryFileCount == 0 {
				validParents[parentConsole] = append(validParents[parentConsole], parent.Filename)
			}
		} else {
			validParents[parentConsole] = append(validParents[parentConsole], parent.Filename)
		}
	}

	// For each component, 
	for _, component := range components {
		isRomDependent := checkComponentForRomsDependency(component.ComponentType.ComponentHomeDirectory)
		romParentSet := make(map[string]bool)
		for _, componentPath := range component.ComponentPaths {
			// For each path in a component, apply all decorations in the path according to the rules
			files, err := GetFileList(componentPath)
			if err == nil {
				for _, file := range files {
					itemName := file.Name()
					itemExt := filepath.Ext(itemName)
					if itemExt == ".png" {
						metaFileCopied := false
						if component.ComponentType.ContainsMetaFiles {
							switch component.ComponentType.ComponentType {
								case ComponentTypeIcon:
									switch itemName {
										case "Collections.png":
											modifyCount = modifyCount + applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/.media/Collections.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
										case "Recently Played.png":
											modifyCount = modifyCount + applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/.media/Recently Played.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
										case "Tools.png":
											modifyCount = modifyCount + applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/Tools/.media/tg5040.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
									}
								case ComponentTypeWallpaper:
									switch itemName {
										case "Collections.png":
											modifyCount = modifyCount + applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/Collections/.media/bg.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
										case "Recently Played.png":
											modifyCount = modifyCount + applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/Recently Played/.media/bg.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
										case "Tools.png":
											modifyCount = modifyCount + applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/Tools/tg5040/.media/bg.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
										case "Root.png":
											modifyCount = modifyCount + applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/bg.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
									}
								case ComponentTypeListWallpaper:
									switch itemName {
										case "Collections.png":
											modifyCount = modifyCount + applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/Collections/.media/bglist.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
										case "Recently Played.png":
											modifyCount = modifyCount + applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/Recently Played/.media/bglist.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
										case "Tools.png":
											modifyCount = modifyCount + applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/Tools/tg5040/.media/bglist.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
									}
							}
						}
						if !metaFileCopied {
							// File is not a meta file, move if possible
							itemBase := strings.TrimSuffix(itemName, itemExt)
							filePathParts := strings.Split(itemBase, folderDelimiter)
							filePathPartsList := [][]string{}
							tryToPlace := true
							if isRomDependent {
								// Rom dependent file. Replace the first piece of the name with the user's console directory. 
								// If the item is for that console directtory and is not the first item for that console directory, then skip.
								consoleTag := FindConsoleTag(filePathParts[0])
								romParentSetTag := consoleTag
								parentNumber := consoleDelimitedCountDefault
								if consoleTag != filePathParts[0] {
									parentNumber = collectConsoleDelimitedNumber(filePathParts[0])
									if parentNumber != consoleDelimitedCountDefault {
										romParentSetTag = romParentSetTag + strconv.Itoa(parentNumber)
									}
								}
								parentConsoleNameList := validParents[consoleTag]
								parentConsoleNameListLength := len(parentConsoleNameList)
								if parentConsoleNameListLength == 0 {
									tryToPlace = false
								} else {
									// Add path parts to list of targets for every valid match
									if parentNumber == consoleDelimitedCountDefault {
										for _, parentConsoleDirectory := range parentConsoleNameList {
											singlefilePathParts := append([]string{}, filePathParts...)
											singlefilePathParts[0] = parentConsoleDirectory
											filePathPartsList = append(filePathPartsList, singlefilePathParts)
										}
									} else {
										if parentConsoleNameListLength > parentNumber && parentNumber >= 0 {
											singlefilePathParts := append([]string{}, filePathParts...)
											singlefilePathParts[0] = parentConsoleNameList[parentNumber]
											filePathPartsList = append(filePathPartsList, singlefilePathParts)
										}
									}
									// don't place if this image name format has been placed already
									if len(filePathParts) == 1 {
										if romParentSet[romParentSetTag] {
											tryToPlace = false
										} else {
											romParentSet[romParentSetTag] = true
										}
									}
								}
							} else {
								filePathPartsList = append(filePathPartsList, filePathParts)
							}
							if tryToPlace {
								for _, filePathPartsIndividual := range filePathPartsList {
									// file has a valid parent. Move if the parent directory exists
									filePathPartsIndividual = append([]string{component.ComponentType.ComponentHomeDirectory}, filePathPartsIndividual...)
									parentConsoleDirectory := filepath.Join(filePathPartsIndividual...)
									if component.ComponentType.ComponentHomeDirectory == GetCollectionDirectory() {
										if !DoesFileExists(parentConsoleDirectory) {
											parentConsoleDirectory = parentConsoleDirectory + ".txt"
										}
									}
									if component.ComponentType.ComponentHomeDirectory == ToolsDirectory {
										if !DoesFileExists(parentConsoleDirectory) {
											parentConsoleDirectory = parentConsoleDirectory + ".pak"
										}
									}
									if DoesFileExists(parentConsoleDirectory) {
										// File exists. Collect the destination path then copy
										destinationPath := ""
										switch component.ComponentType.ComponentType {
											case ComponentTypeIcon:
												parentPath := filepath.Join(filePathPartsIndividual[:len(filePathPartsIndividual) - 1]...)
												destinationPath = GetTrueIconPath(parentPath, parentConsoleDirectory)
											case ComponentTypeWallpaper:
												destinationPath = GetTrueWallpaperPath(parentConsoleDirectory)
											case ComponentTypeListWallpaper:
												destinationPath = GetTrueListWallpaperPath(parentConsoleDirectory)
										}
										if destinationPath != "" {
											modifyCount = modifyCount + applyThemeDecorationSafely(filepath.Join(componentPath, itemName), destinationPath, options.OptionPreserve, options.OptionConfirm)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return modifyCount, nil
}

func applyThemeDecorationSafely(sourcePath string, destinationPath string, existencePreCheck bool, confirmCopy bool) int {
	if confirmCopy {
		sourcePathList := strings.Split(sourcePath, string(filepath.Separator))
		message := "Apply " + filepath.Join(sourcePathList[len(sourcePathList) - 2:]...)
		if ConfirmActionCustomBack(message, sourcePath, "Skip") {
			err := CopyFile(sourcePath, destinationPath)
			if err == nil {
				return 1
			}
		}
		return 0
	}

	if existencePreCheck {
		if DoesFileExists(destinationPath) {
			return 0
		}
	}
	err := CopyFile(sourcePath, destinationPath)
	if err == nil {
		return 1
	}
	return 0
}

func checkComponentForRomsDependency(componentHomeDirectory string) bool {
	return componentHomeDirectory == GetRomDirectory()
}

func getTopLevelRomsDirectories(onlyActive bool) (shared.Items, error) {
	logger := common.GetLoggerInstance()
	fb := filebrowser.NewFileBrowser(logger)


	// TODO: check user settings for hide empty
	if err := fb.CWD(GetRomDirectory(), onlyActive); err != nil {
		common.LogStandardFatal("Error fetching directories", err)
		return nil, err
	}

	return fb.Items, nil
}

func IsCurrentTheme(theme models.Theme) bool {
	if theme != (models.Theme{}) {
		return false
	}
	return true
}