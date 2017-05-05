package btnet

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"net"
	"io"
	"time"
	//"time"
	"util"
)

func StartTCPServer(addr string, handler func(net.Conn)) {
	util.TPrintf("Starting the TCP Server...\n")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		// complain about things dying
		util.EPrintf("labtcp StartTCPServer: %s\n", err)
	}
	go func(ln net.Listener) {
		conn, err := ln.Accept()
		if err != nil {
			// complain about a thing
			util.EPrintf("labtcp StartTCPServer: %s\n", err)
		}
		go handler(conn)
	}(ln)
}

func DoDial(addr *net.TCPAddr, data []byte) *net.TCPConn {
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		// Cry
		util.EPrintf("labtcp DoDial: %s\n", err)
		return conn
	}
	util.TPrintf("writing out to connection\n")
	bytesSent, err := conn.Write(data)
	util.Printf("bytes: %v", bytesSent)
	if err != nil {
		util.EPrintf("labtcp DoDail: %s\n", err)
	}
	return conn
	// return ReadMessage(conn)
}

func ReadHandshake(conn *net.TCPConn) []byte {
  // General strategy for reading handshakes
	// 1) The first byte for the length of the pstr
	// 2) Read that many bytes after to form a packet + 49
	// 3) Hope that nothing goes out of sync
	// 4) Check that the packets are reasonable
	// 5) repeat 3

	reader := bufio.NewReader(conn)

	// Grab the first 4 bytes
	msgLength := make([]byte, 1)
	response, err := reader.ReadByte()
	if err != nil {
		util.EPrintf("labtcp ReadHandshake: %s\n", err)
	}
	msgLength[0] = response

	util.TPrintf("msglength: %v\n", msgLength)
	// Parse the length of the message
	var length int8
	msgLengthDecodeBuf := bytes.NewReader(msgLength)
	errBinary := binary.Read(msgLengthDecodeBuf, binary.BigEndian, &length)
	if errBinary != nil {
		util.EPrintf("labtcp ReadHandshake binaryDecode: %s\n", errBinary)
	}
	util.TPrintf("length: %d\n", int(length))

	// Read zeros
	zerobuf := make([]byte, 8)
	for i:= 0; i < 8; i++ {
		response, err := reader.ReadByte()
		if err != nil {
			util.EPrintf("labtcp ReadHandshake: %s\n", err.Error())
			responseData := append(msgLength, zerobuf...)
			return responseData
			// break
		}
		if (response != 0x00) {
			util.EPrintf("labtcp ReadHandshake: badly formatted handshake\n")
			responseData := append(msgLength, zerobuf...)
			return responseData
		}

	}

	// Cross fingers
	// Read the number of bytes specified by length and hope it doesnt go out of sync
	msgbuf := make([]byte, int(length) + 48)
	for i := 0; i < int(length) + 48; i++ {
		response, err := reader.ReadByte()
		if err != nil {
			util.EPrintf("labtcp ReadHandshake: %s\n", err.Error())
			responseData := append(msgLength, zerobuf...)
			responseData = append(responseData,  msgbuf...)
			return responseData
			// break
		}
		msgbuf[i] = response
	}

	responseData := append(msgLength, zerobuf...)
	responseData = append(responseData, msgbuf...)
	return responseData
}

func ReadMessage(conn *net.TCPConn) []byte {
	// General strategy for reading packets back
	// 1) The first four bytes for the length of the packets
	// 2) Read that many bytes after to form a packet
	// 3) Hope that nothing goes out of sync
	// 4) Check that the packets are reasonable
	// 5) repeat 3

	reader := bufio.NewReader(conn)

	// Grab the first 4 bytes
	msgLength := make([]byte, 4)
	for i := 0; i < 4; i++ {
		response, err := reader.ReadByte()
		if err != nil {
			util.EPrintf("labtcp ReadMessage: %s\n", err)
		}
		msgLength[i] = response
	}

	util.TPrintf("msglength: %v\n", msgLength)
	// Parse the length of the message
	var length int32
	msgLengthDecodeBuf := bytes.NewReader(msgLength)
	errBinary := binary.Read(msgLengthDecodeBuf, binary.BigEndian, &length)
	if errBinary != nil {
		util.EPrintf("labtcp ReadMessage: %s\n", errBinary)
	}
	util.TPrintf("length: %d\n", length)

	// Cross fingers
	// Read the number of bytes specified by length and hope it doesnt go out of sync
	msgbuf := make([]byte, length)
	for i := 0; i < int(length); i++ {
		response, err := reader.ReadByte()
		if err != nil {
			util.EPrintf("labtcp ReadMessage: %s\n", err.Error())
			break
		}
		msgbuf[i] = response
	}

	responseData := append(msgLength, msgbuf...)
	//
	// response, err := bufio.NewReader(conn).ReadString('\n')
	// if err != nil {
	//   // Cry
	//   fmt.Println(err)
	// }
	return responseData
}

// TODO: This doesnt work... wtf
func IsConnectionClosed(conn *net.TCPConn) bool {
	one := []byte{}
	conn.SetReadDeadline(time.Now())
	if data, err := conn.Read(one); err == io.EOF {
	  // l.Printf(logger.LevelDebug, "%s detected closed LAN connection", id)
	  conn.Close()
	  // conn = nil
		return true
	} else {
	  // var zero time.Time
		util.Printf("This is the data: %v, %v\n", data, err)
	  conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		return false
	}
}
