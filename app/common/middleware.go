package common

import (
    "github.com/go-martini/martini"
    "github.com/martini-contrib/sessions"

    "strconv"
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

