package main

import (
    "os"
    "log"
    "strings"
    "io/ioutil"
    "math/rand"
    "os/exec"
    "strconv"
    "net"
    "sync"
    "encoding/json"
    "bufio"
    "net/http"
    "net/url"

    "golang.org/x/crypto/bcrypt"

    "github.com/go-martini/martini"
    "github.com/gorilla/websocket"
    "github.com/martini-contrib/render"
    "github.com/martini-contrib/sessions"

    "gopkg.in/mgo.v2/bson"
)

func RegisterHandlers(m *martini.ClassicMartini) {
    m.Get("/", RootPage)
    m.Get("/repo", CacheRepository)
    m.Get("/socket", SocketPage)
    m.Post("/query", QueryPage)

    m.Get("/login", AlreadyLogRedirect, LoginPage)
    m.Post("/login", AlreadyLogRedirect, LoginAction)
    m.Get("/register", AlreadyLogRedirect, RegisterPage)
    m.Post("/register", AlreadyLogRedirect, RegisterAction)
    m.Get("/logout", LogoutAction)

    m.Get("/githubauth", AlreadyLogRedirect, GithubAuth)
    m.Get("/githubcallback", AlreadyLogRedirect, GithubCallback)

    m.Get("/cextension", ExtensionPage)
}

func CreateData(user User, session sessions.Session) map[string]interface{} {
    data := map[string]interface{} {
        "loggedin": strconv.FormatBool(user.IsLoggedIn()),
        "message": "",
    }
    if user.IsLoggedIn() {
        data["username"] = user.Username()
    }
    if session != nil {
        flashes := session.Flashes("message")
        if len(flashes) != 0 {
            data["message"] = flashes[0].(string)
        }
    }
    return data
}

//Rendering home page with template data
func RootPage(user User, r render.Render) {
    data := CreateData(user, nil)
    r.HTML(200, "index", data)
}

//Retrieves github repository to prepare to be indexed and searched
func CacheRepository(user User, req *http.Request, ren render.Render) {
    if user.IsLoggedIn() {
        query := req.URL.Query().Get("query")
        if query != "" {
            if _, err := os.Stat(strings.Split(query, "/")[1]); os.IsNotExist(err) {
                exec.Command("git", "clone", "https://github.com/" + query + ".git").Run()
            }
        }
    } else {
        ren.Redirect("/")
    }
}

var ws_transfer = map[int64][]string{}

