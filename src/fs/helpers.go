package fs

import (
	"bytes"
	"github.com/zeebo/bencode"
)

func encodeToString(obj interface{}) string {
	var buf bytes.Buffer
	enc := bencode.NewEncoder(&buf)
	err := enc.Encode(obj)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func decodeFromString(str string, obj interface{}) interface{} {
	var buf bytes.Buffer
	buf.WriteString(str)
	dec := bencode.NewDecoder(&buf)
	dec.Decode(&obj)
	return obj
}

// from http://stackoverflow.com/questions/25686109/split-string-by-length-in-golang
func splitEveryN(s string, n int) []string {
	sub := ""
	subs := []string{}

	runes := bytes.Runes([]byte(s))
	l := len(runes)
	for i, r := range runes {
		sub = sub + string(r)
		if (i+1)%n == 0 {
			subs = append(subs, sub)
			sub = ""
		} else if (i + 1) == l {
			subs = append(subs, sub)
		}
	}

	return subs
}
