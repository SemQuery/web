package query

import (
    "github.com/semquery/web/app/common"

    "gopkg.in/mgo.v2/bson"

    "github.com/martini-contrib/render"
    "github.com/martini-contrib/sessions"
    "github.com/gorilla/websocket"

    "log"
    "math/rand"
    "strconv"
    "net/http"
    "net/url"
    "strings"
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

var ws_transfer = map[int64][]string{}

//Rendering search page with template data
func QueryPage(user common.User, r render.Render, req *http.Request) {
    data := struct {
        Loggedin, Indexed bool
        Usrname, Query string
        Ws_id int64
    } {
        Loggedin: user.IsLoggedIn(),
        Usrname: user.Username(),
    }

    req.ParseForm()
    /*
    repoUser := req.FormValue("user")
    repoName := req.FormValue("name")
    status := common.RepositoryStatus(&common.Repository{
        User: repoUser,
        Name: repoName,
    })
    */

    // id := rand.Int63()
    // data["ws_id"] = id
    // ws_transfer[id] = []string{req.FormValue("q"), req.FormValue("repo")}

    // path := "_repos/" + req.FormValue("repo")

    // if _, err := os.Stat(path); os.IsNotExist(err) {
    //     data["indexed"] = false
    // } else {
    //     data["indexed"] = true
    // }

    data.Query = req.FormValue("q")

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

func InitiateIndex(r *http.Request, session sessions.Session) {
    r.ParseForm()
    id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
    repo := session.Get(id).(string)
    indexJob := IndexingJob {
        Token: session.Get("token").(string),
        RepositoryPath: repo,
    }

    QueueIndexingJob(indexJob)
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

    repo := session.Get(id).(string)
    log.Println(repo)

    repo_parts := strings.Split(repo, "/")
    if len(repo_parts) < 2 {
        return
    }

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
            find := bson.M { "repository": repo }
            update := bson.M { "$set": bson.M { "status": "completed" } }
            common.Database.C("repositories").Update(find, update)
            user.AddIndexed(repo)
            break;
        } else {
            ws.WriteMessage(1, []byte(msg.Payload))
        }
    }
    pubsub.Close()

    /*
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

    ws.WriteMessage(1, body) */
    ws.Close()
}
