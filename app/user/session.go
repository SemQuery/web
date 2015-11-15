package user

import (
    "github.com/semquery/web/app/common"

    "github.com/martini-contrib/sessions"
    "github.com/martini-contrib/render"

    "net/http"
    "net/url"
    "strings"
    "math/rand"
    "encoding/json"
    "io/ioutil"
    "strconv"
)

func LogoutAction(session sessions.Session, re render.Render) {
    session.Clear()
    re.Redirect("/")
}

func Login(session sessions.Session, re render.Render, r *http.Request) {
    client_id := common.Config.OAuth2Client_ID

    letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

    b := make([]rune, 10)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    session.AddFlash(string(b), "state")

    redirectBack := r.URL.Query().Get("redirect_back")
    ref          := r.Referer()

    if redirectBack == "true" && ref != "" {
        session.Set("redirect_to", ref)
    } else {
        session.Set("redirect_to", nil)
    }

    query := url.Values{}
    query.Set("client_id", client_id)
    query.Set("state", string(b))

    dest := url.URL{
        Scheme:   "https",
        Host:     "github.com",
        Path:     "/login/oauth/authorize",
        RawQuery: query.Encode(),
    }
    re.Redirect(dest.String())
}

func GithubCallback(session sessions.Session, r *http.Request, re render.Render) {
    code, state := r.URL.Query().Get("code"), r.URL.Query().Get("state")
    client_id, client_secret := common.Config.OAuth2Client_ID, common.Config.OAuth2Client_Secret

    flashes := session.Flashes("state")
    if code == "" || state == "" || session == nil || len(flashes) == 0 || flashes[0].(string) != state {
        re.Redirect("/")
        return
    }

    client := &http.Client {}

    form := url.Values {}
    form.Add("code", code)
    form.Add("client_id", client_id)
    form.Add("client_secret", client_secret)

    req, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token/", strings.NewReader(form.Encode()))
    req.Header.Set("Accept", "application/json")

    resp, _ := client.Do(req)
    body, _ := ioutil.ReadAll(resp.Body)
    parse := map[string]interface{} {}
    json.Unmarshal([]byte(string(body)), &parse)

    token := parse["access_token"].(string)

    req, _ = http.NewRequest("GET", "https://api.github.com/user?access_token=" + token, nil)

    resp, _ = client.Do(req)
    body, _ = ioutil.ReadAll(resp.Body)
    parse = map[string]interface{} {}
    json.Unmarshal([]byte(string(body)), &parse)

    username, id := parse["login"].(string), strconv.FormatFloat(parse["id"].(float64), 'f', -1, 64)

    if !common.UserExist(id) {
        common.NewUser(id)
    }

    session.Set("loggedin", true)
    session.Set("username", username)
    session.Set("id", id)
    session.Set("token", token)

    redirectTo := session.Get("redirect_to")
    if redirectTo != nil {
        re.Redirect(redirectTo.(string))
    } else {
        re.Redirect("/")
    }
    session.Set("redirect_to", nil)
}

