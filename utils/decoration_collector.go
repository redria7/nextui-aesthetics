package utils

import (
	"errors"
	"fmt"
	"nextui-aesthetics/models"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	gaba "github.com/redria7/gabagool/pkg/gabagool"
)

const (
	softConsole	= "(misc)"
	folderDelimiter = "`~`"
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

func ApplyThemeComponentUpdates(theme models.Theme, components []models.Component, options models.ComponentOptionSelections) (string, error) {
	isCurrentTheme := IsCurrentTheme(theme)

	if options.OptionClear {
		// Clear requested stuff
		_, err := gaba.ProcessMessage("Resetting requested components to default", gaba.ProcessMessageOptions{}, func() (interface{}, error) {
			err := resetToDefaultRequestedComponents(components, options)
			if err != nil {
				return nil, err
			}
			return nil, nil
		})
		if err != nil {
			return "Reverting to Defaults", err
		}
		
		// Current theme clears or saves. If clear is done for current theme, then return
		if isCurrentTheme {
			return "", nil
		}
	}

	// If current theme and not exited yet, then save and exit
	if isCurrentTheme {
		// Save current theme
		themeName, err := generateThemeName()
		if err != nil {
			return "Saving Current Theme", err
		}
		_, err = gaba.ProcessMessage("Saving requested components to new theme " + themeName, gaba.ProcessMessageOptions{}, func() (interface{}, error) {
			err := saveCurrentTheme(components, options, themeName)
			if err != nil {
				return nil, err
			}
			return nil, nil
		})
		if err != nil {
			return "Saving Current Theme", err
		}
		return "", nil
	}

	// Apply selected theme
	if !options.OptionConfirm {
		_, err := gaba.ProcessMessage("Applying requested components from theme " + theme.ThemeName, gaba.ProcessMessageOptions{}, func() (interface{}, error) {
			err := applySelectedThemeComponents(theme, components, options)
			if err != nil {
				return nil, err
			}
			return nil, nil
		})
		if err != nil {
			return "Applying Components", err
		}
	} else {
		// When run with the confirm option, don't wrap in a gaba process
		err := applySelectedThemeComponents(theme, components, options)
		if err != nil {
			return "Applying Components", err
		}
	}

	return "", nil
}

func generateThemeName() (string, error) {
	themeName := "LocalTheme-" + time.Now().Format("2006.01.02-15.04.05")
	attemptNumber := 1
	for {
		var suffix string
		if attemptNumber == 1 {
			suffix = ""
		} else {
			suffix = " (" + string(attemptNumber) + ")"
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

func saveCurrentTheme(components []models.Component, options models.ComponentOptionSelections, themeName string) error {
	logger := common.GetLoggerInstance()
	logger.Debug("components: " + fmt.Sprint(components))

	// Save meta components and build component directory/type maps for recursion
	homeDirectories := make(map[string]map[string]bool)
	for _, component := range components {
		if component.ComponentType.ContainsMetaFiles {
			switch component.ComponentType.ComponentType {
				case ComponentTypeIcon:
					CopyFile("/mnt/SDCARD/.media/Collections.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Collections.png"))
					CopyFile("/mnt/SDCARD/.media/Recently Played.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Recently Played.png"))
					CopyFile("/mnt/SDCARD/Tools/.media/tg5040.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Tools.png"))
				case ComponentTypeWallpaper:
					CopyFile("/mnt/SDCARD/Collections/.media/bg.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Collections.png"))
					CopyFile("/mnt/SDCARD/Recently Played/.media/bg.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Recently Played.png"))
					CopyFile("/mnt/SDCARD/Tools/tg5040/.media/bg.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Tools.png"))
					CopyFile("/mnt/SDCARD/bg.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Root.png"))
				case ComponentTypeListWallpaper:
					CopyFile("/mnt/SDCARD/Collections/.media/bglist.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Collections.png"))
					CopyFile("/mnt/SDCARD/Recently Played/.media/bglist.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Recently Played.png"))
					CopyFile("/mnt/SDCARD/Tools/tg5040/.media/bglist.png", filepath.Join(ThemesDirectory, themeName, component.ComponentName, "Tools.png"))
			}
		}
		// Add component directories and types to map while looping
		if homeDirectories[component.ComponentType.ComponentHomeDirectory] == nil {
			homeDirectories[component.ComponentType.ComponentHomeDirectory] = make(map[string]bool)
		}
		homeDirectories[component.ComponentType.ComponentHomeDirectory][component.ComponentType.ComponentType] = true
	}
	logger.Debug("home directories: " + fmt.Sprint(homeDirectories))
	
	// Collect valid parent directories for rom dependent directories
	validParents := make(map[string]bool)
	parentsList, err := getTopLevelRomsDirectories(options.OptionActive)
	if err != nil {
		return err
	}
	for _, parent := range parentsList {
		if options.OptionInactive {
			if parent.DirectoryFileCount == 0 {
				validParents[parent.Filename] = true
			}
		} else {
			validParents[parent.Filename] = true
		}
	}
	logger.Debug("valid parents: " + fmt.Sprint(validParents))

	// Save non meta components
	for homeDirectory, homeDirectoryComponentTypes := range homeDirectories {
		isRomDependent := checkComponentForRomsDependency(homeDirectory)
		logger.Debug("for home directory: " + homeDirectory + ", it is rom dependent: " + fmt.Sprint(isRomDependent))
		saveDecorations(homeDirectory, isRomDependent, validParents, false, homeDirectoryComponentTypes, themeName, homeDirectory)
	}

	return nil
}

