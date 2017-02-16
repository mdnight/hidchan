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
	"compress/lzw"
	"compress/zlib"
	"errors"
	"flag"
	//"github.com/mdnight/hidchan/icmptrans"
	"github.com/mdnight/hidchan/udptrans"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
)

var (
	workMode, defPath, name, compress,
	destIP, srcIP, proto, readFileName *string
	port *int
)

func initFlags() {
	workMode = flag.String("mode", "", "[T]ransmit or [R]eceive mode")
	defPath = flag.String("o", "./", "Saving directory")
	name = flag.String("n", "resData", "File name to save. Default: resData")
	proto = flag.String("proto", "", "[T]CP, [U]DP, [I]CMP")
	port = flag.Int("port", 0, "TCP/UDP port")
	compress = flag.String("c", "", "Use compression: [l]zw, [g]zip, [z]lib, [n]one")
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
		err = ioutil.WriteFile(path.Join(*defPath, *name), data, 0644)
	case "T":
		if *destIP == "" {
			log.Println("destination ip was not set")
			os.Exit(0)
		}

		buf, err := ioutil.ReadFile(*readFileName)
		if err != nil {
			log.Fatal(err)
		}

		err = transmit(bytes.NewBuffer(buf))
		if err != nil {
			log.Fatal(err)
		}

	default:
		log.Println("Select mode ([T]ransmit or [R]eceive)")
		os.Exit(0)
	}

}

func receive() ([]byte, error) {
	if *compress == "" {
		return []byte{}, errors.New("wrong compression flag value")
	}
	//receiving data!!!!!!!!!!!!!!!!!!!!!!!!!!
	var (
		recData []byte
		err     error
	)

	switch *proto {
	case "T":
	case "U":
		recData, err = udptrans.ReceiveWithLen(port)
		if err != nil {
			return []byte{}, err
		}
	case "I":
	default:
		log.Println("Select correct value -proto ([T]CP, [U]DP, [I]CMP)")

	}

	//decompressing data
	switch *compress {
	case "l":
		dataBuf := bytes.NewBuffer(recData)
		var output = make([]byte, dataBuf.Len())
		dec := lzw.NewReader(dataBuf, lzw.LSB, 8)
		defer dec.Close()
		_, err := dec.Read(output)
		if err != nil {
			return []byte{}, err
		}
		dataBuf.Reset()
		dataBuf.Write(output)
	case "g":
		dataBuf := bytes.NewBuffer(recData)
		var output = make([]byte, dataBuf.Len())
		dec, err := gzip.NewReader(dataBuf)
		defer dec.Close()
		dec.Multistream(false)
		if err != nil {
			return []byte{}, err
		}
		dataBuf.Reset()
		dataBuf.Write(output)
	case "z":
		dataBuf := bytes.NewBuffer(recData)
		var output = make([]byte, dataBuf.Len())
		dec, err := zlib.NewReader(dataBuf)
		defer dec.Close()
		if err != nil {
			return []byte{}, err
		}
		dataBuf.Reset()
		dataBuf.Write(output)
	case "n":

	default:
		return []byte{}, errors.New("wrong compression flag value")
	}
	return recData, nil
}

func transmit(data *bytes.Buffer) error {
	switch *compress {
	case "l":
		com := lzw.NewWriter(data, lzw.LSB, 8)
		defer com.Close()
		_, err := com.Write(data.Bytes())
		if err != nil {
			return err
		}
	case "g":
		com, err := gzip.NewWriterLevel(data, gzip.BestCompression)
		defer com.Close()
		if err != nil {
			return err
		}
	case "z":
		com, err := zlib.NewWriterLevel(data, gzip.BestCompression)
		defer com.Close()
		if err != nil {
			return err
		}
	case "n":
		//
	default:
		return errors.New("wrong compression flag value")
	}

	//do transmition
	switch *proto {
	case "T":
	case "U":
		if *srcIP == "" {
			return errors.New("source IP was not set")
		}

		err := udptrans.TransmitWithLen(data.Bytes(), net.ParseIP(*srcIP), net.ParseIP(*destIP), port)
		if err != nil {
			return err
		}
	case "I":
	default:
		log.Println("Select correct value for -proto ([T]CP, [U]DP, [I]CMP)")
		os.Exit(0)
	}
	return nil
}
