package comands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	pattern     string
	replacement string
	recursive   bool
	dryRun      bool
	directory   string
	ruleType    string
	sourcePath  string
	outputDir   string
	// preName 指定复制文件前缀名,用于rule wx-exporter
	preName string
	// foldername-rename rule params
	parentDir string
)

// renameCmd represents the rename command
var RenameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Batch rename files in a directory",
	Long: `Batch rename files in a specified directory based on a pattern.
	
Example:
  pyrgear rename --dir ./my_files --pattern "file_(\d+)" --replacement "document_$1" --recursive
  pyrgear rename --dir ./my_files --rule "timestamp"
  pyrgear rename --rule "wx-exporter" --source-path "/path/to/source" --output-dir "./output"
  pyrgear rename --rule "wx-exporter" --source-path "/path/to/source" --output-dir "./output" --pre-name "my_prefix"
  
This will rename all files matching the pattern "file_(\d+)" to "document_$1" in the ./my_files directory.
If --recursive is specified, it will also process files in subdirectories.
If --rule is specified, it will use a predefined renaming rule instead of pattern/replacement.
For wx-exporter rule, it will extract images from path2/assets/ folders in the specified source directory (path1)
and copy them to the output directory with names like "path2_001". `,
	Run: func(cmd *cobra.Command, args []string) {
		// Special handling for wx-exporter rule
		if strings.ToLower(ruleType) == "wx-exporter" {
			err := processWxExporter(sourcePath, outputDir, dryRun)
			if err != nil {
				fmt.Printf("Error processing wx-exporter: %v\n", err)
			}
			return
		}

		// Special handling for foldername-rename rule
		if strings.ToLower(ruleType) == "foldername-rename" {
			if (directory == "" && parentDir == "") || (directory != "" && parentDir != "") {
				fmt.Println("Error: You must specify either --dir or --pdir, but not both, for foldername-rename rule.")
				return
			}
			if directory != "" {
				err := processFoldernameRename(directory, dryRun)
				if err != nil {
					fmt.Printf("Error processing foldername-rename: %v\n", err)
				}
				return
			}
			if parentDir != "" {
				entries, err := os.ReadDir(parentDir)
				if err != nil {
					fmt.Printf("Error reading parent directory: %v\n", err)
					return
				}
				for _, entry := range entries {
					if entry.IsDir() {
						dirPath := filepath.Join(parentDir, entry.Name())
						err := processFoldernameRename(dirPath, dryRun)
						if err != nil {
							fmt.Printf("Error processing %s: %v\n", dirPath, err)
						}
					}
				}
				return
			}
		}

		if directory == "" {
			fmt.Println("Error: directory is required for this operation")
			cmd.Help()
			return
		}

		// If a rule is specified, use that instead of pattern/replacement
		if ruleType != "" {
			err := processDirectoryWithRule(directory, ruleType, recursive, dryRun)
			if err != nil {
				fmt.Printf("Error processing directory with rule: %v\n", err)
			}
			return
		}

		// Otherwise use the regular pattern/replacement logic
		if pattern == "" {
			fmt.Println("Error: either pattern or rule is required")
			cmd.Help()
			return
		}

		// Compile the regular expression
		re, err := regexp.Compile(pattern)
		if err != nil {
			fmt.Printf("Error compiling regular expression: %v\n", err)
			return
		}

		// Process the directory
		err = processDirectory(directory, re, replacement, recursive, dryRun)
		if err != nil {
			fmt.Printf("Error processing directory: %v\n", err)
		}
	},
}

func init() {
	RenameCmd.Flags().StringVar(&directory, "dir", "", "Directory to process (required for most operations)")
	RenameCmd.Flags().StringVar(&pattern, "pattern", "", "Regular expression pattern to match filenames")
	RenameCmd.Flags().StringVar(&replacement, "replacement", "", "Replacement pattern for new filenames")
	RenameCmd.Flags().BoolVar(&recursive, "recursive", false, "Process subdirectories recursively")
	RenameCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be renamed without actually renaming")
	RenameCmd.Flags().StringVar(
		&ruleType, "rule", "",
		"Predefined rule for renaming (e.g., 'timestamp', 'sequence', 'lowercase', 'wx-exporter')",
	)
	RenameCmd.Flags().StringVar(
		&sourcePath, "source-path", "", "Source path for wx-exporter rule (optional, defaults to current directory)",
	)
	RenameCmd.Flags().StringVar(&outputDir, "output-dir", "wx-export", "Output directory for wx-exporter rule")
	RenameCmd.Flags().StringVar(
		&preName, "pre-name", "", "Predefined name for wx-exporter rule exporter file optional,defaults to source-path",
	)
	RenameCmd.Flags().StringVar(&parentDir, "pdir", "", "Parent directory for foldername-rename rule (batch mode)")
}

