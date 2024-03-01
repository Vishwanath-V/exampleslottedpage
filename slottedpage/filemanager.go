package slottedpage

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
)

type FileManager struct {
	FileDirectory string
}

func (fm FileManager) fullPath(file string) string {
	if fm.FileDirectory == "" {
		return file
	}
	return fmt.Sprintf("%s/%s", fm.FileDirectory, file)
}

func (fm FileManager) filePresent(file string) bool {
	if _, err := os.Stat(fm.fullPath(file)); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func (fm FileManager) ReadFile(file string) (*os.File, fileHeaderMetadata, error) {
	var f *os.File
	var err error
	var fh fileHeaderMetadata
	if fm.filePresent(file) {
		log.Printf("File Present:\n")
		f, err = os.OpenFile(fm.fullPath(file), os.O_CREATE|os.O_RDWR, os.ModePerm)
		tp := make([]byte, 2)
		tr := make([]byte, 4)
		f.ReadAt(tp, 0)
		f.ReadAt(tr, 2)
		fh = fileHeaderMetadata{totalpages: binary.BigEndian.Uint16(tp),
			totalrows: binary.BigEndian.Uint32(tr)}
	} else {
		log.Printf("File Not Present..:\n")
		f, err = os.Create(fm.fullPath(file))
		fh = fileHeaderMetadata{totalpages: 0, totalrows: 0}
		f.WriteAt(fh.getFileHeaderBytes(), 0)
		f.Sync()
	}
	return f, fh, err
}

func (fm FileManager) FileSize(file string) int64 {
	f, err := os.Stat(fm.fullPath(file))
	if err != nil {
		return 0
	}
	return f.Size()
}

func (fm FileManager) WriteDataToFile(file string, item []byte) {
	f, fh, _ := fm.ReadFile(file)
	defer f.Close()
	pi := getPageContent(f, fh.totalpages)
	writeNewItemToPage(f, &fh, &pi, item)
}

func (fm FileManager) WriteBatchDataToFile(file string, items map[uint32][]byte) {
	f, fh, _ := fm.ReadFile(file)
	defer f.Close()
	// Iterate over all keys
	pi := getPageContent(f, fh.totalpages)
	for _, val := range items {
		log.Println("------------------------------------------")
		writeNewItemToPage(f, &fh, &pi, val)
	}
}
