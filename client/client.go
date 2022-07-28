package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/google/gopacket/pcap"
)

func main() {
	filename := *flag.String("fileName", "lo.pcapng", "pcap file directory")
	flag.Parse()
	handle, err := pcap.OpenOffline(filename)
	defer handle.Close()
	if err != nil {
		log.Fatal(err)
		return
	}

	connect, err := net.Dial("tcp", "localhost:8080")

	for {

		data, _, err := handle.ZeroCopyReadPacketData()
		if err == io.EOF || err != nil {
			bin := make([]byte, 8)
			binary.BigEndian.PutUint64(bin, 4)
			connect.Write(bin)
			connect.Write([]byte("STOP"))
			break
		}
		bin := make([]byte, 8)
		binary.BigEndian.PutUint64(bin, uint64(len(data)))
		connect.Write([]byte(bin))

		if err != nil {
			panic(err)

		}

		_, err = connect.Write(data)

	}

	read := make([]byte, 1024)

	_, err = connect.Read(read)

	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println(string(read))

	connect.Close()

}

func file_size(filename string) uint64 {
	handle, err := pcap.OpenOffline(filename)

	if err != nil {
		log.Fatal(err)
		return 0
	}
	var size uint64 = 0
	for {
		data, _, err := handle.ZeroCopyReadPacketData()

		if err == io.EOF || err != nil {
			break
		}
		size += uint64(len(data))

	}

	return size
}
