package btnet
// labtcp

import "net"
import "fmt"
import "bufio"
import "encoding/binary"
import "bytes"
// import "time"

func StartTCPServer(addr string, handler func(net.Conn)) {
  fmt.Println("Starting the TCP Server")
  ln, err := net.Listen("tcp", addr)
  if err != nil {
    // complain about things dying
    fmt.Println(err)
    // fmt.Println("Something went wrong")
  }
  go func(ln net.Listener) {
    conn, err := ln.Accept()
    if err != nil {
      // complain about a thing
      fmt.Println(err)
    }
    go handler(conn)
  }(ln)
}


func DoDial(addr *net.TCPAddr, data []byte) *net.TCPConn {
  conn, err := net.DialTCP("tcp", nil, addr)
  if err != nil {
    // Cry
    fmt.Println(err)
    return conn
  }
  fmt.Println("Writing out to connection")
  conn.Write(data)
  // fmt.Println("Preparing bufio for response")
  return conn
  // return ReadMessage(conn)
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
  for i := 0 ; i < 4; i++ {
    response, err := reader.ReadByte()
    if err != nil {
      fmt.Println("conn err: ", err)
    }
    msgLength[i] = response
  }

  fmt.Println(msgLength)
  // Parse the length of the message
  var length int32
  msgLengthDecodeBuf := bytes.NewReader(msgLength)
  errBinary := binary.Read(msgLengthDecodeBuf, binary.BigEndian, &length)
  if errBinary != nil {
    fmt.Println("binary.Read failed: ", errBinary)
  }
  fmt.Println("length: ", length)

  // Cross fingers
  // Read the number of bytes specified by length and hope it doesnt go out of sync
  msgbuf := make([]byte, length)
  for i := 0; i < int(length); i++ {
    response, err := reader.ReadByte()
    if err != nil {
      fmt.Println("conn err: ", err)
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