func saveDecorations(currentPath string, isRomDependent bool, validRomParents map[string]bool, romParentValidated bool, componentTypes map[string]bool, themeName string, componentHomeDirectory string) {
	logger := common.GetLoggerInstance()

	currentDirectory := filepath.Base(currentPath)
	isMedia := false
	if currentDirectory == ".media" {
		isMedia = true
	}
	logger.Debug("Current path: " + currentPath + " is media: " + fmt.Sprint(isMedia))
	
	files, err := GetFileList(currentPath)
	if err != nil {
		return
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

				logger.Debug("File found: " + itemName + ", is media bg: "+ fmt.Sprint(isMediaBg) + ", is media bglist: " + fmt.Sprint(isMediaBgList) + ", is folder icon: " + fmt.Sprint(isFolderIcon))

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
				logger.Debug("component name prefix: " + componentNamePrefix)
				// Generate starting path list
				decorationPathList := strings.Split(strings.TrimPrefix(currentPath, componentHomeDirectory), string(filepath.Separator))
				decorationPathList = decorationPathList[1:len(decorationPathList) - 1] // remove .media and empty leading element from leading separator
				logger.Debug("decoration path: " + fmt.Sprint(decorationPathList))
				// console name adjustment?
				// Copy file to theme directory with appropriate name
				if isMediaBg && componentTypes[ComponentTypeWallpaper] && len(decorationPathList) > 0 {
					if isRomDependent {
						decorationPathList[0] = FindConsoleTag(decorationPathList[0])
					}
					logger.Debug("adjusted decoration path for wallpaper: " + fmt.Sprint(decorationPathList))
					if decorationPathList[0] != "" {
						destinationName := strings.Join(decorationPathList, folderDelimiter)
						logger.Debug("decoration name: " + destinationName)
						err := CopyFile(filepath.Join(currentPath, itemName), filepath.Join(ThemesDirectory, themeName, componentNamePrefix + "Wallpapers", destinationName + ".png"))
						if err != nil {
							logger.Error("error copying: " + err.Error())
						}
					}
				}
				if isMediaBgList && componentTypes[ComponentTypeListWallpaper] && len(decorationPathList) > 0 {
					if isRomDependent {
						decorationPathList[0] = FindConsoleTag(decorationPathList[0])
					}
					logger.Debug("adjusted decoration path for list wallpaper: " + fmt.Sprint(decorationPathList))
					if decorationPathList[0] != "" {
						destinationName := strings.Join(decorationPathList, folderDelimiter)
						err := CopyFile(filepath.Join(currentPath, itemName), filepath.Join(ThemesDirectory, themeName, componentNamePrefix + "ListWallpapers", destinationName + ".png"))
						if err != nil {
							logger.Error("error copying: " + err.Error())
						}
					}
				}
				if isFolderIcon && componentTypes[ComponentTypeIcon] {
					iconDecorationPathList := append(decorationPathList, strings.TrimSuffix(itemName, itemExt))
					if isRomDependent {
						if romParentValidated {
							iconDecorationPathList[0] = FindConsoleTag(iconDecorationPathList[0])
						} else {
							if validRomParents[iconDecorationPathList[0]] {
								iconDecorationPathList[0] = FindConsoleTag(iconDecorationPathList[0])
							} else {
								iconDecorationPathList[0] = ""
							}
						}
					}
					logger.Debug("adjusted decoration path for icon: " + fmt.Sprint(iconDecorationPathList))
					if iconDecorationPathList[0] != "" {
						destinationName := strings.Join(iconDecorationPathList, folderDelimiter)
						logger.Debug("decoration name: " + destinationName)
						err := CopyFile(filepath.Join(currentPath, itemName), filepath.Join(ThemesDirectory, themeName, componentNamePrefix + "Icons", destinationName + ".png"))
						if err != nil {
							logger.Error("error copying: " + err.Error())
						}
					}
				}
			}
		} else {
			// recurse
			portsFolder := (FindConsoleTag(currentPath) == "(PORTS)" && componentHomeDirectory == "/mnt/SDCARD/Roms" && !(filepath.Dir(currentPath) == componentHomeDirectory || filepath.Dir(filepath.Dir(currentPath)) == componentHomeDirectory))
			toolsFolder := (componentHomeDirectory == ToolsDirectory && !(currentPath == componentHomeDirectory || filepath.Dir(currentPath) == componentHomeDirectory))
			if file.IsDir() && !portsFolder && !toolsFolder {
				logger.Debug("directory found, trying to recurse: " + itemName)
				if !isRomDependent || romParentValidated || itemName == ".media" || validRomParents[itemName] {
					logger.Debug("valid directory. recursing")
					saveDecorations(filepath.Join(currentPath, itemName), isRomDependent, validRomParents, true, componentTypes, themeName, componentHomeDirectory)
				}
			}
		}
	}
}

