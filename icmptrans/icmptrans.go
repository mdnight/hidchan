package icmptrans

import (
	"bytes"
	"encoding/binary"
	"errors"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"os"
	"time"
)

//ICMPEcho for transmittion using Echo
type ICMPEcho struct{}

//Transmit performs transmition with echo ICMP
func (l *ICMPEcho) Transmit(data []byte, destIP string) error {

	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return err
	}
	defer c.Close()
	//send data size first
	{
		bs := make([]byte, 4)
		dataSize := len(data)
		if dataSize > 50000 {
			return errors.New("upload size exceeded (max 50kB)")
		}
		binary.LittleEndian.PutUint32(bs, uint32(len(data)))
		wm := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID: os.Getpid() & 0xffff, Seq: 2,
				Data: bs,
			},
		}
		wb, err := wm.Marshal(nil)
		if err != nil {
			return err
		}
		log.Println(wb)
		if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(destIP)}); err != nil {
			return err
		}
	}

	var blocksize int
	if len(data) < 20 {
		blocksize = len(data)
	} else {
		blocksize = 20
	}

	for i := 0; i < len(data); i = i + blocksize {
		wm := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID: os.Getpid() & 0xffff, Seq: 2,
				Data: data[i : i+blocksize],
			},
		}
		wb, err := wm.Marshal(nil)
		if err != nil {
			return err
		}
		if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(destIP)}); err != nil {
			time.Sleep(time.Second)
			i = i - blocksize
			continue
		}
	}
	if i := len(data) % blocksize; i != 0 {
		wm := icmp.Message{
			Type: ipv4.ICMPTypeEchoReply,
			Code: 0,
			Body: &icmp.Echo{
				ID: os.Getpid() & 0xffff, Seq: 2,
				Data: data[len(data)-i:],
			},
		}
		wb, err := wm.Marshal(nil)
		if err != nil {
			return err
		}
		if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(destIP)}); err != nil {
			return err
		}
	}
	return nil
}

//Receive performs receiving data, sent with echo ICMP
func (l *ICMPEcho) Receive() (result []byte, err error) {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	c.SetReadDeadline(time.Now().Add(5 * time.Minute))
	if err != nil {
		return []byte{}, err
	}
	defer c.Close()
	rb := make([]byte, 1500)
	//first packet contains size of transmitting data in bytes
	n, _, err := c.ReadFrom(rb)
	if err != nil {
		return []byte{}, err
	}
	rm, err := icmp.ParseMessage(1, rb[:n])
	if err != nil {
		return []byte{}, err
	}
	size, _ := rm.Body.Marshal(1)
	dataSize := int(binary.LittleEndian.Uint32(size[4:]))
	log.Println("dataSize: ", dataSize)
	if dataSize > 1024*20 {
		return []byte{}, errors.New("upload size exceeded (max 1GiB)")
	}
	//data upload
	var i int
	for i < dataSize {
		n, _, err = c.ReadFrom(rb)
		if err != nil {
			return []byte{}, err
		}

		rm, err = icmp.ParseMessage(1, rb[:n])
		if err != nil {
			return []byte{}, err
		}
		tmp, _ := rm.Body.Marshal(1)
		tmp = tmp[4:]
		result = bytes.Join([][]byte{result, tmp}, []byte{})
		i = i + len(tmp)
	}
	return result, nil
}
