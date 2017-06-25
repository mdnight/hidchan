/*
Hidchan is intended for hidden data transmition bypassing firewall.

Copyright (C) 2017  Roman Isaev goose<AT>riseup.net

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License version 3 as
published by the Free Software Foundation.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"flag"
	"github.com/mdnight/hidchan/icmptrans"
	"github.com/mdnight/hidchan/tcptrans"
	"github.com/mdnight/hidchan/udptrans"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
)

var (
	workMode, defPath, name,
	destIP, srcIP, proto, readFileName, udpSubProto *string
	port     *int
	compress *bool
)

func initFlags() {
	workMode = flag.String("mode", "", "[T]ransmit or [R]eceive mode")
	defPath = flag.String("o", "./", "Saving directory")
	name = flag.String("n", "resData", "File name to save. Default: resData")
	proto = flag.String("proto", "", "[T]CP, [U]DP, [I]CMP")
	udpSubProto = flag.String("usp", "", "UDP[L]en, UDP[s]portLen, ")
	port = flag.Int("port", 0, "TCP/UDP port")
	compress = flag.Bool("c", false, "Use compression: true/false")
	destIP = flag.String("d", "", "Destination IPv4")
	srcIP = flag.String("s", "", "Source IPv4")
	readFileName = flag.String("i", "", "File for transmition")
	flag.Parse()
}

func main() {
	initFlags()

	if *proto == "" {
		log.Println("protocol was not set")
		os.Exit(0)
	}
	switch *workMode {
	case "R":
		if *name == "" {
			log.Println("name flag was not set")
			os.Exit(0)
		}
		data, err := receive()
		if err != nil {
			log.Fatal(err)
		}
		if err = ioutil.WriteFile(path.Join(*defPath, *name), data, 0644); err != nil {
			log.Fatal(err)
		}
	case "T":
		if *destIP == "" {
			log.Println("destination ip was not set")
			os.Exit(0)
		}

		buf, err := ioutil.ReadFile(*readFileName)
		if err != nil {
			log.Fatal(err)
		}

		if err = transmit(bytes.NewBuffer(buf)); err != nil {
			log.Fatal(err)
		}

	default:
		log.Println("Select mode ([T]ransmit or [R]eceive)")
		os.Exit(0)
	}

}

func receive() ([]byte, error) {
	//receiving data!!!!!!!!!!!!!!!!!!!!!!!!!!
	var (
		recData []byte
		err     error
	)

	switch *proto {
	case "T":
		l := tcptrans.TCPDict{}
		l.Dict = genTCPDict(27000, 28000)
		recData, err = l.Receive()
		if err != nil {
			return []byte{}, err
		}
	case "U":
		switch *udpSubProto {
		case "L":
			l := udptrans.UDPLen{}
			recData, err = l.Receive(port)
		case "s":
			l := udptrans.UDPsPortLen{}
			recData, err = l.Receive(*destIP, port)
		default:
			return []byte{}, errors.New("wrong udp subprotocol flag value")
		}
		if err != nil {
			return []byte{}, err
		}
	case "I":
		l := icmptrans.ICMPEcho{}
		recData, err = l.Receive()
	default:
		log.Println("Select correct value -proto ([T]CP, [U]DP, [I]CMP)")

	}

	//decompressing data
	if *compress == true {
		plainSize := binary.LittleEndian.Uint32(recData[len(recData)-4:])
		dataBuf := bytes.NewBuffer(recData)
		var output = make([]byte, plainSize)
		dec, _ := gzip.NewReader(dataBuf)
		dec.Multistream(false)
		if _, err = dec.Read(output); err != io.EOF {
			return []byte{}, err
		}
		dec.Close()
		dataBuf.Reset()
		return output, nil
	}
	return recData, nil
}

func transmit(data *bytes.Buffer) error {
	if *compress == true {
		var buf bytes.Buffer
		com, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
		_, err = com.Write(data.Bytes())
		if err != nil {
			return err
		}
		com.Close()
		log.Println(data.Bytes())
		data.Reset()
		data = &buf
		log.Println(data.Bytes())
	}
	//do transmition
	switch *proto {
	case "T":
		l := tcptrans.TCPDict{}
		l.Dict = genTCPDict(27000, 28000)
		err := l.Transmit(data.Bytes(), []byte{123, 32, 34, 34, 2, 95, 45, 99, 44}, *destIP)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	case "U":
		if *srcIP == "" {
			return errors.New("source IP was not set")
		}
		var err error
		switch *udpSubProto {
		case "L":
			l := udptrans.UDPLen{}
			err = l.Transmit(data.Bytes(), net.ParseIP(*srcIP), net.ParseIP(*destIP), port)
		case "s":
			l := udptrans.UDPsPortLen{}
			err = l.Transmit(data.Bytes(), net.ParseIP(*srcIP), net.ParseIP(*destIP), port)

		default:
			return errors.New("wrong udp subprotocol flag value")
		}
		if err != nil {
			return err
		}

	case "I":
		l := icmptrans.ICMPEcho{}
		if err := l.Transmit(data.Bytes(), *destIP); err != nil {
			return err
		}
	default:
		log.Println("Select correct value for -proto ([T]CP, [U]DP, [I]CMP)")
		os.Exit(0)
	}
	return nil
}

func genTCPDict(low, high uint16) []uint16 {
	m := make([]uint16, 256)
	if high-low < 256 || high >= 65535 {
		return []uint16{}
	}
	for i := 0; i < 256; i++ {
		m[i] = low
		low++
	}
	return m
}
