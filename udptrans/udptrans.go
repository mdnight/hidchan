package udptrans

import (
	"encoding/binary"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

//TransmitWithLen performs transmitting using field "Length" of UDP datagram
func TransmitWithLen(data []byte, srcIP, dstIP net.IP, port *int) error {
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

//ReceiveWithLen performs receiving using field "Length" of UDP datagram
func ReceiveWithLen(port *int) ([]byte, error) {
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
