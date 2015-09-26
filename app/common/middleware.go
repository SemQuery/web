package common

import (
    "github.com/go-martini/martini"
    "github.com/martini-contrib/sessions"
)

type User interface {
    IsLoggedIn() bool
    Username() string
    Token() string
}

type user struct {
    isLoggedIn bool
    username string
    token string
}

func (u user) IsLoggedIn() bool {
    return u.isLoggedIn
}

func (u user) Username() string {
    return u.username
}

func (u user) Token() string {
    return u.token
}

func UserInject(session sessions.Session, ctx martini.Context) {
    u := user { isLoggedIn: false, username: "", token: "" }

    if session.Get("loggedin") != nil {
        u.isLoggedIn = true
        u.username = session.Get("username").(string)
        u.token = session.Get("username").(string)
    }

    ctx.MapTo(u, (*User) (nil))
}
