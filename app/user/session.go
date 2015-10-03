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
)

func LogoutAction(session sessions.Session, re render.Render) {
    session.Clear()
    re.Redirect("/")
}

func Login(session sessions.Session, re render.Render) {
    client_id := common.Config.OAuth2Client_ID

    letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

    b := make([]rune, 10)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    session.AddFlash(string(b), "state")

    re.Redirect("https://github.com/login/oauth/authorize?client_id=" + client_id + "&state=" + string(b))
}

func GithubCallback(session sessions.Session, r *http.Request, re render.Render) {
    code, state := r.URL.Query().Get("code"), r.URL.Query().Get("state")
    client_id, client_secret := common.Config.OAuth2Client_ID, common.Config.OAuth2Client_Secret

    flashes := session.Flashes("state")
    if code == "" || state == "" || session == nil || len(flashes) == 0 || flashes[0].(string) != state {
        re.Redirect("/")
        return
    }

    form := url.Values {}
    form.Add("code", code)
    form.Add("client_id", client_id)
    form.Add("client_secret", client_secret)

    req, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token/", strings.NewReader(form.Encode()))
    req.Header.Set("Accept", "application/json")

    client := &http.Client {}
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

    username := parse["login"].(string)

    session.Set("loggedin", true)
    session.Set("username", username)
    session.Set("token", token)
    re.Redirect("/")

}

