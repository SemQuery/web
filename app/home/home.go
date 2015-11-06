package home

import (
    "github.com/semquery/web/app/common"

    "github.com/martini-contrib/render"
)


func HomePage(user common.User, r render.Render) {
    data := struct {
        Loggedin bool
        Usrname string
    } {
        Loggedin: user.IsLoggedIn(),
        Usrname: user.Username(),
    }
    r.HTML(200, "index", data)
}
