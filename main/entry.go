package main

import (
	"encoding/binary"
	"fmt"
	"log"

	"github.com/Vishwanath-V/exampleslottedpage/slottedpage"
)

func main() {
	fm := slottedpage.FileManager{FileDirectory: "/Users/vvelpula/Documents/Vishwanath/go_learning/exampleslottedpage"}
	//log.Printf("File Present:%d\n", fm.FilePresent("test.hex"))
	var i uint32 = 89
	m := make(map[uint32][]byte)
	for i = 0; i < 10000; i++ {
		log.Println("------------------------------------------")
		//fm.WriteDataToFile("test.hex", convertKVtoByte(i+1, "{key:"+fmt.Sprint(i+1)+",value:\"Vishwanath-Test-"+fmt.Sprint(i+1)+"\"}"))
		m[i] = convertKVtoByte(i+1, "{key:"+fmt.Sprint(i+1)+",value:\"Vishwanath-Test-"+fmt.Sprint(i+1)+"\"}")
		if i%1000 == 0 {
			fm.WriteBatchDataToFile("test.hex", m)
			clear(m)
		}
	}
	fm.WriteBatchDataToFile("test.hex", m)
	clear(m)

	f, fh, _ := fm.ReadFile("test.hex")
	defer f.Close()
	log.Println(fh)
	//fm.WriteDataToFile("test.hex", convertKVtoByte(i+1, "{key:"+fmt.Sprint(i+1)+",value:\"Vishwanath-Test-"+fmt.Sprint(i+1)+"\"}"))

}

func convertKVtoByte(k uint32, v string) []byte {
	item := make([]byte, binary.Size(k)+binary.Size([]byte(v)))
	binary.BigEndian.PutUint32(item[0:4], k)
	copy(item[4:], v)
	//log.Printf("Item:%b\n", item)
	return item
}
