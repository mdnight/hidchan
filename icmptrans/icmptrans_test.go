package icmptrans_test

import (
	"bytes"
	"github.com/mdnight/hidchan/icmptrans"
	"log"
	"testing"
)

func TestEchoTransmission(t *testing.T) {
	//receiveBuf := make([]byte, 100)
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
	l := icmptrans.ICMPEcho{}

	//err := l.Transmit(transmitBuf, "127.0.0.1")
	receiveBuf, err := l.Receive()
	if err != nil {
		log.Fatalln(err.Error())
		t.Fail()
	}
	if !bytes.Equal(receiveBuf, transmitBuf) {
		log.Println("FUCK")
		t.Fail()
	}
}
