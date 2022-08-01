package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func main() {
	server, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("Server has started")

	for {
		connect, err := server.Accept()

		if err != nil {
			log.Fatal(err)
			return
		}
		go countTCPAndUDP(connect)

	}

}

type Protocols struct {
	TCP  int
	UDP  int
	IPv4 int
	IPv6 int
}

var (
	eth layers.Ethernet
	ip4 layers.IPv4
	ip6 layers.IPv6
	tcp layers.TCP
	udp layers.UDP
	dns layers.DNS
)

func countTCPAndUDP(connect net.Conn) {

	parser := gopacket.NewDecodingLayerParser(
		layers.LayerTypeEthernet,
		&eth,
		&ip4,
		&ip6,
		&tcp,
		&udp,
		&dns,
	)

	decoded := make([]gopacket.LayerType, 0, 10)
	counter := Protocols{}
	for {
		read, err := receiveAll(connect, 8)

		if err != nil {
			log.Fatal(err)
			return
		}

		size := binary.BigEndian.Uint64(read)

		read, err = receiveAll(connect, size)
		read = make([]byte, size)

		if size == 4 && string(read) == "STOP" {
			break
		}

		fmt.Printf("File size: %v\n", size)

		parser.DecodeLayers(read, &decoded)

		for _, layer := range decoded {
			if layer == layers.LayerTypeTCP {
				counter.TCP++
			}
			if layer == layers.LayerTypeUDP {
				counter.UDP++
			}
			if layer == layers.LayerTypeIPv4 {
				counter.IPv4++
			}
			if layer == layers.LayerTypeIPv6 {
				counter.IPv6++
			}
		}

	}

	res := "TCP: " + strconv.Itoa(counter.TCP) + "\n" +
		"UDP: " + strconv.Itoa(counter.UDP) + "\n" +
		"IPv4: " + strconv.Itoa(counter.IPv4) + "\n" +
		"IPv6: " + strconv.Itoa(counter.IPv6) + "\n"

	connect.Write([]byte(res))
	connect.Close()
	fmt.Println("File receiving has ended")
	fmt.Println()
}

func receiveAll(connect net.Conn, size uint64) ([]byte, error) {
	read := make([]byte, int(size))
	n, err := connect.Read(read)

	if err != nil {

		return []byte{}, err
	}
	for uint64(n) != size {
		n1 := size - uint64(n)
		read2 := make([]byte, n1)
		n2, err := connect.Read(read2)
		read = append(read, read2...)
		if err != nil {
			return []byte{}, err
		}
		n += n2
	}
	return read, nil
}
