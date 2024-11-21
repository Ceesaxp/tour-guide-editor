// internal/types/types.go - Add custom types
package types

import "os"

// MultipartFile implements multipart.File interface
type MultipartFile struct {
    *os.File
}

func NewMultipartFile(file *os.File) *MultipartFile {
    return &MultipartFile{File: file}
}
