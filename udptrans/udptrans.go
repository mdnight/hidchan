package udptrans

import (
	"encoding/binary"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

type UDPLen struct {
}

type UDPsPortLen struct {
}

type UDPsPortAlphabet struct {
}

//Transmit performs transmitting using field "Length" of UDP datagram
func (l *UDPLen) Transmit(data []byte, srcIP, dstIP net.IP, port *int) error {
	destAddr, err := net.ResolveUDPAddr("udp4", strings.Join([]string{dstIP.String(), strconv.Itoa(*port)}, ":"))
	if err != nil {
		return err
	}
	sourceAddr, err := net.ResolveUDPAddr("udp4", strings.Join([]string{srcIP.String(), "0"}, ":"))
	if err != nil {
		return err
	}
	c, err := net.DialUDP("udp4", sourceAddr, destAddr)
	if err != nil {
		return err
	}
	defer c.Close()
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	commonBuf := make([]byte, 0xff00)
	for i := 0; i < 0xff; i++ {
		commonBuf[i] = byte(r.Intn(0x100))
	}

	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(len(data)))
	_, err = c.Write(bs)
	if err != nil {
		return err
	}
	for _, val := range data {
		size := uint(val)
		size = size << 8
		tmpBuf := commonBuf[0:size]

		tmpBuf = func(b []byte) []byte {
			tbuf := make([]byte, len(b))
			for i, v := range b {
				tbuf[i] = v ^ byte(r.Intn(0x100))
			}
			return tbuf
		}(tmpBuf)

		_, err := c.Write(tmpBuf)
		if err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 5)
	}
	return nil
}

//Receive performs receiving using field "Length" of UDP datagram
func (l *UDPLen) Receive(port *int) ([]byte, error) {
	destAddr, err := net.ResolveUDPAddr("udp4", ":"+strconv.Itoa(int(*port)))
	if err != nil {
		return []byte{}, err
	}

	c, err := net.ListenUDP("udp4", destAddr)
	if err != nil {
		return []byte{}, err
	}
	defer c.Close()
	bs := make([]byte, 4)
	_, _, err = c.ReadFromUDP(bs)

	buf := make([]byte, binary.LittleEndian.Uint32(bs))
	commonBuf := make([]byte, 0xff00)
	for i := 0; i < len(buf); i++ {
		n, _, err := c.ReadFromUDP(commonBuf)
		if err != nil {
			return []byte{}, err
		}
		buf[i] = byte(n >> 8)
	}
	return buf, nil
}

//Transmit performs transmittion using fields sPort and Len of UDP header
func (l *UDPsPortLen) Transmit(data []byte, srcIP, dstIP net.IP, port *int) error {
	var (
		sourceAddr, destAddr *net.UDPAddr
	)
	conns := make([]*net.UDPConn, 0xff)
	destAddr, err := net.ResolveUDPAddr("udp4", strings.Join([]string{dstIP.String(), strconv.Itoa(*port)}, ":"))
	if err != nil {
		return err
	}

	sourceAddr, err = net.ResolveUDPAddr("udp4", strings.Join([]string{srcIP.String(), "0"}, ":"))
	if err != nil {
		return err
	}

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	commonBuf := make([]byte, 0xff00)
	for i := 0; i < 0xff; i++ {
		commonBuf[i] = byte(r.Intn(0x100))
	}

	for i := range conns {
		sourceAddr, err = net.ResolveUDPAddr("udp4", strings.Join([]string{srcIP.String(), strconv.Itoa((i << 8) | r.Intn(0xff))}, ":"))
		conns[i], err = net.DialUDP("udp4", sourceAddr, destAddr)
		if err != nil {
			return err
		}
		defer conns[i].Close()
	}

	bs := make([]byte, 4)
	if len(data)%2 != 0 {
		data = append(data, byte(0x00))
	}
	log.Println(data)
	binary.LittleEndian.PutUint32(bs, uint32(len(data)))
	if _, err = conns[r.Intn(0x100)].Write(bs); err != nil {
		return err
	}
	time.Sleep(40 * time.Millisecond)

	for i := 0; i < len(data); i = i + 2 {
		size := uint(data[i+1])
		size = size << 8
		tmpBuf := commonBuf[0:size]

		tmpBuf = func(b []byte) []byte {
			tbuf := make([]byte, len(b))
			for j, v := range b {
				tbuf[j] = v ^ byte(r.Intn(0x100))
			}
			return tbuf
		}(tmpBuf)
		_, err := conns[data[i]].Write(tmpBuf)
		if err != nil {
			return err
		}
		time.Sleep(35 * time.Millisecond)
	}

	return nil
}

//Receive receives data using fields sPort and Len of UDP header
func (l *UDPsPortLen) Receive(destIP string, port *int) ([]byte, error) {
	destAddr, err := net.ResolveUDPAddr("udp4", ":"+strconv.Itoa(int(*port)))
	if err != nil {
		return []byte{}, err
	}

	c, err := net.ListenUDP("udp4", destAddr)
	if err != nil {
		return []byte{}, err
	}
	defer c.Close()

	dstaddrs, err := net.LookupIP(destIP)
	if err != nil {
		log.Fatal(err.Error())
	}
	dstIP := dstaddrs[0].To4()
	var workInterface net.Interface
	if ifaces, err := net.Interfaces(); err != nil {
		log.Fatal(err.Error())
	} else {
		for _, j := range ifaces {
			addrs, _ := j.Addrs()
			for _, k := range addrs {
				if strings.Split(k.String(), "/")[0] == dstIP.String() {
					workInterface = j
					break
				}
			}
		}
	}

	bs := make([]byte, 4)
	if _, _, err := c.ReadFromUDP(bs); err != nil {
		log.Fatal(err)
	}
	go fakeUDPReceive(c)

	buf, err := receiveBytes(workInterface.Name, binary.LittleEndian.Uint32(bs))
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("HERE")
	log.Println(buf)

	return buf, nil
}

func fakeUDPReceive(c *net.UDPConn) {
	commonBuf := make([]byte, 0xff00)
	for {
		if _, _, err := c.ReadFromUDP(commonBuf); err != nil {
			log.Fatal(err.Error())
		}
	}
}

func receiveBytes(ifName string, bufLen uint32) (buf []byte, err error) {
	buf = make([]byte, 0)
	handle, err := pcap.OpenLive(ifName, 500, false, pcap.BlockForever)
	if err != nil {
		return []byte{}, err
	}
	defer handle.Close()
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	handle.SetBPFFilter("udp and dst port 40400")
	for packet := range packetSource.Packets() {
		log.Println(len(buf), byte(packet.NetworkLayer().LayerPayload()[0]),
			byte(packet.NetworkLayer().LayerPayload()[4]))
		buf = append(buf, byte(packet.NetworkLayer().LayerPayload()[0]),
			byte(packet.NetworkLayer().LayerPayload()[4]))
		if uint32(len(buf)) >= bufLen {
			break
		}

	}
	return buf, err
}
