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
	//var conns [256]net.Conn
	dstaddrs, err := net.LookupIP(destIP)
	if err != nil {
		log.Fatal(err)
	}
	if len(d.Dict) == 0 {
		for i := 0; i < cap(d.Dict); i++ {
			d.Dict[i] = uint16(i) + 27000
		}
	}
	data = bytes.Join([][]byte{data, []byte{0x55, 0xaa}}, []byte{})
	// for _, value := range data {
	// 	if conns[value] == nil {
	// 		conns[value], err = net.Dial("tcp4",
	// 			dstaddrs[0].String()+":"+
	// 				strconv.Itoa(int(d.Dict[value])))
	// 		if err != nil {
	// 			return err
	// 		}
	// 		defer conns[value].Close()
	// 	}
	// 	_, err := conns[value].Write(noise)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	time.Sleep(10 * time.Millisecond)
	//}
	for _, value := range data {
		tmp, err := net.Dial("tcp4", dstaddrs[0].String()+":"+strconv.Itoa(int(d.Dict[value])))
		if err != nil {
			return err
		}
		_, err = tmp.Write(noise)
		if err != nil {
			return err
		}
		tmp.Close()
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
		defer tmp.Close()
		if err != nil {
			return []byte{}, err
		}
		listens = append(listens, tmp)
	}

	for key, val := range listens {
		go handleConnection(val, key, resBytes)
	}

	for {
		result = append(result, <-resBytes)
		tmpLen := len(result)
		if tmpLen >= 2 && bytes.Equal([]byte{0x55, 0xaa}, result[tmpLen-2:]) {
			for _, val := range listens {
				val.Close()
			}
			return result[:tmpLen-2], nil
		}
	}
}

func handleConnection(c net.Listener, port int, ch chan byte) {
	for {
		c.Accept()
		//buf := make([]byte, 1024)
		//n, _ := conn.Read(buf)
		//buf = buf[:0]
		//if n > 0
		{
			ch <- byte(port)
		}
	}
}
