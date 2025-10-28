package utils

import (
	"fmt"
	"path"
	"strings"
)

// BuildProcessedPath формирует путь для обработанного изображения.
// Например:
//
//	original/uuid.jpg + "resize" → processed/uuid_resize.jpg
func BuildProcessedPath(originalPath, suffix string) string {
	dir, file := path.Split(originalPath) // dir = "original/", file = "uuid.jpg"

	ext := path.Ext(file)                                      // ".jpg"
	name := strings.TrimSuffix(file, ext)                      // "uuid"
	processedFile := fmt.Sprintf("%s_%s%s", name, suffix, ext) // "uuid_resize.jpg"

	// Меняем директорию original → processed
	processedDir := strings.Replace(dir, "original", "processed", 1)

	return path.Join(processedDir, processedFile)
}
