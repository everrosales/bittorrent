package fs
// For reading torrent metadata files

import "os"
import "strconv"
import "github.com/zeebo/bencode"
import "strings"
import "bytes"
import "fmt"

type Metadata struct {
    // internal representation
    TrackerUrl string
    Name string
    PieceLen uint64
    PieceHashes []string
    Files []FileData
}

type FileData struct {
    Length uint64
    Path []string
}

type TorrentFile struct {
    // according to bittorrent spec
    Announce string
    Info map[string]string
}

func read(path string) Metadata {
    file, err := os.Open(path)
    if err != nil {
        panic(err)
    }
    dec := bencode.NewDecoder(file)
    torrent := TorrentFile{}
    dec.Decode(&torrent)
    file.Close()

    metadata := Metadata{}
    metadata.TrackerUrl = torrent.Announce
    metadata.Name = torrent.Info["name"]
    metadata.PieceLen, _ = strconv.ParseUint(torrent.Info["piece length"], 0, 64)
    metadata.PieceHashes = SplitEveryN(torrent.Info["pieces"], 20)
    if _, ok := torrent.Info["length"]; ok {
        // single file
        length, _ := strconv.ParseUint(torrent.Info["length"], 0, 64)
        metadata.Files = []FileData{
            FileData{
                Length: length,
                Path: []string{} } }
    } else {
        panic("currently no support for multiple files in a torrent file")
        // // multiple files
        // metadata.Files = []FileData{}
        // obj := []interface{}{}
        // files := decodeFromString(torrent.Info["files"], obj)
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

func write(path string, data Metadata) {
    torrent := TorrentFile{}
    torrent.Announce = data.TrackerUrl
    torrent.Info = make(map[string]string)
    torrent.Info["name"] = data.Name
    torrent.Info["piece length"] = strconv.FormatUint(data.PieceLen, 10)
    torrent.Info["pieces"] = strings.Join(data.PieceHashes, "")
    if len(data.Files) == 1 {
        torrent.Info["length"] = strconv.FormatUint(data.Files[0].Length, 10)
    } else {
        panic("currently no support for multiple files in a torrent file")
        // // multiple files
        // files := []map[string]string{}
        // for _, file := range data.Files {
        //     newFile := map[string]string{
        //         "length": strconv.FormatUint(file.Length, 10),
        //         "path": encodeToString(file.Path) }
        //     files = append(files, newFile)
        // }
        // torrent.Info["files"] = encodeToString(files)
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
func SplitEveryN(s string, n int) []string {
    sub := ""
    subs := []string{}

    runes := bytes.Runes([]byte(s))
    l := len(runes)
    for i, r := range runes {
        sub = sub + string(r)
        if (i + 1) % n == 0 {
            subs = append(subs, sub)
            sub = ""
        } else if (i + 1) == l {
            subs = append(subs, sub)
        }
    }

    return subs
}

func main() {
    file := FileData{Length:1234}
    write("test.torrent", Metadata{"blah", "blah", 1, []string{"aaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbb"}, []FileData{file}})
    fmt.Println("%v", read("test.torrent"))

    // write("test2.torrent", Metadata{"blah", "blah", 1, []string{"aaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbb"}, []FileData{file, file, file}})
    // fmt.Println("%v", read("test2.torrent"))
}