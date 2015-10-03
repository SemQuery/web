package query

import (
    "github.com/semquery/web/app/common"

    "github.com/martini-contrib/render"
    "github.com/gorilla/websocket"

    "os"
    "math/rand"
    "log"
    "bufio"
    "strconv"
    "net/http"
    "strings"
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


    path := "_repos/" + req.FormValue("repo")

    if _, err := os.Stat(path); os.IsNotExist(err) {
        data["indexed"] = false
    } else {
        data["indexed"] = true
    }

    data["query"] = req.FormValue("q")

    r.HTML(200, "query", data)
}

func SocketPage(user common.User, r *http.Request, w http.ResponseWriter) {
    ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
    if _, ok := err.(websocket.HandshakeError); (ok || err != nil) {
        log.Fatal(err)
        return
    }

    _, msg, err := ws.ReadMessage()
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
    executable := common.Config.EngineExecutable

    if _, err := os.Stat(path); os.IsNotExist(err) && user.IsLoggedIn() {

        repository, _, e := user.Github().Repositories.Get(strings.Split(repo, "/")[0], strings.Split(repo, "/")[1])

        if e != nil {
            ws.WriteMessage(1, []byte("!This repository was not found"))
            ws.Close()
            return
        }

        limit := 1000000
        if *repository.Size > limit {
            ws.WriteMessage(1, []byte("!This repository exceeds the size limit"))
            ws.Close()
            return
        }

        os.MkdirAll(path, 0777)
        c := exec.Command("git", "clone", "https://github.com/" + repo + ".git", path)
        c.Run()
        c.Wait()

        cmd := exec.Command("java", "-jar", executable, "index", path, repo)

        cmdReader, _ := cmd.StdoutPipe()

        scanner := bufio.NewScanner(cmdReader)

        go func() {
            cmd.Start()
            for scanner.Scan() {
                ws.WriteMessage(1, []byte(scanner.Text()))
            }
        }()
        cmd.Wait()
    } else if !user.IsLoggedIn() {
        ws.WriteMessage(1, []byte("!You must be logged in order to index a respository"))
        ws.Close()
        return
    }

    cmd := exec.Command("java", "-jar", executable, "query", query, repo)

    cmdReader, _ := cmd.StdoutPipe()

    scanner := bufio.NewScanner(cmdReader)

    go func() {
        cmd.Start()
        for scanner.Scan() {
            text := scanner.Text()
            parts := strings.Split(text, ",")
            if len(parts) == 1 {
                ws.WriteMessage(1, []byte("#" + parts[0]))
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
            ws.WriteMessage(1, []byte(jstr))
        }
    }()
    cmd.Wait()

    log.Print("Finished indexing.")
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
