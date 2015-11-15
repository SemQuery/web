package query

import (
    "github.com/semquery/web/app/common"


    "github.com/google/go-github/github"
    "github.com/martini-contrib/render"
    "github.com/martini-contrib/sessions"
    "github.com/gorilla/websocket"

//    "log"
    "math/rand"
//    "strconv"
    "net/http"
    "net/url"
//    "strings"
    "encoding/json"
//    "io/ioutil"

    "errors"
)

func SearchPage(user common.User, session sessions.Session, r render.Render, req *http.Request) {
    src, err := handleSearch(user, req)

    if err != nil {
        r.Error(400)
    }

    status := common.GetCodeSourceStatus(src)

    id := rand.Int63()
    params := req.URL.Query()
    usr := params.Get("user")
    repo := params.Get("repo")
    session.Set(id, usr + "/" + repo)

    data := struct {
        common.User
        Pagename     string
        SourceStatus string
        WS_ID int64
    } {user, "search", string(status), id}

    r.HTML(200, "search", data)
}

// Creates a CodeSource from the URL query string,
// returning (source, nil) successful or (nil, error)
// if a code source could not be created
func handleSearch(user common.User, req *http.Request) (common.CodeSource, error) {
    params := req.URL.Query()
    source := params.Get("source")

    switch source {
    case common.CodeSourceGitHub:
        user := params.Get("user")
        repo := params.Get("repo")
        if user == "" || repo == "" {
            return nil, errors.New("User or repository blank")
        }
        return &common.RepositorySource{user, repo}, nil

    case common.CodeSourceLink:
        link     := params.Get("link")
        url, err := url.Parse(link)
        if err != nil {
            return nil, errors.New("Invalid link")
        }
        return &common.LinkSource{url}, nil

    default:
        return nil, errors.New("Invalid source")
    }
}

type Packet struct {
    Action string `json:"action"`
    Payload map[string]interface{} `json:"payload"`
}

func (p Packet) Json() string {
    raw, _ := json.Marshal(p)
    return string(raw)
}

func (p Packet) Send(ws *websocket.Conn) {
    ws.WriteMessage(1, []byte(p.Json()))
}

func InitiateIndex(user common.User, r *http.Request, session sessions.Session) string {
    if !user.IsLoggedIn() {
        return Packet {
            Action: "warning",
            Payload: map[string]interface{} {
                "message": "You are not logged in",
            },
        }.Json()
    }

    r.ParseForm()
    params, _ := url.ParseQuery(r.FormValue("search"))
    if params.Get("user") == "" || params.Get("repo") == "" {
        return Packet {
            Action: "warning",
            Payload: map[string]interface{} {
                "message": "Null repository owner or name",
            },
        }.Json()
    }

    repo := &common.RepositorySource {
        User: params.Get("user"),
        Name: params.Get("repo"),
    }

    if common.GetCodeSourceStatus(repo) != common.CodeSourceStatusNotFound {
        return Packet {
            Action: "warning",
            Payload: map[string]interface{} {
                "message": "This repository is either already indexed or is currently being indexed",
            },
        }.Json()
    }

    _, _, e := user.Github().Repositories.Get(repo.User, repo.Name)

    if e != nil {
        return Packet {
            Action: "warning",
            Payload: map[string]interface{} {
                "message": "This repository does not exist",
            },
        }.Json()
    }

    _, _, e = github.NewClient(nil).Repositories.Get(repo.User, repo.Name)

    if e != nil && user.GetPlan()["name"] == "normal" {
        return Packet {
            Action: "warning",
            Payload: map[string]interface{} {
                "message": "You must be on a paid plan in order to index a private repository",
            },
        }.Json()
    }

    indexJob := IndexingJob {
        Token: session.Get("token").(string),
        RepositoryPath: repo.User + "/" + repo.Name,
    }

    common.UpdateStatus(repo, common.CodeSourceStatusWorking)
    QueueIndexingJob(indexJob)
    go func() {
        progress := Packet { "", map[string]interface{} {} }
        pubsub, _ := common.Rds.Subscribe(indexJob.RepositoryPath)
        defer pubsub.Close()
        for {
            msg, err := pubsub.ReceiveMessage()
            if err != nil {
                continue
            }
            progress = Packet {}
            json.Unmarshal([]byte(msg.Payload), &progress)
            if progress.Action == "finished" {
                common.UpdateStatus(repo, common.CodeSourceStatusDone)
            }
        }
    }()
    return Packet {
        Action: "success",
        Payload: map[string]interface{} {},
    }.Json()
}

func SocketPage(user common.User, session sessions.Session, r *http.Request, w http.ResponseWriter) {

    params := r.URL.Query()
    if params.Get("user") == "" || params.Get("repo") == "" {
        return
    }

    repo := &common.RepositorySource {
        User: params.Get("user"),
        Name: params.Get("repo"),
    }

    if common.GetCodeSourceStatus(repo) != common.CodeSourceStatusWorking {
        return
    }

    ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
    if _, ok := err.(websocket.HandshakeError); (ok || err != nil) {
        return
    }

    progress := Packet { "", map[string]interface{} {} }
    pubsub, _ := common.Rds.Subscribe(repo.User + "/" + repo.Name)
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
            user.AddIndexed(repo.User + "/" + repo.Name)
            break;
        }
    }
    ws.Close()
}
