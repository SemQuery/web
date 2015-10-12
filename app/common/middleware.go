package common

import (
    "github.com/go-martini/martini"
    "github.com/martini-contrib/sessions"

    "golang.org/x/oauth2"
    "github.com/google/go-github/github"

    "strconv"
)

type User interface {
    IsLoggedIn() bool
    Username() string
    Github() *github.Client
}

type user struct {
    isLoggedIn bool
    username string
    github *github.Client
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

func UserInject(session sessions.Session, ctx martini.Context) {
    u := user { isLoggedIn: false, username: "", github: nil }

    if session.Get("loggedin") != nil {
        u.isLoggedIn = true
        u.username = session.Get("username").(string)

        ts := oauth2.StaticTokenSource(
            &oauth2.Token {AccessToken: session.Get("token").(string)},
        )

        tc := oauth2.NewClient(oauth2.NoContext, ts)
        client := github.NewClient(tc)

        u.github = client
    }

    ctx.MapTo(u, (*User) (nil))
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

