package query

import (
    "github.com/semquery/web/app/common"


    // "github.com/google/go-github/github"
    "github.com/martini-contrib/render"
    "github.com/martini-contrib/sessions"
    "github.com/gorilla/websocket"

    "gopkg.in/mgo.v2/bson"

    "math/rand"
    "net/http"
    "net/url"
    "encoding/json"
    "os/exec"
    "time"

    "errors"
)

type wsConn struct {
    subscribed bool
    conn *websocket.Conn
}

var wsMap map[string][]wsConn = map[string][]wsConn {}

func SearchPage(user common.User, session sessions.Session, r render.Render, req *http.Request) {
    src, err := handleSearch(req.URL.Query())

    if err != nil {
        r.Error(400)
    }

    status := common.GetCodeSourceStatus(src)

    id     := rand.Int63()
    params := req.URL.Query()
    usr    := params.Get("user")
    repo   := params.Get("repo")
    session.Set(id, usr + "/" + repo)

    data := struct {
        common.User
        Pagename     string
        Theme        string
        SourceStatus string
        WS_ID int64

        Source *common.CodeSource
    } {user, "search", "standard", string(status), id, src}

    r.HTML(200, "search", data)
}

// Creates a CodeSource from the URL query string,
// returning (source, nil) successful or (nil, error)
// if a code source could not be created
func handleSearch(params url.Values) (*common.CodeSource, error) {
    source := params.Get("source")

    switch source {
    case common.CodeSourceGitHub:
        user := params.Get("user")
        repo := params.Get("repo")
        return common.CreateGitHubSource(user, repo), nil

    case common.CodeSourceLink:
        link     := params.Get("link")
        url, err := url.Parse(link)
        if err != nil {
            return nil, errors.New("Invalid link")
        }
        return common.CreateGitSource(url), nil

    default:
        return nil, errors.New("Invalid source")
    }
}

type Packet struct {
    Action string `json:"action"`
    Payload map[string]interface{} `json:"payload"`
}

func WarningPacket(message string) Packet {
    return Packet {
        Action: "warning",
        Payload: map[string]interface{} {
            "message": message,
        },
    }
}

func (p Packet) Json() string {
    raw, _ := json.Marshal(p)
    return string(raw)
}

func (p Packet) Send(ws *websocket.Conn) {
    ws.WriteMessage(1, []byte(p.Json()))
}

func InitiateIndex(user common.User, r *http.Request,
    session sessions.Session) (string, int) {

    if !user.IsLoggedIn() {
        return WarningPacket("You are not logged in").Json(),
            http.StatusForbidden
    }

    r.ParseForm()
    if r.FormValue("search") == "" {
        return WarningPacket("Null information").Json(),
            http.StatusBadRequest
    }

    params, err := url.ParseQuery(r.FormValue("search")[1:])
    if err != nil {
        return WarningPacket("Invalid search").Json(),
            http.StatusBadRequest
    }
    src, err := handleSearch(params)
    if err != nil {
        return WarningPacket("Invalid search").Json(),
            http.StatusBadRequest
    }

    if src.Git != nil {
        return IndexGit(src, session)
    }

    return WarningPacket("No source found").Json(),
        http.StatusBadRequest
}

func IndexGit(src *common.CodeSource, s sessions.Session) (string, int) {
    git := src.Git

    err := validateGitURL(git.URL.String())
    if err != nil {
        return WarningPacket(err.Error()).Json(),
            http.StatusNotFound
    }

    id, err := common.InsertSource(src, common.CodeSourceStatusWorking)
    if err != nil {
        return WarningPacket("Internal error").Json(),
            http.StatusInternalServerError
    }

    job := &IndexingJob{
        URL: git.URL.String(),
        ID: id.Hex(),
        Type: GitURLIndexingJob,
    }
    if git.IsGitHub() {
        job.Type  = GitHubIndexingJob
        job.Token = s.Get("token").(string)
    }

    if !QueueIndexingJob(job) {
        return WarningPacket("Unable to queue").Json(),
            http.StatusInternalServerError
    }

    go redisPubSub(id)

    res := Packet{
        Action: "queued",
        Payload: map[string]interface{} {
            "id": id.Hex(),
        },
    }
    return res.Json(), http.StatusOK
}

