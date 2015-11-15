package common

import (
    "github.com/go-martini/martini"
    "github.com/martini-contrib/sessions"
    "golang.org/x/oauth2"
    "github.com/google/go-github/github"

    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"

    "time"
    "strconv"
)

var UsersColl *mgo.Collection

type User interface {
    IsLoggedIn() bool
    Username() string
    Id() string
    AddIndexed(string)
    GetIndexed() []string
    GetPlan() map[string]string
    Github() *github.Client
}

type user struct {
    isLoggedIn bool
    username, id string
    github *github.Client
}

func UserExist(id string) bool {
    var usrdat bson.M
    query := bson.M { "id": id }

    err := UsersColl.Find(query).One(&usrdat)
    return err == nil
}

func NewUser(id string) {
    now := time.Now()
    doc := bson.M {
        "id": id,
        "repos": []string {},
        "plan": map[string]string {
            "name": "normal",
            "expire": strconv.Itoa(now.Day()) + " " + strconv.Itoa(int(now.Month())) + " " + strconv.Itoa(now.Year()),
        },
    }
    UsersColl.Insert(doc)
}

func (u user) AddIndexed(repo string) {
    list := append(u.GetIndexed(), repo)

    query := bson.M { "id": u.id }
    updt := bson.M { "$set": bson.M { "repos": list } }

    UsersColl.Update(query, updt)
}

func (u user) GetIndexed() []string {
    var usrdat bson.M
    query := bson.M { "id": u.id }
    list := []string {}
    if err := UsersColl.Find(query).One(&usrdat); err == nil {
        for _, s := range usrdat["repos"].([]interface{}) {
            list = append(list, s.(string))
        }
    }
    return list
}

func (u user) GetPlan() map[string]string {
    var usrdat bson.M
    query := bson.M { "id": u.id }
    UsersColl.Find(query).One(&usrdat)

    return usrdat["plan"].(map[string]string)
}

func (u user) IsLoggedIn() bool {
    return u.isLoggedIn
}

func (u user) Username() string {
    return u.username
}

func (u user) Github() *github.Client {
    return u.github
}

func (u user) Id() string {
    return u.id
}

func UserInject(session sessions.Session, ctx martini.Context) {
    u := user { isLoggedIn: false, username: "", id: "", github: nil }

    if session.Get("loggedin") != nil {
        u.isLoggedIn = true
        u.username = session.Get("username").(string)
        u.id = session.Get("id").(string)

        ts := oauth2.StaticTokenSource(
            &oauth2.Token {AccessToken: session.Get("token").(string)},
        )

        tc := oauth2.NewClient(oauth2.NoContext, ts)
        client := github.NewClient(tc)

        u.github = client
    }

    ctx.MapTo(u, (*User) (nil))
}
