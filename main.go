package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

const file_to_copy = "/home/adnan/data/Videos/[1080p] Jujutsu Kaisen 0 [Japanese] [MSub] [jjk_001].mp4"
const file_to_copy2 = "/home/adnan/data/Videos/[DUFORUM.IN]  Take Charge of Your Time by Ankur Warikoo.mp4"
const file_to_copy3 = "/home/adnan/data/Videos/Shutter Island (2010) [1080p]/Shutter.Island.2010.1080p.BluRay.x264.YIFY.mp4"

type FileServer struct{}

func (fs *FileServer) start() {
	ln, err := net.Listen("tcp", ":3000")

	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}

		go fs.readLoop(conn)
	}
}

func (fs *FileServer) readLoop(conn net.Conn) {

	var size int64
	binary.Read(conn, binary.LittleEndian, &size)

	buffer := make([]byte, size)
	n, err := conn.Read(buffer)

	file, err := os.Create(string(buffer[:n]))
	if err != nil {
		log.Println(err)
	}

	defer file.Close()

	binary.Read(conn, binary.LittleEndian, &size)
	fmt.Printf("size: %d\n", size)

	var total int64
	for total < size {
		var size_recieved int64

		if size-total > 4096 {
			size_recieved = 4096
		} else {
			size_recieved = size - total
		}

		// buf := new(bytes.Buffer)
		log.Println("loop")

		n, err := io.CopyN(file, conn, size_recieved)
		if err != nil {
			log.Println(err)
			if err != io.EOF {
				log.Println(err)
			}
			break
		}
		// fmt.Println(buf.Bytes())
		fmt.Printf("%d bytes recieved--------------------\n", n)
		total += n
	}
}

func sendFile(path string) {

	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
	}

	defer file.Close()

	stat, _ := file.Stat()
	size := stat.Size()
	fname := stat.Name()

	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		log.Println(err)
	}

	binary.Write(conn, binary.LittleEndian, int64(len(fname)))
	conn.Write([]byte(fname))

	binary.Write(conn, binary.LittleEndian, int64(size))

	var total int64
	for total < size {
		var size_to_send int64

		if size-total > 4096 {
			size_to_send = 4096
		} else {
			size_to_send = size - total
		}

		buf := make([]byte, size_to_send)

		_, err := io.ReadFull(file, buf)
		if err != nil {
			log.Println(err)
		}

		n, err := io.CopyN(conn, bytes.NewReader(buf), int64(size_to_send))
		if err != nil {
			log.Println(err)
		}

		total += n
		log.Printf("sent %d bytes chunk over the network\n", n)

	}

	log.Printf("written %d bytes over the network\n", size)

}

func main() {

	go func() {
		time.Sleep(1 * time.Second)
		sendFile(file_to_copy)
	}()

	go func() {
		time.Sleep(2 * time.Second)
		sendFile(file_to_copy2)
	}()

	go func() {
		time.Sleep(1 * time.Second)
		sendFile(file_to_copy3)
	}()

	fs := &FileServer{}
	fs.start()
}
