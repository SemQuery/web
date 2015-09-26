package query

import (
    "github.com/semquery/web/app/common"

    "github.com/martini-contrib/render"
    "github.com/gorilla/websocket"

    "os"
    "math/rand"
    "log"
    "net"
    "net/http"
    "bufio"
    "strconv"
    "strings"
    "sync"
    "os/exec"
    "encoding/json"
    "io/ioutil"
)

var ws_transfer = map[int64][]string{}

//Rendering search page with template data
func QueryPage(user common.User, r render.Render, req *http.Request) {
    data := common.CreateData(user, nil)

    req.ParseForm()
    id := rand.Int63()
    data["ws_id"] = id
    ws_transfer[id] = []string{req.FormValue("q"), req.FormValue("repo")}
    log.Println("Hello", req.FormValue("q"), req.FormValue("repo"))

    file1 := []string{"foo", "bar", "baz"}
    file2 := []string{"a", "b", "c"}
    files := [][]string{file1, file2}

    data["files"] = files

    path := "_repos/" + req.FormValue("repo")

    if _, err := os.Stat(path); os.IsNotExist(err) {
        data["indexed"] = false
    } else {
        data["indexed"] = true
    }

    data["query"] = req.FormValue("q")

    r.HTML(200, "query", data)
}

var ActiveClients = map[ClientConn]int {}
var ActiveClientsRWMutex sync.RWMutex

type ClientConn struct {
    websocket *websocket.Conn
    clientIP net.Addr
}

func SocketPage(r *http.Request, w http.ResponseWriter) {
    ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
    if _, ok := err.(websocket.HandshakeError); (ok || err != nil) {
        log.Fatal(err)
        return
    }
    //Initial connection, store
    client := ws.RemoteAddr()
    sockCli := ClientConn {ws, client}
    ActiveClientsRWMutex.Lock()
    ActiveClients[sockCli] = 0
    ActiveClientsRWMutex.Unlock()

    log.Print("Starting")
    _, msg, err := sockCli.websocket.ReadMessage()
    if err != nil {
        log.Print(err)
        return
    }

    id, err := strconv.ParseInt(string(msg), 10, 64)
    if err != nil {
        log.Print(err)
        return
    }

    arr := ws_transfer[id]
    query := arr[0]
    repo := arr[1]
    delete(ws_transfer, id)

    repo_parts := strings.Split(repo, "/")
    if len(repo_parts) < 2 {
        log.Print("Invalid repo")
        return
    }

    path := "_repos/" + repo

    if _, err := os.Stat(path); os.IsNotExist(err) {
        os.MkdirAll(path, 0777)
        c := exec.Command("git", "clone", "https://github.com/" + repo + ".git", path)
        c.Run()
        c.Wait()

        cmd := exec.Command("java", "-jar", "/Users/August/Code/projects/semquery/engine/target/engine-1.0-SNAPSHOT.jar", "index", path, repo)

        cmdReader, _ := cmd.StdoutPipe()

        scanner := bufio.NewScanner(cmdReader)

        go func() {
            cmd.Start()
            for scanner.Scan() {
                sockCli.websocket.WriteMessage(1, []byte(scanner.Text()))
            }
        }()
        cmd.Wait()
    }


    cmd := exec.Command("java", "-jar", "/Users/August/Code/projects/semquery/engine/target/engine-1.0-SNAPSHOT.jar", "query", query, repo)

    cmdReader, _ := cmd.StdoutPipe()

    scanner := bufio.NewScanner(cmdReader)

    go func() {
        cmd.Start()
        for scanner.Scan() {
            text := scanner.Text()
            parts := strings.Split(text, ",")
            if len(parts) == 1 {
                sockCli.websocket.WriteMessage(1, []byte("#" + parts[0]))
                continue
            }
            file := parts[0]
            src, _ := ioutil.ReadFile(file)
            start, _ := strconv.Atoi(parts[1])
            end, _ := strconv.Atoi(parts[2])
            lines, relStart, relEnd := extractLines(string(src), start, end)
            j := map[string]interface{}{}
            for k, v := range lines {
                j[strconv.Itoa(k)] = v
            }
            jstr, _ := json.Marshal(map[string]interface{}{
                "lines": j,
                "file": file,
                "relative_start": relStart,
                "relative_end": relEnd,
            })
            sockCli.websocket.WriteMessage(1, []byte(jstr))
        }
    }()
    cmd.Wait()

    log.Print("DONE WITH INDEXING!")
}

// Extracts the lines encapsulating characters in
// the range (start..end)
//
// returns: (line pairs, relative start position, relative end position)
func extractLines(src string, start int, end int) (map[int]string, int, int) {
    lines := map[int]string{}

    currentLine := 1
    lineStartPos := 0
    relativeStartPos := 0

    for i := 0; i < start; i++ {
        if src[i] == '\n' {
            currentLine++;
            lineStartPos = i + 1
        }
        if (i == start - 1) {
            relativeStartPos = i - lineStartPos + 1
        }
    }

    relativeEndPos := 0

    for i := start; i < len(src); i++ {
        if i == end {
            relativeEndPos = i - lineStartPos
        }
        if src[i] == '\n' || i == len(src) - 1 {
            sub := src[lineStartPos : i]
            lines[currentLine] = sub

            if len(lines) == 15 {
                return lines, relativeStartPos, relativeEndPos
            }

            currentLine += 1
            lineStartPos = i + 1

            if i >= end {
                break
            }
        }
    }

    return lines,relativeStartPos, relativeEndPos
}
