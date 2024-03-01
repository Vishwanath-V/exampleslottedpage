package slottedpage

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

const (
	defaultPageSize uint16 = 4096 // 4KB Page
	pageHeaderSize  uint8  = 5    //5Bytes Header Struct
	slotInfoSize    uint8  = 8    //8 Bytes
)

type Header struct {
	TotalItems      uint8  //totaldataitems in a page
	FreeOffsetStart uint16 //first free offset from page start
	FreeOffsetEnd   uint16 //first free offset from page end
}

type Slot struct {
	Key      uint32 //key of the value
	Offset   uint16 //offset wrt to pageheader - final offset = pageHeaderSize + offset gives the final offset
	DataSize uint16 //datasize to be retrieved from offset
	Item     []byte //slot data byte array
} //each Slot = SlotInfo is of 7Bytes + Item byte array

type PageInfo struct {
	PageHeader *Header
	Slots      []*Slot
}

func (h Header) getSize() int {
	return binary.Size(h.TotalItems) +
		binary.Size(h.FreeOffsetStart) +
		binary.Size(h.FreeOffsetEnd)
}

func (s Slot) getSize() int {
	return binary.Size(s.Key) +
		binary.Size(s.Offset) +
		binary.Size(s.DataSize)
}

func (h Header) getByteArray() []byte {
	b := make([]byte, h.getSize())
	//x := make([]byte, 0)
	//return fmt.Append(b, h.TotalItems, binary.LittleEndian.AppendUint16(x, h.FreeOffsetStart), binary.LittleEndian.AppendUint16(x, h.FreeOffsetEnd))
	//return fmt.Append(b, h.TotalItems, binary.LittleEndian.AppendUint16(binary.LittleEndian.AppendUint16(x, h.FreeOffsetStart), h.FreeOffsetEnd))
	b[0] = h.TotalItems
	binary.BigEndian.PutUint16(b[1:3], h.FreeOffsetStart)
	binary.BigEndian.PutUint16(b[3:5], h.FreeOffsetEnd)
	return b
}

func (s Slot) getByteArrayWithOutItem() []byte {
	b := make([]byte, slotInfoSize)
	binary.BigEndian.PutUint32(b[0:4], s.Key)
	binary.BigEndian.PutUint16(b[4:6], s.Offset)
	binary.BigEndian.PutUint16(b[6:8], s.DataSize)
	log.Printf("SlotInfoBytes:%b\n", b)
	return b
}

func getDefaultPageHeader() Header {
	var h Header
	h.TotalItems = 0
	h.FreeOffsetStart = uint16(h.getSize())
	h.FreeOffsetEnd = defaultPageSize
	return h
}

func getPageHeaderBytes(file *os.File, pageId uint16) ([]byte, error) {
	headerBytes := make([]byte, getDefaultPageHeader().getSize())
	_, err := file.ReadAt(headerBytes, int64(pageId*defaultPageSize))
	var h Header
	h.TotalItems = headerBytes[0]
	h.FreeOffsetStart = binary.BigEndian.Uint16(headerBytes[1:3])
	h.FreeOffsetEnd = binary.BigEndian.Uint16(headerBytes[1:3])
	return headerBytes, err
}

func (h Header) getSlotInfo(pageContent []byte) []*Slot {
	var slotInfo []*Slot
	//pos := h.getSize() // page header size
	log.Println("header:", h)
	for i := 0; i < int(h.TotalItems); i++ {
		//pos+(i*8) // starts the index of slotdetails
		//pos+(i*8) // key for 4 bytes
		//pos+(i*8)+4 // offset for 2 byte
		//pos+(i*8)+6 // data for 2 bytes
		pos := int(pageHeaderSize) + (i * int(slotInfoSize))
		offset := binary.BigEndian.Uint16(pageContent[pos+4 : pos+6])
		datasize := binary.BigEndian.Uint16(pageContent[pos+6 : pos+8])

		slotInfo = append(slotInfo, &Slot{
			Key:      binary.BigEndian.Uint32(pageContent[pos : pos+4]),
			Offset:   offset,
			DataSize: datasize,
			Item:     pageContent[uint16(pageHeaderSize)+offset-datasize : uint16(pageHeaderSize)+offset],
		})
	}
	return slotInfo
}

func getPageContent(file *os.File, pageId uint16) PageInfo {
	pagebytes := make([]byte, defaultPageSize)
	bytesread, _ := file.ReadAt(pagebytes, int64(fileHeaderSize)+(int64(pageId-1)*int64(defaultPageSize))) //firstpage starts at just after fileheader
	log.Printf("BytesRead:%d\n", bytesread)
	log.Printf("Bytes:%b\n", pagebytes[0:20])
	var h Header = Header{TotalItems: pagebytes[0],
		FreeOffsetStart: binary.BigEndian.Uint16(pagebytes[1:3]),
		FreeOffsetEnd:   binary.BigEndian.Uint16(pagebytes[3:5])}
	pi := PageInfo{PageHeader: &h, Slots: h.getSlotInfo(pagebytes)}
	return pi
}

