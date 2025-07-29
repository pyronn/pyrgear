package comands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"github.com/spf13/cobra"
)

var (
	exifImagePath    string
	exifOutputFormat string
	exifRecursive    bool
)

// ExifCmd represents the exif command
var ExifCmd = &cobra.Command{
	Use:   "exif",
	Short: "Read EXIF information from image files",
	Long: `Read and display all EXIF information from image files.
	
Examples:
  # Read EXIF from a single image
  pyrgear exif --image /path/to/image.jpg
  
  # Read EXIF from all images in a directory
  pyrgear exif --dir /path/to/images
  
  # Read EXIF recursively from all subdirectories
  pyrgear exif --dir /path/to/images --recursive
  
  # Output in JSON format
  pyrgear exif --image /path/to/image.jpg --format json
  
Supported image formats: JPEG, TIFF`,
	Run: func(cmd *cobra.Command, args []string) {
		if exifImagePath == "" && directory == "" {
			fmt.Println("Error: either --image or --dir is required")
			cmd.Help()
			return
		}

		if exifImagePath != "" {
			// Process single image
			err := processImageExif(exifImagePath, exifOutputFormat)
			if err != nil {
				fmt.Printf("Error processing image: %v\n", err)
			}
		} else {
			// Process directory
			err := processDirectoryExif(directory, exifOutputFormat, exifRecursive)
			if err != nil {
				fmt.Printf("Error processing directory: %v\n", err)
			}
		}
	},
}

func init() {
	ExifCmd.Flags().StringVar(&exifImagePath, "image", "", "Path to a single image file")
	ExifCmd.Flags().StringVar(&directory, "dir", "", "Directory containing image files")
	ExifCmd.Flags().StringVar(&exifOutputFormat, "format", "text", "Output format: text or json")
	ExifCmd.Flags().BoolVar(&exifRecursive, "recursive", false, "Process subdirectories recursively")
}

// processImageExif processes a single image file and extracts EXIF data
func processImageExif(imagePath string, format string) error {
	// Check if file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return fmt.Errorf("image file does not exist: %s", imagePath)
	}

	// Check if it's a supported image format
	ext := strings.ToLower(filepath.Ext(imagePath))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".tiff" && ext != ".tif" {
		return fmt.Errorf("unsupported image format: %s (supported: jpg, jpeg, tiff, tif)", ext)
	}

	// Open the image file
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image file: %v", err)
	}
	defer file.Close()

	// Decode EXIF data
	exifData, err := exif.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode EXIF data: %v", err)
	}

	// Display EXIF information
	fmt.Printf("\n=== EXIF Information for %s ===\n", imagePath)

	if format == "json" {
		return displayExifAsJSON(exifData)
	} else {
		return displayExifAsText(exifData)
	}
}

// processDirectoryExif processes all images in a directory
func processDirectoryExif(dirPath string, format string, recursive bool) error {
	// Check if directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("failed to access directory %s: %v", dirPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dirPath)
	}

	return filepath.Walk(
		dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("Warning: Error accessing %s: %v\n", path, err)
				return nil
			}

			// Skip directories
			if info.IsDir() {
				if !recursive && path != dirPath {
					return filepath.SkipDir
				}
				return nil
			}

			// Check if it's a supported image format
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".tiff" || ext == ".tif" {
				err := processImageExif(path, format)
				if err != nil {
					fmt.Printf("Warning: Failed to process %s: %v\n", path, err)
				}
			}

			return nil
		},
	)
}

// textWalker implements the Walker interface for text output
type textWalker struct{}

func (w textWalker) Walk(name exif.FieldName, tag *tiff.Tag) error {
	// Get the tag value as a string
	val, err := tag.StringVal()
	if err != nil {
		val = fmt.Sprintf("(error: %v)", err)
	}

	// Display the tag name and value
	fmt.Printf("%-30s: %s\n", string(name), val)
	return nil
}

// displayExifAsText displays EXIF data in human-readable text format
func displayExifAsText(exifData *exif.Exif) error {
	// Walk through all EXIF tags
	walker := textWalker{}
	err := exifData.Walk(walker)
	if err != nil {
		return err
	}

	// Try to get some common GPS coordinates if available
	lat, lon, err := exifData.LatLong()
	if err == nil {
		fmt.Printf("%-30s: %f, %f\n", "GPS Coordinates", lat, lon)
	}

	fmt.Println()
	return nil
}

// jsonWalker implements the Walker interface for JSON output
type jsonWalker struct {
	first bool
}

func (w *jsonWalker) Walk(name exif.FieldName, tag *tiff.Tag) error {
	if !w.first {
		fmt.Print(",")
	}
	fmt.Print("\n")

	// Get the tag value as a string
	val, err := tag.StringVal()
	if err != nil {
		val = fmt.Sprintf("error: %v", err)
	}

	// Escape quotes in the value
	val = strings.ReplaceAll(val, "\"", "\\\"")

	fmt.Printf("  \"%s\": \"%s\"", string(name), val)
	w.first = false
	return nil
}

// displayExifAsJSON displays EXIF data in JSON format
func displayExifAsJSON(exifData *exif.Exif) error {
	fmt.Println("{")

	walker := &jsonWalker{first: true}
	err := exifData.Walk(walker)
	if err != nil {
		return err
	}

	// Try to get GPS coordinates if available
	lat, lon, err := exifData.LatLong()
	if err == nil {
		if !walker.first {
			fmt.Print(",")
		}
		fmt.Printf("\n  \"GPS_Latitude\": %f,", lat)
		fmt.Printf("\n  \"GPS_Longitude\": %f", lon)
	}

	fmt.Println("\n}")
	fmt.Println()
	return nil
}
