package fs

// Encoder helpers

import (
	"bytes"
	//"encoding/gob"
	"github.com/zeebo/bencode"
)

// encode interface into a bencoded string
/*func GetBytes(data interface{}) []byte {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(data)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}*/

// encode interface to bencoded string
func Encode(obj interface{}) string {
	var buf bytes.Buffer
	enc := bencode.NewEncoder(&buf)
	err := enc.Encode(obj)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

// decode string to interface
func Decode(data string, obj interface{}) {
	var buf bytes.Buffer
	buf.WriteString(data)
	dec := bencode.NewDecoder(&buf)
	dec.Decode(&obj)
}
