package util

import (
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// upload file to s3
// assign: save file in /tmp folder and return the absolute path
func SaveFile(file io.Reader) (string, error) {
	// Generate a random UUID for the filename
	fileUUID := uuid.New().String()
	fileName := fileUUID + ".csv"

	// Create the absolute path for saving the file in the /tmp directory
	absPath := filepath.Join("/tmp", fileName)

	// Create the file in /tmp
	out, err := os.Create(absPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Copy the file content to the new file without reading it into memory
	_, err = io.Copy(out, file)
	return absPath, err
}

// GetLocalFileRowCount gets the row count without reading entire file
// RowSize depends on csv structure
//not accurate, just approximate row length
// func GetLocalFileRowCount(filePath string, rowSize int64) (int64, error) {
// 	fileInfo, err := os.Stat(filePath)
// 	if err != nil {
// 		return 0, err
// 	}

// 	// Retrieve the file size from the FileInfo
// 	fileSize := fileInfo.Size()
// 	// Assuming a uniform row size, you can estimate the number of rows
// 	// Adjust the rowSize variable based on your actual CSV file structure
// 	log.Println("file sz ", fileSize, rowSize)
// 	numRows := fileSize / int64(rowSize)
// 	return numRows, nil
// }