func createNewPageContent() PageInfo {
	pagebytes := make([]byte, defaultPageSize)
	var h Header = getDefaultPageHeader()
	copy(pagebytes, h.getByteArray())
	pi := PageInfo{PageHeader: &h, Slots: h.getSlotInfo(pagebytes)}
	return pi
}

func createNewPageContentUpdatePageInfoPtr(pi *PageInfo) {
	pagebytes := make([]byte, defaultPageSize)
	var h Header = getDefaultPageHeader()
	copy(pagebytes, h.getByteArray())
	//pi := PageInfo{PageHeader: &h, Slots: h.getSlotInfo(pagebytes)}
	pi.PageHeader.TotalItems = 0
	pi.PageHeader.FreeOffsetStart = uint16(pageHeaderSize)
	pi.PageHeader.FreeOffsetEnd = defaultPageSize
	return
}

// only when slotInfoSize+dataSize < PageFreeSpace, data item can be written in the page else no create new page
func (p PageInfo) calculatePageFreeSpace() int {
	return int(p.PageHeader.FreeOffsetEnd) - int(pageHeaderSize) - int(p.PageHeader.TotalItems)*int(slotInfoSize)
}

func writeNewItemToPage(file *os.File, fh *fileHeaderMetadata, pi *PageInfo, item []byte) (int, error) {
	log.Printf("File Name to writeItem: %s\n", file.Name())
	log.Println(fh)
	lastPageId := fh.totalpages //pageid
	var pageOffsetInFile int64
	var sl Slot
	//var ph Header
	//var newPageOffset uint16

	//check for free space for the record lastPageId=0 no pages present
	log.Println(pi.calculatePageFreeSpace())
	if lastPageId != 0 && pi.calculatePageFreeSpace() > len(item)+int(slotInfoSize) {
		log.Println("Enough space available")
		pageOffsetInFile = int64(uint16(fileHeaderSize) + ((lastPageId - 1) * defaultPageSize)) //firstpage starts at just after fileheader
	} else {
		log.Println("Not Enough space available...Save changes & Create new Page")
		file.Sync()
		//pi = createNewPageContent()
		createNewPageContentUpdatePageInfoPtr(pi)
		fs, _ := file.Stat()
		pageOffsetInFile = fs.Size()
		fh.totalpages += 1 //increment new page counter in header
	}
	log.Printf("PageOffsetInFile: %d\n", pageOffsetInFile)
	log.Println("PageInfo:", pi.PageHeader)
	currStartOffset := pi.PageHeader.FreeOffsetStart
	pi.PageHeader.TotalItems += 1
	pi.PageHeader.FreeOffsetStart += uint16(slotInfoSize)
	pi.PageHeader.FreeOffsetEnd -= uint16(len(item))
	//newPageOffset = pi.PageHeader.FreeOffsetEnd - uint16(len(item))
	//prepare slot info definition
	sl = Slot{
		Key: binary.BigEndian.Uint32(item[0:4]),
		//Offset:   newPageOffset,
		Offset:   pi.PageHeader.FreeOffsetEnd,
		DataSize: uint16(len(item)),
		Item:     item,
	}
	pi.Slots = append(pi.Slots, &sl)
	//ph = Header{TotalItems: pi.PageHeader.TotalItems + 1, FreeOffsetStart: pi.PageHeader.FreeOffsetStart + uint16(slotInfoSize), FreeOffsetEnd: newPageOffset}

	log.Println("ph:", pi.PageHeader)
	log.Println("sl:", sl)
	/*file.WriteAt(item, pageOffsetInFile+int64(newPageOffset))
	file.WriteAt(sl.getByteArrayWithOutItem(), pageOffsetInFile+int64(pi.PageHeader.FreeOffsetStart))
	file.WriteAt(ph.getByteArray(), pageOffsetInFile+0)*/
	file.WriteAt(item, pageOffsetInFile+int64(pi.PageHeader.FreeOffsetEnd))
	file.WriteAt(sl.getByteArrayWithOutItem(), pageOffsetInFile+int64(currStartOffset)) //pi.Slots[len(pi.Slots)-1].getByteArrayWithOutItem()
	file.WriteAt(pi.PageHeader.getByteArray(), pageOffsetInFile+0)

	//update the file header metadata
	fh.totalrows += 1 //increment totalrows
	log.Println(fh)
	bw, _ := file.WriteAt(fh.getFileHeaderBytes(), 0)
	log.Println(bw)

	//push changes to disk
	//file.Sync()

	return 0, nil
}

func appendNewPage(file *os.File) {
	//var h Header = Header{TotalItems: 10, FreeOffsetStart: 255, FreeOffsetEnd: 4695}
	//var s Slot = Slot{}

	//pi := &PageInfo{PageHeader: &h, Slots: []*Slot{}}
	//fmt.Print(unsafe.Sizeof(*pi))

	var h Header
	var pi PageInfo
	fmt.Printf("HeaderSize:%d\n", binary.Size(pi))

	b := make([]byte, 4096)
	copy(b, h.getByteArray())
	t, _ := file.Stat()
	file.WriteAt(b, t.Size())
	file.Sync()
	//return
}
