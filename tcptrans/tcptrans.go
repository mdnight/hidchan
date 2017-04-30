package tcptrans

import (
	"bytes"
	"log"
	"net"
	"strconv"
)

type TCPDict struct {
	Dict []uint16
}

//Transmit performs transmitting using tcp destination ports
func (d *TCPDict) Transmit(data, noise []byte, destIP string) error {
	dstaddrs, err := net.LookupIP(destIP)
	if err != nil {
		log.Fatal(err)
	}
	var conns []net.Conn
	if len(d.Dict) == 0 {
		for i := 0; i < cap(d.Dict); i++ {
			d.Dict[i] = uint16(i) + 27000
		}
	}
	for _, i := range d.Dict {
		tmp, err := net.Dial("tcp4", dstaddrs[0].String()+strconv.Itoa(int(i)))
		if err != nil {
			return err
		}
		conns = append(conns, tmp)
		defer tmp.Close()
	}
	for _, i := range data {
		_, err := conns[i].Write(noise)
		if err != nil {
			return err
		}
	}
	_, err = conns[0x55].Write(noise)
	_, err = conns[0xaa].Write(noise)
	if err != nil {
		return err
	}
	return nil
}

//Receive performs receiving using TCP destination ports
func (d *TCPDict) Receive() ([]byte, error) {
	var listens []net.Listener
	var result []byte
	resBytes := make(chan byte)

	for _, i := range d.Dict {
		tmp, err := net.Listen("tcp4", ":"+strconv.Itoa(int(i)))
		if err != nil {
			return []byte{}, err
		}
		listens = append(listens, tmp)
		defer tmp.Close()
	}
	for key, val := range listens {
		go func() {
			for {
				tmpBuf := make([]byte, 8192)
				c, _ := val.Accept()
				n, err := c.Read(tmpBuf)
				if n > 0 && err == nil {
					resBytes <- byte(key)
				}
				tmpBuf = tmpBuf[:0]
			}
		}()
	}
	for {
		result = append(result, <-resBytes)
		tmpLen := len(result)
		if tmpLen >= 2 && bytes.Equal([]byte{0x55, 0xaa}, result[tmpLen-2:]) {
			break
		}

	}
	return result, nil
}
