package home

import (
    "github.com/semquery/web/app/common"

    "github.com/martini-contrib/render"
)


func HomePage(user common.User, r render.Render) {
    data := struct {
        common.User
        Pagename string
        Theme    string
    } {user, "home", "home"}
    r.HTML(200, "index", data)
}
