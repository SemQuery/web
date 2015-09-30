package user

import (
    "github.com/semquery/web/app/common"

    "github.com/martini-contrib/sessions"
    "github.com/martini-contrib/render"

    "golang.org/x/crypto/bcrypt"
    "gopkg.in/mgo.v2/bson"

    "net/http"
    "net/url"
    "strings"
    "math/rand"
    "encoding/json"
    "log"
    "io/ioutil"
    "strconv"
)

func LogoutAction(session sessions.Session, re render.Render) {
    session.Clear()
    re.Redirect("/")
}

func AlreadyLogRedirect(re render.Render, user common.User) {
    if user.IsLoggedIn() {
        re.Redirect("/")
    }
}

func LoginPage(user common.User, session sessions.Session, re render.Render) {
    template := common.CreateData(user, session)

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
    err := common.Database.C("users").Find(query).One(&result)

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

func RegisterPage(user common.User, session sessions.Session, re render.Render) {
    template := common.CreateData(user, session)

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
    err := common.Database.C("users").Find(query).One(&result)

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

    if err = common.Database.C("users").Insert(doc); err == nil {
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
    client_id := common.Config.OAuth2Client_ID

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
    client_id, client_secret := common.Config.OAuth2Client_ID, common.Config.OAuth2Client_Secret

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

    err := common.Database.C("users").Find(query).One(&result)
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
                err := common.Database.C("users").Find(query).One(&parse)

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

                    common.Database.C("users").Insert(doc)
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
        common.Database.C("users").Insert(doc)
        session.Set("loggedin", true)
        session.Set("username", username)
        session.Set("token", token)
        re.Redirect("/")
    }
}