//Rendering search page with template data
func QueryPage(user User, r render.Render, req *http.Request) {
    data := CreateData(user, nil)

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

func ExtensionPage(w http.ResponseWriter, r *http.Request, re render.Render) {
    re.HTML(200, "cextension", map[string]string {
        "loggedin": "false",
        "q": r.URL.Query().Get("q"),
        "repo": r.URL.Query().Get("repo"),
    })
}

func LogoutAction(session sessions.Session, re render.Render) {
    session.Clear()
    re.Redirect("/")
}

func AlreadyLogRedirect(re render.Render, user User) {
    if user.IsLoggedIn() {
        re.Redirect("/")
    }
}

func LoginPage(user User, session sessions.Session, re render.Render) {
    template := CreateData(user, session)

    re.HTML(200, "login", template)
}

func LoginAction(session sessions.Session, r *http.Request, re render.Render) {
    r.ParseForm()
    username, password := r.FormValue("username"), r.FormValue("password")

    if username == "" || password == "" {
        session.AddFlash("invalid length", "message")
        re.Redirect("/login")
        return
    }

    query := bson.M { "username_lower": strings.ToLower(username) }

    var result bson.M
    err := database.C("users").Find(query).One(&result)

    if err != nil {
        session.AddFlash("Cannot find a user with that username and password", "message")
        re.Redirect("/login")
        return
    }

    bytes := []byte(password)
    hashed := []byte(result["password"].(string))

    if err = bcrypt.CompareHashAndPassword(hashed, bytes); err == nil {
        session.Set("loggedin", true)
        session.Set("username", username)
        session.Set("token", result["github_token"].(string))
        re.Redirect("/")
        return
    }
    session.AddFlash("Cannot find a user with that username and password", "message")
    re.Redirect("/login")
}

func RegisterPage(user User, session sessions.Session, re render.Render) {
    template := CreateData(user, session)

    re.HTML(200, "register", template)
}

func RegisterAction(session sessions.Session, r *http.Request, re render.Render) {
    r.ParseForm()
    username, password := r.FormValue("username"), r.FormValue("password")

    if username == "" || password == "" {
        session.AddFlash("invalid length", "message")
        re.Redirect("/register")
        return
    }

    query := bson.M { "username_lower": strings.ToLower(username) }

    var result bson.M
    err := database.C("users").Find(query).One(&result)

    if err == nil {
        session.AddFlash("Username is taken", "message")
        re.Redirect("/register")
        return
    }

    hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    doc := bson.M {
        "_id": bson.NewObjectId(),
        "username": username,
        "username_lower": strings.ToLower(username),
        "password": string(hashed),
        "github_id": "",
        "github_token": "",
    }

    if err = database.C("users").Insert(doc); err == nil {
        session.Set("loggedin", true)
        session.Set("username", username)
        session.Set("token", "")
        re.Redirect("/")
        return
    }
    session.AddFlash("Database error", "message")
    re.Redirect("/")
}

func GithubAuth(session sessions.Session, re render.Render) {
    client_id := config.OAuth2Client_ID

    letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

    b := make([]rune, 10)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    session.Set("state", string(b))

    re.Redirect("https://github.com/login/oauth/authorize?client_id=" + client_id + "&state=" + string(b))
}

func GithubCallback(session sessions.Session, r *http.Request, re render.Render) {
    code, state := r.URL.Query().Get("code"), r.URL.Query().Get("state")
    client_id, client_secret := config.OAuth2Client_ID, config.OAuth2Client_Secret

    if code == "" || state == "" || session.Get("state") == nil || session.Get("state").(string) != state {
        re.Redirect("/")
        return
    }
    session.Delete("state")

    form := url.Values {}
    form.Add("code", code)
    form.Add("client_id", client_id)
    form.Add("client_secret", client_secret)

    req, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token/", strings.NewReader(form.Encode()))
    req.Header.Set("Accept", "application/json")

    client := &http.Client {}
    resp, _ := client.Do(req)

    body, _ := ioutil.ReadAll(resp.Body)

    log.Println(string(body))

    parse := map[string]interface{} {}
    json.Unmarshal([]byte(string(body)), &parse)

    log.Println(parse)

    token := parse["access_token"].(string)

    log.Println(token)

    req, _ = http.NewRequest("GET", "https://api.github.com/user?access_token=" + token, nil)
    resp, _ = client.Do(req)

    body, _ = ioutil.ReadAll(resp.Body)

    parse = map[string]interface{} {}

    json.Unmarshal([]byte(string(body)), &parse)

    log.Println(string(body))

    username, id := parse["login"].(string), strconv.FormatFloat(parse["id"].(float64), 'f', -1, 64)

    log.Println(id)

    query := bson.M { "username_lower": strings.ToLower(username) }

    var result bson.M

    err := database.C("users").Find(query).One(&result)
    if err == nil {
        if result["github_id"].(string) == id {
            session.Set("loggedin", true)
            session.Set("username", username)
            session.Set("token", token)
            re.Redirect("/")
        } else {
            i := 0
            for {
                usr_i := username + "_" + strconv.Itoa(i)
                query := bson.M { "username_lower": strings.ToLower(usr_i) }
                err := database.C("users").Find(query).One(&parse)

                if err != nil {
                    i++
                } else {
                    doc := bson.M {
                        "_id": bson.NewObjectId(),
                        "username": usr_i,
                        "username_lower": strings.ToLower(usr_i),
                        "password": "",
                        "github_id": id,
                        "github_token": token,
                    }

                    database.C("users").Insert(doc)
                    session.Set("loggedin", true)
                    session.Set("username", usr_i)
                    session.Set("token", token)
                    re.Redirect("/")
                    break
                }
            }
        }
    } else {
        doc := bson.M {
            "_id": bson.NewObjectId(),
            "username": username,
            "username_lower": strings.ToLower(username),
            "password": "",
            "github_id": id,
            "github_token": token,
        }
        database.C("users").Insert(doc)
        session.Set("loggedin", true)
        session.Set("username", username)
        session.Set("token", token)
        re.Redirect("/")
    }
}

