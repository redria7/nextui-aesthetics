package utils

import (
	"path/filepath"
	"fmt"
	"io"
	"net/http"
	"strings"
	"encoding/json"
	"unicode"
	"sort"
	"nextui-aesthetics/models"
	"os"
	"time"
	"errors"
	"archive/zip"
	gaba "github.com/redria7/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
)

const (
	ThemeLibrary = "/mnt/SDCARD/.userdata/shared/Aesthetics/Themes"
	previewStandardName = "preview.png"
	previewHiddenName = "hidden.preview.png"
	catalogURL = "https://raw.githubusercontent.com/Leviathanium/NextUI-Themes/main/Catalog/catalog.json"
	previewURLPrefix = "https://raw.githubusercontent.com/Leviathanium/NextUI-Themes/main/"
)

func extractThemeZip(zipPath, destDir string) error {
	logger := common.GetLoggerInstance()
	logger.Debug("Extracting ZIP file " + zipPath + " to " + destDir)

	// Open the ZIP file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("error opening ZIP file: %w", err)
	}
	defer reader.Close()

	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("error creating destination directory: %w", err)
	}

	// Analyze the ZIP structure to detect common root directories
	// This helps prevent issues like Theme/Theme nesting
	rootDirs := make(map[string]int)
	totalFiles := 0

	for _, file := range reader.File {
		// Skip __MACOSX directories and hidden files
		if strings.Contains(file.Name, "__MACOSX") || strings.HasPrefix(filepath.Base(file.Name), ".") {
			continue
		}

		totalFiles++

		// Get the top-level directory from the path
		pathParts := strings.Split(file.Name, "/")
		if len(pathParts) > 1 && pathParts[0] != "" {
			rootDirs[pathParts[0]]++
		}
	}

	// Check if all files are in a single root directory
	var commonRoot string
	destBaseName := filepath.Base(destDir)

	// Find if there's a single common root directory that contains all files
	for dir, count := range rootDirs {
		if count == totalFiles || (float64(count)/float64(totalFiles) > 0.9) {
			commonRoot = dir
			break
		}
	}

	logger.Debug(fmt.Sprintf("ZIP analysis - Total files: %d, Common root: %s, Dest dir: %s",
		totalFiles, commonRoot, destBaseName))

	// Extract each file in the ZIP archive
	for _, file := range reader.File {
		// Skip __MACOSX directories and hidden files
		if strings.Contains(file.Name, "__MACOSX") || strings.HasPrefix(filepath.Base(file.Name), ".") {
			continue
		}

		// Determine the target path for extraction
		var targetPath string

		// If there's a common root that matches the destination directory name or ends with the same extension,
		// strip it to avoid Theme/Theme nesting
		if commonRoot != "" && (commonRoot == destBaseName ||
			(strings.HasSuffix(commonRoot, filepath.Ext(destBaseName)) &&
				strings.HasSuffix(destBaseName, filepath.Ext(destBaseName)))) {

			if strings.HasPrefix(file.Name, commonRoot+"/") {
				// Strip the common root to avoid nesting
				relativePath := strings.TrimPrefix(file.Name, commonRoot+"/")

				// Special case: if we have an entry for just the directory itself (resulting in empty path)
				// Skip this entry as we've already created the destination directory
				if relativePath == "" {
					logger.Debug(fmt.Sprintf("Skipping root directory entry: %s", file.Name))
					continue
				}

				targetPath = filepath.Join(destDir, relativePath)
				logger.Debug(fmt.Sprintf("Stripping common root from: %s to: %s", file.Name, relativePath))
			} else {
				// Normal file, not in common root
				targetPath = filepath.Join(destDir, file.Name)
				logger.Debug(fmt.Sprintf("File doesn't have common root prefix: %s", file.Name))
			}
		} else {
			// No common root or it doesn't match destination - extract normally
			targetPath = filepath.Join(destDir, file.Name)
			logger.Debug(fmt.Sprintf("Normal extraction for: %s", file.Name))
		}

		// Check for directory traversal attacks - only for non-empty paths
		// The destDir path itself is always safe since we create it explicitly
		cleanPath := filepath.Clean(targetPath)
		if cleanPath != destDir && !strings.HasPrefix(cleanPath, destDir+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", file.Name)
		}

		// Handle directories
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("error creating directory %s: %w", targetPath, err)
			}
			continue
		}

		// Create the directory structure for the file
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("error creating directory structure: %w", err)
		}

		// Open the file in the ZIP
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("error opening file in ZIP: %w", err)
		}

		// Create the destination file
		outFile, err := os.Create(targetPath)
		if err != nil {
			rc.Close()
			return fmt.Errorf("error creating file %s: %w", targetPath, err)
		}

		// Copy the content
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return fmt.Errorf("error extracting file %s: %w", targetPath, err)
		}
	}

	logger.Debug(fmt.Sprintf("ZIP extraction completed successfully"))
	return nil
}

func DownloadTheme(theme models.ThemeSummary) error {
	// Download theme zip
	tmp, completed, err := downloadThemeZip(theme)
	if err != nil {
		return err
	} else if !completed {
		return errors.New("Download incomplete")
	}
	defer os.Remove(tmp)

	// Ensure Themes directory exists
	themePath := filepath.Join(ThemeLibrary, theme.ThemeName)
	EnsureDirectoryExists(themePath)

	// Extract the ZIP file
	_, err = gaba.ProcessMessage("Unzipping " + theme.ThemeName, gaba.ProcessMessageOptions{}, func() (interface{}, error) {
		err = extractThemeZip(tmp, themePath)
		if err != nil {
			return nil, err
		}
		time.Sleep(1 * time.Second)
		return nil, nil
	})
	return err
}

