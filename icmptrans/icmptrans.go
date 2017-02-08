package icmptrans

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"os"
	"time"
)

//EchoTransmit performs transmition with echo ICMP
func EchoTransmit(data []byte, destIP string) error {

	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return err
	}
	defer c.Close()
	{
		bs := make([]byte, 4)
		binary.LittleEndian.PutUint32(bs, uint32(len(data)))
		wm := icmp.Message{
			Type: ipv4.ICMPTypeEchoReply,
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
		if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(destIP)}); err != nil {
			return err
		}
	}
	for i := 0; i < len(data); i = i + 20 {
		wm := icmp.Message{
			Type: ipv4.ICMPTypeEchoReply,
			Code: 0,
			Body: &icmp.Echo{
				ID: os.Getpid() & 0xffff, Seq: 2,
				Data: data[i : i+20],
			},
		}
		wb, err := wm.Marshal(nil)
		if err != nil {
			return err
		}
		if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(destIP)}); err != nil {
			time.Sleep(time.Second)
			i = i - 20
			continue
		}
	}
	if i := len(data) % 20; i != 0 {
		wm := icmp.Message{
			Type: ipv4.ICMPTypeEchoReply,
			Code: 0,
			Body: &icmp.Echo{
				ID: os.Getpid() & 0xffff, Seq: 2,
				//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
				Data: data[len(data)-i-1 : len(data)-1],
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

//EchoReceive performs receiving data, sent with echo ICMP
func EchoReceive() (result []byte, err error) {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return []byte{}, err
	}
	defer c.Close()
	rb := make([]byte, 1500)

	//first packet contains data len

	n, _, err := c.ReadFrom(rb)
	if err != nil {
		return []byte{}, err
	}
	rm, err := icmp.ParseMessage(1, rb[:n])
	if err != nil {
		return []byte{}, err
	}
	tmp, _ := rm.Body.Marshal(1)
	result = bytes.Join([][]byte{result, tmp}, []byte{})

	return result, nil
}
