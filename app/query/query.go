package query

import (
    "github.com/semquery/web/app/common"

    "gopkg.in/mgo.v2/bson"

    "github.com/martini-contrib/render"
    "github.com/martini-contrib/sessions"
    "github.com/gorilla/websocket"

    "os"
    "math/rand"
    "strconv"
    "net/http"
    "net/url"
    "strings"
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

type Packet struct {
    Action string `json:"action"`
    Payload map[string]interface{} `json:"payload"`
}

func (p Packet) Send(ws *websocket.Conn) {
    raw, _ := json.Marshal(p)
    ws.WriteMessage(1, raw)
}

func SocketPage(user common.User, session sessions.Session, r *http.Request, w http.ResponseWriter) {
    ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
    if _, ok := err.(websocket.HandshakeError); (ok || err != nil) {
        return
    }

    _, msg, err := ws.ReadMessage()
    if err != nil {
        return
    }

    id, err := strconv.ParseInt(string(msg), 10, 64)
    if err != nil {
        return
    }

    arr := ws_transfer[id]
    query := arr[0]
    repo := arr[1]
    delete(ws_transfer, id)

    repo_parts := strings.Split(repo, "/")
    if len(repo_parts) < 2 {
        return
    }

    var store bson.M
    err = common.Database.C("repositories").Find(bson.M {
       "repository": repo,
    }).One(&store)

    if err != nil {
        if user.IsLoggedIn() {
            repository, _, e := user.Github().Repositories.Get(strings.Split(repo, "/")[0], strings.Split(repo, "/")[1])

            if e != nil {
                Packet {
                    Action: "warning",
                    Payload: map[string]interface{} {
                        "message": "This repository was not found",
                    },
                }.Send(ws)
                ws.Close()
                return
            }

            limit := 1000000
            if *repository.Size > limit {
                Packet {
                    Action: "warning",
                    Payload: map[string]interface{} {
                        "message": "This repository exceeds the size limit",
                    },
                }.Send(ws)
                ws.Close()
                return
            }

            doc := bson.M {
                "repository": repo,
                "status": "standby",
            }
            common.Database.C("repostiories").Insert(doc)

            indexJob := IndexingJob {
                Token: session.Get("token").(string),
                RepositoryPath: repo,
            }

            QueueIndexingJob(indexJob)

            progress := Packet { "", map[string]interface{} {} }
            pubsub, _ := common.Rds.Subscribe(repo)
            for {
                msg, err := pubsub.ReceiveMessage()
                if err != nil {
                    continue
                }
                progress = Packet {}
                json.Unmarshal([]byte(msg.Payload), &progress)
                if progress.Action == "finished" {
                    find := bson.M {
                        "repository": repo,
                    }
                    update := bson.M {
                        "status": "completed",
                    }
                    common.Database.C("repositories").Update(find, update)
                    break;
                } else {
                    ws.WriteMessage(1, []byte(msg.Payload))
                }
            }
            pubsub.Close()
        } else if !user.IsLoggedIn() {
            Packet {
                Action: "warning",
                Payload: map[string]interface{} {
                    "message": "You must be logged in order to index a repository",
                },
            }.Send(ws)
            ws.Close()
            return
        }
    }

    letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

    b := make([]rune, 10)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }

    common.Rds.Set(query + "|" + repo, string(b), 0)

    form := url.Values {}
    form.Add("token", string(b))
    form.Add("repo", repo)
    form.Add("query", query)

    client := &http.Client{}
    resp, _ := client.PostForm("http://localhost:3001/", form)

    body, _ := ioutil.ReadAll(resp.Body)

    ws.WriteMessage(1, body)
    ws.Close()
}
