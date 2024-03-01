package slottedpage

import (
	"encoding/binary"
	"log"
)

const (
	fileHeaderSize uint8 = 6 //2 Bytes noofpages + 4 Bytes totalrows
)

type fileHeaderMetadata struct {
	totalpages uint16
	totalrows  uint32
}

type fileHeaderPtr struct {
	*fileHeaderMetadata
}

func (fhm fileHeaderMetadata) getFileHeaderBytes() []byte {
	b := make([]byte, fileHeaderSize)
	binary.BigEndian.PutUint16(b[0:2], fhm.totalpages)
	binary.BigEndian.PutUint32(b[2:6], fhm.totalrows)
	log.Printf("FileHeaderBytes:%b\n", b)
	return b
}

func (fhm fileHeaderMetadata) GetTotalPages() uint16 {
	return fhm.totalpages
}
