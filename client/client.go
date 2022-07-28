package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

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

	size := file_size(filename)
	if size == 0 {
		fmt.Println("error")
		return
	}
	connect.Write([]byte(strconv.Itoa(size)))
	for {
		data, _, err := handle.ZeroCopyReadPacketData()
		if err == io.EOF || err != nil {
			break
		}
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

func file_size(filename string) int {
	handle, err := pcap.OpenOffline(filename)

	if err != nil {
		log.Fatal(err)
		return 0
	}
	size := 0
	for {
		data, _, err := handle.ZeroCopyReadPacketData()

		if err == io.EOF || err != nil {
			break
		}
		size += len(data)

	}

	return size
}
