package fileutils

import (
	"fmt"
	"time"
)

// ExcelExtension defines exel extension.
const ExcelExtension = "xlsx"

// DefaultSheetName defines default exel sheet name.
const DefaultSheetName = "Sheet1"

// NameForExportFiles returns name of export file with object name, date and extension.
func NameForExportFiles(exportObject, extension string) string {
	prefix := "export"
	now := time.Now().UTC()
	datePart := fmt.Sprintf("%d-%d-%d_%d:%d", now.Day(), now.Month(), now.Year(), now.Hour(), now.Minute())

	fileName := prefix + "_" + exportObject + "_" + datePart + "." + extension
	return fileName
}
