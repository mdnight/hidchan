package tcptrans_test

import (
	"bytes"
	"github.com/mdnight/hidchan/tcptrans"
	"log"
	"testing"
)

// func TestTCPTransmitting(t *testing.T) {
// 	l, err := net.Listen("tcp4", ":8081")
// 	if err != nil {
// 		log.Fatalln(err.Error())
// 	}
// 	go func() {
// 		for {
// 			l.Accept()

// 		}
// 	}()
// }

func TestTCPSeqTransmition(t *testing.T) {
	receiveBuf := make([]byte, 100)
	transmitBuf := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
		1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
		1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
		1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
		1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
		1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
		1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
		1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
		1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
		1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
	}
	go func() {
		receiveBuf, _ = tcptrans.ReceiveWithSeq("enp0s31f6", 8081)
	}()

	err := tcptrans.TransmitWithSeq(receiveBuf, receiveBuf, "192.168.1.121", 8081)
	if err != nil {
		log.Fatalln(err.Error())
	}
	if !bytes.Equal(transmitBuf, receiveBuf) {
		t.Error("buffers are not equal")
	}
}
