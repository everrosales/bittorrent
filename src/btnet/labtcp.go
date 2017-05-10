package btnet

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"time"
    // "errors"
	//"time"
	"util"
)

func StartTCPServer(addr string, handler func(*net.TCPConn)) {
	util.TPrintf("Starting the TCP Server...\n")
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		util.EPrintf("Labtcp StartTCPServer: %s", err)
	}
	ln, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		// complain about things dying
		util.EPrintf("labtcp StartTCPServer: %s\n", err)
	}
	go func(ln *net.TCPListener) {
		for {
			conn, err := ln.AcceptTCP()
			if err != nil {
				// complain about a thing
				util.EPrintf("labtcp StartTCPServer: %s\n", err)
			}
			go handler(conn)
		}
	}(ln)
}

func DoDial(addr *net.TCPAddr, data []byte) *net.TCPConn {
	util.IPrintf("Dialing: %v\n", addr.String())
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		// Cry
		util.EPrintf("labtcp DoDial: %s\n", err)
		return conn
	}
	_, err = conn.Write(data)
	if err != nil {
		util.EPrintf("labtcp DoDail: %s\n", err)
	}
	return conn
	// return ReadMessage(conn)
}

func ReadHandshake(conn *net.TCPConn) ([]byte, error) {
	// General strategy for reading handshakes
	// 1) The first byte for the length of the pstr
	// 2) Read that many bytes after to form a packet + 49
	// 3) Hope that nothing goes out of sync
	// 4) Check that the packets are reasonable
	// 5) repeat 3
    util.TPrintf("Reading Handshake from: %s\n", conn.RemoteAddr().String())
	reader := bufio.NewReader(conn)
	// Grab the first 4 bytes
	msgLength := make([]byte, 1)
    // conn.SetReadDeadline(time.Now().Add(time.Millisecond * 2000))
	response, err := reader.ReadByte()
	if err != nil {
		util.EPrintf("labtcp ReadHandshake: %s\n", err)
        util.EPrintf("Failed Handshake\n")
        return []byte{}, err
	}
	msgLength[0] = response

	// util.TPrintf("msglength: %v\n", msgLength)
	// Parse the length of the message
	var length uint8
	msgLengthDecodeBuf := bytes.NewReader(msgLength)
	errBinary := binary.Read(msgLengthDecodeBuf, binary.BigEndian, &length)
	if errBinary != nil {
		util.EPrintf("labtcp ReadHandshake binaryDecode: %s\n", errBinary)
        util.EPrintf("Failed Handshake\n")
        return []byte{}, err
	}
	// util.TPrintf("length: %d\n", int(length))

	// Read pstr
	pstrbuf := make([]byte, int(length))
	for i := 0; i < int(length); i++ {
        // conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
		response, err := reader.ReadByte()
		if err != nil {
			util.EPrintf("labtcp ReadHandshake: %s\n", err.Error())
			responseData := append(msgLength, pstrbuf...)
            util.EPrintf("Failed Handshake\n")
			return responseData, err
			// break
		}
		pstrbuf[i] = response

	}

	// Read zeros
	zerobuf := make([]byte, 8)
	for i := 0; i < 8; i++ {
        // conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
		response, err := reader.ReadByte()
		if err != nil {
			util.EPrintf("labtcp ReadHandshake: %s\n", err.Error())
			responseData := append(msgLength, zerobuf...)
            util.EPrintf("Failed Handshake\n")
			return responseData, err
			// break
		}
		if response != 0 {
			util.EPrintf("labtcp ReadHandshake: badly formatted handshake\n")
			responseData := append(msgLength, zerobuf...)
            util.EPrintf("Failed Handshake\n")
			return responseData, err
		}

	}

	// Cross fingers
	// Read the number of bytes specified by length and hope it doesnt go out of sync
	msgbuf := make([]byte, 40)
	for i := 0; i < 40; i++ {
        // conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
		response, err := reader.ReadByte()
		if err != nil {
			util.EPrintf("labtcp ReadHandshake: %s\n", err.Error())
			responseData := append(msgLength, zerobuf...)
			responseData = append(responseData, msgbuf...)
            util.EPrintf("Failed Handshake\n")
			return responseData, err
			// break
		}
		msgbuf[i] = response
	}

	responseData := append(msgLength, pstrbuf...)
	responseData = append(responseData, zerobuf...)
	responseData = append(responseData, msgbuf...)
    util.TPrintf("Finished Handshake from: %s\n", conn.RemoteAddr().String())
	return responseData, nil
}

func ReadMessage(conn *net.TCPConn) ([]byte, error) {
	// General strategy for reading packets back
	// 1) The first four bytes for the length of the packets
	// 2) Read that many bytes after to form a packet
	// 3) Hope that nothing goes out of sync
	// 4) Check that the packets are reasonable
	// 5) repeat 3
    util.TPrintf("Reading Message from:%s\n", conn.RemoteAddr().String())
    msgLength := make([]byte, 4)
    _, err := io.ReadFull(conn, msgLength)
    if err != nil {
        if err == io.EOF {
            return []byte{}, nil
        }
        return []byte{}, err
    }
    length := binary.BigEndian.Uint32(msgLength)
    msg := make([]byte, length)
    _, err = io.ReadFull(conn, msg)
    if err != nil {
        return []byte{}, err
    }
    util.TPrintf("Finished Reading Message from:%s\n", conn.RemoteAddr().String())
    return append(msgLength, msg...), nil
}

// TODO: This doesnt work... wtf
func IsConnectionClosed(conn *net.TCPConn) bool {
	one := []byte{}
	// conn.SetReadDeadline(time.Now())
	if data, err := conn.Read(one); err == io.EOF {
		util.ColorPrintf(util.Purple, "Closing the connection in labtcp/IsConnectionClosed\n")
		conn.Close()
		// conn = nil
		return true
	} else {
		// var zero time.Time
		util.TPrintf("This is the data: %v, %v\n", data, err)
		conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		return false
	}
}