// processWxExporter processes the wx-exporter rule
func processWxExporter(sourcePath string, outputDir string, dryRun bool) error {
	// If sourcePath is not specified, use current directory
	if sourcePath == "" {
		var err error
		sourcePath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %v", err)
		}
	}

	// Create output directory if it doesn't exist
	if !dryRun {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create output directory %s: %v", outputDir, err)
		}
	}

	// Map to track sequence numbers for each path2
	sequenceMap := make(map[string]int)

	// First, find all subdirectories (path2) in the source directory (path1)
	path2Dirs, err := findPath2Directories(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to find subdirectories in %s: %v", sourcePath, err)
	}

	if len(path2Dirs) == 0 {
		fmt.Printf("Warning: No subdirectories found in %s\n", sourcePath)
	}

	sourceName := filepath.Base(sourcePath)
	// If preName is specified, use it as the prefix
	if preName != "" {
		sourceName = preName
	}

	// Process each path2 directory
	for _, path2Dir := range path2Dirs {
		// Get the path2 name (just the directory name, not the full path)
		path2Name := filepath.Base(path2Dir)

		// Check if assets directory exists
		assetsDir := filepath.Join(path2Dir, "assets")
		assetsInfo, err := os.Stat(assetsDir)
		if err != nil || !assetsInfo.IsDir() {
			// Skip if assets directory doesn't exist
			continue
		}

		// Process all files in the assets directory
		assetFiles, err := os.ReadDir(assetsDir)
		if err != nil {
			fmt.Printf("Warning: Failed to read assets directory %s: %v\n", assetsDir, err)
			continue
		}

		// Process each file in the assets directory
		for _, file := range assetFiles {
			if file.IsDir() {
				// Skip subdirectories in assets
				continue
			}

			filePath := filepath.Join(assetsDir, file.Name())

			// Check if the file is an image (simple check by extension)
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
				continue
			}

			// Increment sequence number for this path2
			sequenceMap[path2Name]++
			sequence := sequenceMap[path2Name]

			// Create new filename: path2_sequence with original extension
			newName := fmt.Sprintf("%s_%s_%03d%s", sourceName, path2Name, sequence, ext)
			newPath := filepath.Join(outputDir, newName)

			if dryRun {
				fmt.Printf("Would copy: %s -> %s\n", filePath, newPath)
			} else {
				fmt.Printf("Copying: %s -> %s\n", filePath, newPath)
				err := copyFile(filePath, newPath)
				if err != nil {
					fmt.Printf("Error copying %s: %v\n", filePath, err)
				}
			}
		}
	}

	return nil
}

// findPath2Directories finds all immediate subdirectories in the given path1 directory
func findPath2Directories(path1 string) ([]string, error) {
	entries, err := os.ReadDir(path1)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(path1, entry.Name()))
		}
	}

	return dirs, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