func redisPubSub(id bson.ObjectId) {
    pubsub, _ := common.Rds.Subscribe("indexing:" + id.Hex())
    time.Sleep(time.Second)
    defer pubsub.Close()
    for {
        msg, err := pubsub.ReceiveMessage()
        if err != nil {
            continue
        }

        data := Packet {}
        err  = json.Unmarshal([]byte(msg.Payload), &data)
        if err == nil {
            if data.Action == "finished" {
                common.UpdateStatus(id, common.CodeSourceStatusDone)
            }
            if clients, ok := wsMap[id.Hex()]; ok {
                for _, c := range clients {
                    data.Send(c.conn)
                }
            }
        }
    }
}

// Validates a git URL
// returns:
//   * nil if valid
//   * error if not valid or if command execution
//     exceeds 10 seconds.
func validateGitURL(url string) error {
    cmd := exec.Command("git", "ls-remote", url)
    cmd.Start()
    done := make(chan error, 1)
    go func () {
        done <- cmd.Wait()
    }()
    select {
        case <-time.After(10 * time.Second):
            cmd.Process.Kill()
            <-done
            return errors.New("Git took too long")
        case err := <-done:
            if err != nil {
                if !cmd.ProcessState.Success() {
                    return errors.New("Invalid Git URL")
                }
                return errors.New("Git command failed")
            }
            return nil
    }
}

func SocketPage(r *http.Request, w http.ResponseWriter) {
    ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
    if err != nil {
        return
    }

    subscribeMsg := map[string]interface{} {}
    err = ws.ReadJSON(&subscribeMsg)

    if err != nil {
        ws.Close()
        return
    }

    id, ok := subscribeMsg["id"]
    if ok {
        if val, ok := id.(string); ok {
            addClient(val, ws)
            return
        }
    }
    ws.Close()
}

func addClient(id string, ws *websocket.Conn) {
    if arr, ok := wsMap[id]; ok {
        wsMap[id] = append(arr, wsConn{true, ws})
    } else {
        wsMap[id] = []wsConn{
            wsConn{true, ws},
        }
    }
}

/*
func SocketPage(user common.User, session sessions.Session, r *http.Request, w http.ResponseWriter) {

    params := r.URL.Query()

    if params.Get("source") == "" {
        return
    }

    var name string

    if params.Get("source") == "github" {
        if params.Get("user") == "" || params.Get("repo") == "" {
            return
        }
        source := &common.RepositorySource {
            User: params.Get("user"),
            Name: params.Get("repo"),
        }
        name = source.User + "/" + source.Name
        if common.GetCodeSourceStatus(source) != common.CodeSourceStatusWorking {
            return
        }
    } else if params.Get("source") == "link" {
        if params.Get("link") == "" {
            return
        }
        url, _ := url.Parse(params.Get("link"))
        source := &common.LinkSource {
           URL: url,
        }
        name = source.URL.String()
        if common.GetCodeSourceStatus(source) != common.CodeSourceStatusWorking {
            return
        }
    }

    log.Println("Connected to web socket")
    log.Println(name)

    ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
    if _, ok := err.(websocket.HandshakeError); (ok || err != nil) {
        return
    }

    progress := Packet { "", map[string]interface{} {} }
    pubsub, _ := common.Rds.Subscribe(name)
    defer pubsub.Close()
    for {
        msg, err := pubsub.ReceiveMessage()
        if err != nil {
            continue
        }

        progress = Packet {}
        json.Unmarshal([]byte(msg.Payload), &progress)
        progress.Send(ws)
        if progress.Action == "finished" {
            user.AddIndexed(name)
            break;
        }
    }
    ws.Close()
}
*/
