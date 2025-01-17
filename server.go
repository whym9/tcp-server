package main

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"net"
	"os"
	"strconv"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const dirName = "./saved_files"

var gormDB *gorm.DB

func main() {
	DB()
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		log.Fatalf("couldn't create path, %v", err)
	}
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

type Capture struct {
	TimeStamp      time.Time     `json: "time"`
	CaptureLength  int           `json: "caplength"`
	Length         int           `json: "length"`
	InterfaceIndex int           `json :  "index"`
	AccalaryData   []interface{} `json: "accalary"`
}

func countTCPAndUDP(connect net.Conn) {

	counter := Protocols{}

	fileName := dirName + "/" + time.Now().Format("02-01-2006-111111.456") + ".pcap"

	file, err := os.Create(fileName)

	if err != nil {
		log.Fatal(err)
		return
	}

	defer file.Close()

	w := pcapgo.NewWriter(file)
	w.WriteFileHeader(65535, layers.LinkTypeEthernet)
	for {
		read, err := receiveALL(connect, 8)

		if err != nil {
			log.Fatal(err)
			return
		}

		size := binary.BigEndian.Uint64(read)
		read, err = receiveALL(connect, size)
		if size == 4 && string(read) == "STOP" {
			break
		}
		var capi Capture
		json.Unmarshal(read, &capi)

		read, err = receiveALL(connect, 8)
		size = binary.BigEndian.Uint64(read)
		read, err = receiveALL(connect, size)

		fmt.Printf("File size: %v\n", size)

		packet := gopacket.NewPacket(read, layers.LayerTypeEthernet, gopacket.Default)

		packet.Metadata().CaptureInfo.Timestamp = capi.TimeStamp
		packet.Metadata().CaptureInfo.CaptureLength = capi.CaptureLength
		packet.Metadata().CaptureInfo.Length = capi.Length
		packet.Metadata().CaptureInfo.InterfaceIndex = capi.InterfaceIndex
		packet.Metadata().CaptureInfo.AncillaryData = capi.AccalaryData

		w.WritePacket(packet.Metadata().CaptureInfo, packet.Data())

	}
	counter, _ = Process(fileName)
	go saveToDB(counter, dirName+fileName)

	res := "TCP: " + strconv.Itoa(counter.TCP) + "\n" +
		"UDP: " + strconv.Itoa(counter.UDP) + "\n" +
		"IPv4: " + strconv.Itoa(counter.IPv4) + "\n" +
		"IPv6: " + strconv.Itoa(counter.IPv6) + "\n"

	connect.Write([]byte(res))
	connect.Close()
	fmt.Println("File receiving has ended")
	fmt.Println()

}

func receiveALL(connect net.Conn, size uint64) ([]byte, error) {
	read := make([]byte, size)

	_, err := io.ReadFull(connect, read)
	if err != nil {
		log.Fatal(err)
		return []byte{}, err
	}

	return read, nil
}

type Statistics struct {
	gorm.Model
	PathToFile string
	TCP        int
	UDP        int
	IPv4       int
	IPv6       int
}

func DB() *gorm.DB {
	sqlDB, err := sql.Open("mysql", "myserver_dsn")
	if err != nil {

		log.Fatal(err)
		return nil
	}
	gormDB, err = gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{})

	return gormDB
}

func saveToDB(counter Protocols, filePath string) {

	result := gormDB.Create(&Statistics{
		Model:      gorm.Model{},
		PathToFile: filePath,
		TCP:        counter.TCP,
		UDP:        counter.UDP,
		IPv4:       counter.IPv4,
		IPv6:       counter.IPv6,
	})

	if result.Error != nil {
		log.Fatal(result.Error)
		return
	}

	fmt.Printf("Record saved to Database!")

}

func Process(fileName string) (Protocols, error) {

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

	handle, err := pcap.OpenOffline(fileName)

	defer handle.Close()

	if err != nil {
		log.Fatal(err)
		return Protocols{}, err
	}

	counter := Protocols{}

	for {
		data, _, err := handle.ZeroCopyReadPacketData()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
			return Protocols{}, err
		}

		parser.DecodeLayers(data, &decoded)

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
	return counter, nil

}
