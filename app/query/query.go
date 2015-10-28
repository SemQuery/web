package query

import (
    "github.com/semquery/web/app/common"

    "gopkg.in/mgo.v2/bson"

    "github.com/martini-contrib/render"
    "github.com/martini-contrib/sessions"
    "github.com/gorilla/websocket"

    "os"
    "math/rand"
    "log"
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
    Action string
    Payload map[string]interface{}
}

func SocketPage(user common.User, session sessions.Session, r *http.Request, w http.ResponseWriter) {
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

    var store bson.M
    err = common.Database.C("Repositories").Find(bson.M {
       "repository": repo,
    }).One(&store)

    if err != nil {
        if user.IsLoggedIn() {
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
                json.Unmarshal([]byte(msg.Payload), progress)
                if progress.Action == "Finished" {
                    find := bson.M {
                        "repository": repo,
                    }
                    update := bson.M {
                        "status": "completed",
                    }
                    common.Database.C("repositories").Update(find, update)
                    break;
                }
                ws.WriteMessage(1, []byte(msg.Payload))
            }
            pubsub.Close()
        } else if !user.IsLoggedIn() {
            ws.WriteMessage(1, []byte("!You must be logged in order to index a respository"))
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

    req, _ := http.NewRequest("POST", "", strings.NewReader(form.Encode()))
    client := http.Client{}
    resp, _ := client.Do(req)

    body, _ := ioutil.ReadAll(resp.Body)

    ws.WriteMessage(1, body)
    ws.Close()
}
