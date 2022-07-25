package main

import (
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
			break
		}
		if err != nil {
			fmt.Printf("oh no")
			panic(err)

		}

		fmt.Println(len(data))
		_, err = connect.Write(data)

		fmt.Println(".")
	}
	connect.Write([]byte("ok"))

	read := make([]byte, 1024)
	for {
		_, err = connect.Read(read)

		if err != nil || err == io.EOF {

			break
		}

		fmt.Println(string(read))
	}

	connect.Close()

}