func resetToDefaultRequestedComponents(components []models.Component, options models.ComponentOptionSelections) error {
	// Reset meta components and build component directory/type maps for recursion
	homeDirectories := make(map[string]map[string]bool)
	for _, component := range components {
		if component.ComponentType.ContainsMetaFiles && !options.OptionInactive {
			switch component.ComponentType.ComponentType {
				case ComponentTypeIcon:
					common.DeleteFile("/mnt/SDCARD/.media/Collections.png")
					common.DeleteFile("/mnt/SDCARD/.media/Recently Played.png")
					common.DeleteFile("/mnt/SDCARD/Tools/.media/tg5040.png")
				case ComponentTypeWallpaper:
					common.DeleteFile("/mnt/SDCARD/Collections/.media/bg.png")
					common.DeleteFile("/mnt/SDCARD/Recently Played/.media/bg.png")
					common.DeleteFile("/mnt/SDCARD/Tools/tg5040/.media/bg.png")
					common.DeleteFile("/mnt/SDCARD/bg.png")
				case ComponentTypeListWallpaper:
					common.DeleteFile("/mnt/SDCARD/Collections/.media/bglist.png")
					common.DeleteFile("/mnt/SDCARD/Recently Played/.media/bglist.png")
					common.DeleteFile("/mnt/SDCARD/Tools/tg5040/.media/bglist.png")
			}
		}
		// Add component directories and types to map while looping
		if homeDirectories[component.ComponentType.ComponentHomeDirectory] == nil {
			homeDirectories[component.ComponentType.ComponentHomeDirectory] = make(map[string]bool)
		}
		homeDirectories[component.ComponentType.ComponentHomeDirectory][component.ComponentType.ComponentType] = true
	}
	
	// Collect valid parent directories for rom dependent directories
	validParents := make(map[string]bool)
	parentsList, err := getTopLevelRomsDirectories(options.OptionActive)
	if err != nil {
		return err
	}
	for _, parent := range parentsList {
		if options.OptionInactive {
			if parent.DirectoryFileCount == 0 {
				validParents[parent.Filename] = true
			}
		} else {
			validParents[parent.Filename] = true
		}
	}

	// Reset non meta components
	for homeDirectory, homeDirectoryComponentTypes := range homeDirectories {
		isRomDependent := checkComponentForRomsDependency(homeDirectory)
		if !options.OptionInactive || isRomDependent {
			resetDecorations(homeDirectory, isRomDependent, validParents, false, homeDirectoryComponentTypes, homeDirectory)
		}
	}

	return nil
}

