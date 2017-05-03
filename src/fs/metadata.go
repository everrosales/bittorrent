package fs

// For reading torrent metadata files

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/zeebo/bencode"
	"os"
	"strings"
	"util"
)

type Torrent struct {
	// according to bittorrent spec
	Announce string
	Info     map[string]interface{}
}

type Metadata struct {
	// easier internal representation to use
	TrackerUrl  string
	Name        string
	PieceLen    int64
	PieceHashes []string
	Files       []FileData
}

type FileData struct {
	Length int64
	Path   []string
}

// Open a .torrent file and decodne its contents
func ReadTorrent(path string) Torrent {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	dec := bencode.NewDecoder(file)
	torrent := Torrent{}
	dec.Decode(&torrent)
	if len(torrent.Info) == 0 {
		// empty torrent
		panic("Torrent " + path + " was empty or not decoded properly")
	}
	file.Close()
	return torrent
}

// Get escaped string of SHA1 hash of torrent's info field
func GetInfoHash(torrent Torrent) string {
	bytes := GetBytes(torrent.Info)
	h := sha1.New()
	h.Write(bytes)
	sha := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return sha
}

// Given a file path, read the info field into a Metadata struct
func Read(path string) Metadata {
	torrent := ReadTorrent(path)
	metadata := Metadata{}
	metadata.TrackerUrl = torrent.Announce
	metadata.Name = torrent.Info["name"].(string)
	metadata.PieceLen, _ = torrent.Info["piece length"].(int64)
	metadata.PieceHashes = util.SplitEveryN(torrent.Info["pieces"].(string), 20)
	if _, ok := torrent.Info["length"]; ok {
		// single file
		metadata.Files = []FileData{
			FileData{
				Length: torrent.Info["length"].(int64),
				Path:   []string{}}}
	} else {
		panic("currently no support for multiple files in a torrent file")
		// // multiple files
		// metadata.Files = []FileData{}
		// obj := []interface{}{}
		// files := Decode(torrent.Info["files"], obj)
		// fmt.Println("%v", files)
		// // for _, file := range files {
		// //     fmt.Println("%T", file)
		// //     // metadata.Files = append(metadata.Files, FileData{
		// //     //     Length: file["length"],
		// //     //     Path: file["path"] })
		// // }
	}
	return metadata
}

// Write torrent metadata into a .torrent file
func Write(path string, data Metadata) {
	torrent := Torrent{}
	torrent.Announce = data.TrackerUrl
	torrent.Info = make(map[string]interface{})
	torrent.Info["name"] = data.Name
	torrent.Info["piece length"] = data.PieceLen
	torrent.Info["pieces"] = strings.Join(data.PieceHashes, "")
	if len(data.Files) == 1 {
		torrent.Info["length"] = data.Files[0].Length
	} else {
		panic("currently no support for multiple files in a torrent file")
		// // multiple files
		// files := []map[string]string{}
		// for _, file := range data.Files {
		//     newFile := map[string]string{
		//         "length": strconv.FormatUint(file.Length, 10),
		//         "path": Encode(file.Path) }
		//     files = append(files, newFile)
		// }
		// torrent.Info["files"] = Encode(files)
	}
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	enc := bencode.NewEncoder(file)
	err = enc.Encode(torrent)
	if err != nil {
		panic(err)
	}
	file.Close()
}

func main() {
	file := FileData{Length: 1234}
	Write("test.torrent", Metadata{"blah", "blah", 1, []string{"aaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbb"}, []FileData{file}})
	fmt.Println("%v", Read("test.torrent"))
}