func downloadThemeZip(theme models.ThemeSummary) (tempFile string, completed bool, error error) {
	tmp := filepath.Join("/tmp", theme.ThemeName)

	res, err := gaba.DownloadManager([]gaba.Download{{
		URL:         theme.ZipPath,
		Location:    tmp,
		DisplayName: "Downloading " + theme.ThemeName,
	}}, make(map[string]string), true)

	if err == nil && len(res.Errors) > 0 {
		err = res.Errors[0]
	}

	if err != nil {
		return "", false, err
	} else if res.Cancelled {
		return "", false, nil
	}

	return tmp, true, nil
}

func DownloadThemePreviews(themes []models.ThemeSummary) {
	var downloads []gaba.Download
	for _, theme := range themes {
		downloads = append(downloads, gaba.Download{
			URL: previewURLPrefix + theme.PreviewPath,
			Location: filepath.Join(ThemeLibrary, theme.ThemeName, previewStandardName),
			DisplayName: "Downloading " + theme.ThemeName + " Preview",
		})
	}

	gaba.DownloadManager(downloads, make(map[string]string), true)
}

func fetch(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func GenerateThemeCatalog() []models.ThemeSummary {
	themeCatalogSummary := []models.ThemeSummary{}
	
	// First download catalog.json
	data, err := fetch(catalogURL)
	if err != nil {
		return themeCatalogSummary
	}

	// Then convert to usable data
	var catalogData models.CatalogData
	if err := json.Unmarshal(data, &catalogData); err != nil {
		return themeCatalogSummary
	}
	for themeName, itemInfo := range catalogData.Themes {
		if itemInfo.PreviewPath != "" {
			themeCatalogSummary = append(themeCatalogSummary, models.ThemeSummary{
				ThemeName: 		themeName,
				PreviewPath: 	itemInfo.PreviewPath,
				Author: 		itemInfo.Author,
				Description: 	itemInfo.Description,
				ZipPath: 		itemInfo.URL,
				LastUpdated: 	itemInfo.LastUpdated,
				ThemeType: 		"Theme",
			})
		}
		
	}
	for themeType, nameMap := range catalogData.Components {
		themeTypePretty := themeTypePrettify(themeType)
		for themeName, itemInfo := range nameMap {
			if itemInfo.PreviewPath != "" {
				themeCatalogSummary = append(themeCatalogSummary, models.ThemeSummary{
					ThemeName: 		themeName,
					PreviewPath: 	itemInfo.PreviewPath,
					Author: 		itemInfo.Author,
					Description: 	itemInfo.Description,
					ZipPath: 		itemInfo.URL,
					LastUpdated: 	itemInfo.LastUpdated,
					ThemeType: 		themeTypePretty,
				})
			}
		}
	}

	// Finally, sort the result for proper display and return
	sort.Slice(themeCatalogSummary, func(i, j int) bool {
		return themeCatalogSummary[i].ThemeName < themeCatalogSummary[j].ThemeName
	})
	return themeCatalogSummary
}

func themeTypePrettify(themeType string) string {
	if themeType == "" {
		return themeType
	}
	themeType = strings.TrimSuffix(themeType, "s")
	runes := []rune(themeType)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func SwapHiddenState(theme models.ThemeSummary) {
	themePath := filepath.Join(ThemeLibrary, theme.ThemeName)
	normalPath := filepath.Join(themePath, previewStandardName)
	hiddenPath := filepath.Join(themePath, previewHiddenName)
	if DoesFileExists(hiddenPath) {
		os.Rename(hiddenPath, normalPath)
	} else if DoesFileExists(normalPath) {
		os.Rename(normalPath, hiddenPath)
	}
}

func GetDownloadedThemes() map[string]models.Theme {
	themeMap := make(map[string]models.Theme)
	
	themes, err := GetFileList(ThemeLibrary)
	if err != nil {
		return themeMap
	}

	for _, theme := range themes {
		themeName := theme.Name()
		themePath := filepath.Join(ThemeLibrary, theme.Name())
		themePreview := DoesFileExists(filepath.Join(themePath, previewStandardName))
		themeHiddenPreview := DoesFileExists(filepath.Join(themePath, previewHiddenName))
		baseThemeCount := 0
		if themePreview {
			baseThemeCount++
		}
		if themeHiddenPreview {
			baseThemeCount++
			themePreview = true
		}
		themeDownloaded := false
		themeContents, err := GetFileList(themePath)
		if err == nil && len(themeContents) > baseThemeCount {
			themeDownloaded = true
		}
		themeMap[themeName] = models.Theme{
			ThemeName: themeName,
			ThemePath: themePath,
			PreviewFound: themePreview,
			ContainsTheme: themeDownloaded,
			IsHidden: themeHiddenPreview,
		}
	}

	return themeMap
}

func GetPreviewPath(themeName string) string {
	themePath := filepath.Join(ThemeLibrary, themeName)
	normalPath := filepath.Join(themePath, previewStandardName)
	if DoesFileExists(normalPath) {
		return normalPath
	}
	return filepath.Join(themePath, previewHiddenName)
}