// processDirectoryWithRule processes files in the given directory using a predefined rule
func processDirectoryWithRule(dir string, rule string, recursive bool, dryRun bool) error {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("failed to access directory %s: %v", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	// Read directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %v", dir, err)
	}

	// Process each entry based on the rule
	switch strings.ToLower(rule) {
	case "timestamp":
		// Add timestamp prefix to each file
		for _, entry := range entries {
			if entry.IsDir() {
				if recursive {
					if err := processDirectoryWithRule(
						filepath.Join(dir, entry.Name()), rule, recursive, dryRun,
					); err != nil {
						fmt.Printf("Warning: %v\n", err)
					}
				}
				continue
			}

			// Get file info to access modification time
			fileInfo, err := entry.Info()
			if err != nil {
				fmt.Printf("Error getting file info for %s: %v\n", entry.Name(), err)
				continue
			}

			// Format timestamp as YYYYMMDD_HHMMSS
			timestamp := fileInfo.ModTime().Format("20060102_150405")
			newName := fmt.Sprintf("%s_%s", timestamp, entry.Name())
			oldPath := filepath.Join(dir, entry.Name())
			newPath := filepath.Join(dir, newName)

			if dryRun {
				fmt.Printf("Would rename: %s -> %s\n", oldPath, newPath)
			} else {
				fmt.Printf("Renaming: %s -> %s\n", oldPath, newPath)
				if err := os.Rename(oldPath, newPath); err != nil {
					fmt.Printf("Error renaming %s: %v\n", oldPath, err)
				}
			}
		}

	case "sequence":
		// Rename files with sequential numbers
		for i, entry := range entries {
			if entry.IsDir() {
				if recursive {
					if err := processDirectoryWithRule(
						filepath.Join(dir, entry.Name()), rule, recursive, dryRun,
					); err != nil {
						fmt.Printf("Warning: %v\n", err)
					}
				}
				continue
			}

			// Get file extension
			ext := filepath.Ext(entry.Name())
			// Create new name with sequence number
			newName := fmt.Sprintf("file_%03d%s", i+1, ext)
			oldPath := filepath.Join(dir, entry.Name())
			newPath := filepath.Join(dir, newName)

			if dryRun {
				fmt.Printf("Would rename: %s -> %s\n", oldPath, newPath)
			} else {
				fmt.Printf("Renaming: %s -> %s\n", oldPath, newPath)
				if err := os.Rename(oldPath, newPath); err != nil {
					fmt.Printf("Error renaming %s: %v\n", oldPath, err)
				}
			}
		}

	case "lowercase":
		// Convert all filenames to lowercase
		for _, entry := range entries {
			if entry.IsDir() {
				if recursive {
					if err := processDirectoryWithRule(
						filepath.Join(dir, entry.Name()), rule, recursive, dryRun,
					); err != nil {
						fmt.Printf("Warning: %v\n", err)
					}
				}
				continue
			}

			// Convert name to lowercase
			newName := strings.ToLower(entry.Name())
			if newName == entry.Name() {
				// Skip if name is already lowercase
				continue
			}

			oldPath := filepath.Join(dir, entry.Name())
			newPath := filepath.Join(dir, newName)

			if dryRun {
				fmt.Printf("Would rename: %s -> %s\n", oldPath, newPath)
			} else {
				fmt.Printf("Renaming: %s -> %s\n", oldPath, newPath)
				if err := os.Rename(oldPath, newPath); err != nil {
					fmt.Printf("Error renaming %s: %v\n", oldPath, err)
				}
			}
		}

	default:
		return fmt.Errorf("unknown rule type: %s", rule)
	}

	return nil
}

// processDirectory processes files in the given directory
func processDirectory(dir string, re *regexp.Regexp, repl string, recursive bool, dryRun bool) error {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("failed to access directory %s: %v", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	// Read directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %v", dir, err)
	}

	// Process each entry
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			if recursive {
				if err := processDirectory(path, re, repl, recursive, dryRun); err != nil {
					fmt.Printf("Warning: %v\n", err)
				}
			}
			continue
		}

		// Process file
		if re.MatchString(entry.Name()) {
			newName := re.ReplaceAllString(entry.Name(), repl)
			newPath := filepath.Join(dir, newName)

			if dryRun {
				fmt.Printf("Would rename: %s -> %s\n", path, newPath)
			} else {
				fmt.Printf("Renaming: %s -> %s\n", path, newPath)
				if err := os.Rename(path, newPath); err != nil {
					fmt.Printf("Error renaming %s: %v\n", path, err)
				}
			}
		}
	}

	return nil
}

// processFoldernameRename renames all files in a directory to foldername_序号.扩展名
func processFoldernameRename(targetDir string, dryRun bool) error {
	info, err := os.Stat(targetDir)
	if err != nil {
		return fmt.Errorf("failed to access directory %s: %v", targetDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", targetDir)
	}
	folderName := filepath.Base(targetDir)
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %v", targetDir, err)
	}
	seq := 1
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		oldPath := filepath.Join(targetDir, entry.Name())
		ext := filepath.Ext(entry.Name())
		newName := fmt.Sprintf("%s_%03d%s", folderName, seq, ext)
		newPath := filepath.Join(targetDir, newName)
		if dryRun {
			fmt.Printf("Would rename: %s -> %s\n", oldPath, newPath)
		} else {
			fmt.Printf("Renaming: %s -> %s\n", oldPath, newPath)
			err := os.Rename(oldPath, newPath)
			if err != nil {
				fmt.Printf("Error renaming %s: %v\n", oldPath, err)
			}
		}
		seq++
	}
	return nil
}

func getDirectoryLevels(path string) []string {
	// 使用 filepath.Split 分割路径
	var parts []string
	for {
		dir, file := filepath.Split(path)
		if dir == "" && file == "" {
			break
		}
		if file != "" {
			parts = append([]string{file}, parts...)
		}
		path = filepath.Clean(dir)
		if path == "." || path == "/" || path == "\\" {
			break
		}
	}
	return parts
}