func resetDecorations(currentPath string, isRomDependent bool, validRomParents map[string]bool, romParentValidated bool, componentTypes map[string]bool, componentHomeDirectory string) {
	currentDirectory := filepath.Base(currentPath)
	isMedia := false
	if currentDirectory == ".media" {
		isMedia = true
	}
	
	files, err := GetFileList(currentPath)
	if err != nil {
		return
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
					common.DeleteFile(filepath.Join(currentPath, itemName))
				}
				if isMediaBgList && componentTypes[ComponentTypeListWallpaper] {
					common.DeleteFile(filepath.Join(currentPath, itemName))
				}
				if isFolderIcon && componentTypes[ComponentTypeIcon] {
					common.DeleteFile(filepath.Join(currentPath, itemName))
				}
			}
		} else {
			// recurse
			portsFolder := (FindConsoleTag(currentPath) == "(PORTS)" && componentHomeDirectory == "/mnt/SDCARD/Roms" && !(filepath.Dir(currentPath) == componentHomeDirectory || filepath.Dir(filepath.Dir(currentPath)) == componentHomeDirectory))
			toolsFolder := (componentHomeDirectory == ToolsDirectory && !(currentPath == componentHomeDirectory || filepath.Dir(currentPath) == componentHomeDirectory))
			if file.IsDir() && !portsFolder && !toolsFolder {
				if !isRomDependent || romParentValidated || itemName == ".media" || validRomParents[itemName] {
					resetDecorations(filepath.Join(currentPath, itemName), isRomDependent, validRomParents, true, componentTypes, componentHomeDirectory)
				}
			}
		}
	}
}

