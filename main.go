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
	"io/ioutil"
	"log"
	"os"
	"path"
)

var (
	workMode, defPath, name, compress *string
)

func initFlags() {
	workMode = flag.String("mode", "", "[T]ransmit or [R]eceive mode")
	defPath = flag.String("o", "./", "Saving directory")
	name = flag.String("n", "resData", "File name to save. Default: resData")
	compress = flag.String("c", "", "Use compression: [l]zw, [g]zip, [z]lib, [n]one")
	flag.Parse()
}

func main() {
	initFlags()

	switch *workMode {
	case "R":
		if *name == "" {
			log.Println("name flag was not set")
			os.Exit(0)
		}
		var data bytes.Buffer
		err := receive(&data)
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile(path.Join(*defPath, *name), data.Bytes(), 0644)
	case "T":
	default:
		log.Println("Select mode ([T]ransmit or [R]eceive)")
		os.Exit(0)
	}

}

func receive(data *bytes.Buffer) error {
	if *compress == "" {
		return errors.New("wrong compression flag value")
	}
	//receiving data!!!!!!!!!!!!!!!!!!!!!!!!!!

	//decompressing data
	switch *compress {
	case "l":
		var output = make([]byte, data.Len())
		dec := lzw.NewReader(data, lzw.LSB, 8)
		defer dec.Close()
		_, err := dec.Read(output)
		if err != nil {
			return err
		}
		data.Reset()
		data.Write(output)
	case "g":
		var output = make([]byte, data.Len())
		dec, err := gzip.NewReader(data)
		defer dec.Close()
		dec.Multistream(false)
		if err != nil {
			return err
		}
		data.Reset()
		data.Write(output)
	case "z":
		var output = make([]byte, data.Len())
		dec, err := zlib.NewReader(data)
		defer dec.Close()
		if err != nil {
			return err
		}
		data.Reset()
		data.Write(output)
	case "n":
		return nil
	}
	return nil
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
		return nil
	default:
		return errors.New("wrong compression flag value")
	}

	//do transmition
	return nil
}
