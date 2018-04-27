package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

func main() {
	port := flag.Int("port", 8085, "The http server port")
	flag.Parse()

	router := routing.New()
	router.Get("/<server>", handler)
	router.Get("/<server>/<port>", handler)
	server := fasthttp.Server{
		Handler: router.HandleRequest,
		GetOnly: true,
		DisableKeepalive: true,
		ReadBufferSize: 0,
		WriteBufferSize: 0,
		ReadTimeout: time.Duration(time.Second * 2),
		WriteTimeout: time.Duration(time.Second * 2),
		MaxConnsPerIP: 3,
		MaxRequestsPerConn: 2,
		MaxRequestBodySize: 0,
	}

	go handleStop()

	panic(server.ListenAndServe(fmt.Sprintf(":%v", *port)))
}

func handler(c *routing.Context) error {
	c.SetContentType("application/json")
	server := c.Param("server")
	port := parsePort(c.Param("port"))
	status := ping(server, port)
	_, err := c.WriteString(status)
	return err
}

func handleStop() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "stop" {
			fmt.Println("Stopping...")
			os.Exit(0)
			break
		}
	}
}

func parsePort(port string) int {
	if port == "" {
		return 25565
	}

	i, e := strconv.ParseInt(port, 10, 32)
	if e != nil {
		return 25565
	}

	return int(i)
}

func ping(server string, port int) (string) {
	con, err := net.Dial("tcp", fmt.Sprintf("%s:%v", server, port))
	if err != nil {
		return err.Error()
	}

	defer con.Close()
	deadline := time.Now().Add(5 * time.Second)
	con.SetDeadline(deadline)
	con.SetReadDeadline(deadline)
	con.SetWriteDeadline(deadline)

	// handshake
	_, err = con.Write(handshake(server, port).Bytes())
	if err != nil {
		return err.Error()
	}

	// requestStatus request
	_, err = con.Write(requestStatus().Bytes())
	if err != nil {
		return err.Error()
	}

	// read length
	r := bufio.NewReader(con)
	l, err := binary.ReadUvarint(r)
	if err != nil {
		return err.Error()
	}

	// read data
	data := make([]byte, l)
	_, err = io.ReadFull(r, data)
	if err != nil {
		return err.Error()
	}

	// trim to json string
	_, i0 := binary.Uvarint(data)
	_, i1 := binary.Uvarint(data[i0:])
	return string(data[i0+i1:])
}

// handshake payload
func handshake(server string, port int) *bytes.Buffer {
	var buf bytes.Buffer
	buf.WriteByte(0x00) // id
	buf.WriteByte(0x47) // proto
	buf.Write(varint(len(server))) // length
	buf.WriteString(server) // ip
	binary.Write(&buf, binary.BigEndian, int16(port)) // port
	buf.WriteByte(0x01)
	return wrap(&buf)
}

// requestStatus request payload
func requestStatus() *bytes.Buffer {
	var buf bytes.Buffer
	buf.WriteByte(0x00) // id
	return wrap(&buf)
}

// wraps the buffer into a packet
func wrap(b *bytes.Buffer) *bytes.Buffer {
	var buf bytes.Buffer
	buf.Write(varint(len(b.Bytes())))
	buf.Write(b.Bytes())
	return &buf
}

// encode int
func varint(i int) []byte {
	buf := make([]byte, 10)
	l := binary.PutUvarint(buf, uint64(i))
	return buf[:l]
}