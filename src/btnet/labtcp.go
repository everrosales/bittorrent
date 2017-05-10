package btnet

import (
	// "bufio"
	// "bytes"
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
    msgLength := make([]byte, 1)
    _, err := io.ReadFull(conn, msgLength)
    if err != nil {
        if err == io.EOF {
            return []byte{}, nil
        }
        return []byte{}, err
    }
    length := int(msgLength[0]) + 48
    msg := make([]byte, length)
    _, err = io.ReadFull(conn, msg)
    if err != nil {
        return []byte{}, err
    }
    util.TPrintf("Finished Handshake from: %s\n", conn.RemoteAddr().String())
	return append(msgLength, msg...), nil
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