func applySelectedThemeComponents(theme models.Theme, components []models.Component, options models.ComponentOptionSelections) error {
	logger := common.GetLoggerInstance()
	
	// Collect valid parent directories for non-meta components
	validParents := make(map[string]string)
	parentsList, err := getTopLevelRomsDirectories(options.OptionActive)
	if err != nil {
		return err
	}
	for _, parent := range parentsList {
		if options.OptionInactive {
			if parent.DirectoryFileCount == 0 {
				validParents[FindConsoleTag(parent.Filename)] = parent.Filename
			}
		} else {
			validParents[FindConsoleTag(parent.Filename)] = parent.Filename
		}
	}

	// For each component, 
	for _, component := range components {
		isRomDependent := checkComponentForRomsDependency(component.ComponentType.ComponentHomeDirectory)
		romParentSet := make(map[string]bool)
		for _, componentPath := range component.ComponentPaths {
			// For each path in a component, apply all decorations in the path according to the rules
			files, err := GetFileList(componentPath)
			if err != nil {
				logger.Error("Error collecting items from " + componentPath)
			} else {
				for _, file := range files {
					itemName := file.Name()
					itemExt := filepath.Ext(itemName)
					if filepath.Ext(itemName) == ".png" {
						metaFileCopied := false
						if component.ComponentType.ContainsMetaFiles {
							switch component.ComponentType.ComponentType {
								case ComponentTypeIcon:
									switch itemName {
										case "Collections.png":
											applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/.media/Collections.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
										case "Recently Played.png":
											applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/.media/Recently Played.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
										case "Tools.png":
											applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/Tools/.media/tg5040.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
									}
								case ComponentTypeWallpaper:
									switch itemName {
										case "Collections.png":
											applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/Collections/.media/bg.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
										case "Recently Played.png":
											applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/Recently Played/.media/bg.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
										case "Tools.png":
											applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/Tools/tg5040/.media/bg.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
										case "Root.png":
											applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/bg.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
									}
								case ComponentTypeListWallpaper:
									switch itemName {
										case "Collections.png":
											applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/Collections/.media/bglist.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
										case "Recently Played.png":
											applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/Recently Played/.media/bglist.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
										case "Tools.png":
											applyThemeDecorationSafely(filepath.Join(componentPath, file.Name()), "/mnt/SDCARD/Tools/tg5040/.media/bglist.png", options.OptionPreserve, options.OptionConfirm)
											metaFileCopied = true
									}
							}
						}
						if !metaFileCopied {
							// File is not a meta file, move if possible
							itemBase := strings.TrimSuffix(itemName, itemExt)
							filePathParts := strings.Split(itemBase, folderDelimiter)
							tryToPlace := true
							if isRomDependent {
								// Rom dependent file. Replace the first piece of the name with the user's console directory. 
								// If the item is for that console directtory and is not the first item for that console directory, then skip.
								consoleTag := FindConsoleTag(filePathParts[0])
								parentConsoleName := validParents[consoleTag]
								if parentConsoleName == "" {
									tryToPlace = false
								} else {
									if len(filePathParts) == 1 {
										if romParentSet[consoleTag] {
											tryToPlace = false
										} else {
											romParentSet[consoleTag] = true
										}
									}
									filePathParts[0] = parentConsoleName
								}
							}
							if tryToPlace {
								// file has a valid parent. Move if the parent directory exists
								filePathParts = append([]string{component.ComponentType.ComponentHomeDirectory}, filePathParts...)
								parentConsoleDirectory := filepath.Join(filePathParts...)
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
											parentPath := filepath.Join(filePathParts[:len(filePathParts) - 1]...)
											destinationPath = GetTrueIconPath(parentPath, parentConsoleDirectory)
										case ComponentTypeWallpaper:
											destinationPath = GetTrueWallpaperPath(parentConsoleDirectory)
										case ComponentTypeListWallpaper:
											destinationPath = GetTrueListWallpaperPath(parentConsoleDirectory)
									}
									if destinationPath != "" {
										applyThemeDecorationSafely(filepath.Join(componentPath, itemName), destinationPath, options.OptionPreserve, options.OptionConfirm)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func applyThemeDecorationSafely(sourcePath string, destinationPath string, existencePreCheck bool, confirmCopy bool) {
	if confirmCopy {
		sourcePathList := strings.Split(sourcePath, string(filepath.Separator))
		message := "Apply " + filepath.Join(sourcePathList[len(sourcePathList) - 2:]...)
		if ConfirmActionCustomBack(message, sourcePath, "Skip") {
			CopyFile(sourcePath, destinationPath)
		}
		return
	}

	if existencePreCheck {
		if DoesFileExists(destinationPath) {
			return
		}
	}
	CopyFile(sourcePath, destinationPath)
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