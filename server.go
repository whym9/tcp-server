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

var counter Protocols

var (
	eth     layers.Ethernet
	ip4     layers.IPv4
	ip6     layers.IPv6
	tcp     layers.TCP
	udp     layers.UDP
	dns     layers.DNS
	llc     layers.LLC
	payload gopacket.Payload
	tls     layers.TLS
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
		&llc,
		&payload,
		&tls,
	)

	decoded := make([]gopacket.LayerType, 0, 10)

	for {
		read := make([]byte, 8)

		_, err := connect.Read(read)
		if err != nil {
			log.Fatal(err)
			return
		}
		size := binary.BigEndian.Uint64(read)
		read = make([]byte, size)
		_, err = connect.Read(read)
		if size == 4 && string(read) == "STOP" {
			break
		}
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Printf("File size: %v\n", size)

		err = parser.DecodeLayers(read, &decoded)

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
	//fmt.Println(string(res))
	connect.Write([]byte(res))
	connect.Close()
}
