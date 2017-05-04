package fs

// For reading torrent metadata files

import (
	"crypto/sha1"
	//"encoding/base64"
	"io/ioutil"
	"net/url"
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
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	torrent := Torrent{}
	Decode(string(fileBytes), &torrent)
	if len(torrent.Info) == 0 {
		// empty torrent
		panic("Torrent " + path + " was empty or not decoded properly")
	}
	return torrent
}

// Get escaped string of SHA1 hash of torrent's info field
func GetInfoHash(torrent Torrent) string {
	bencodedStr := Encode(torrent.Info)
	sha := sha1.Sum([]byte(bencodedStr))
	n := len(sha)
	if n != 20 {
		panic("SHA hash generation failed")
	}
	shaStr := string(sha[:n])
	return url.QueryEscape(shaStr)
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
	outBytes := []byte(Encode(torrent))
	err := ioutil.WriteFile(path, outBytes, 0644)
	if err != nil {
		panic(err)
	}
}
