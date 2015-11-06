package common

import (

    "github.com/go-martini/martini"
    "github.com/martini-contrib/sessions"
    "golang.org/x/oauth2"
    "github.com/google/go-github/github"

    "labix.org/v2/mgo/bson"

)

type User interface {
    IsLoggedIn() bool
    Username() string
    Id() string
    AddIndexed(string)
    GetIndexed() []string
    Github() *github.Client
}

type user struct {
    isLoggedIn bool
    username, id string
    github *github.Client
}

func (u user) AddIndexed(repo string) {
    list := append(u.GetIndexed(), repo)

    query := bson.M { "id": u.id }
    updt := bson.M { "$set": bson.M { "repos": list } }

    Database.C("users").Update(query, updt)
}

func (u user) GetIndexed() []string {
    var usrdat bson.M
    query := bson.M { "id": u.id }
    list := []string {}
    if err := Database.C("users").Find(query).One(&usrdat); err == nil {
        for _, s := range usrdat["repos"].([]interface{}) {
            list = append(list, s.(string))
        }
    }
    return list
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
