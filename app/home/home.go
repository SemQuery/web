package home

import (
    "github.com/semquery/web/app/common"

    "github.com/martini-contrib/render"
)


func HomePage(user common.User, r render.Render) {
    data := common.CreateData(user, nil)
    r.HTML(200, "index", data)
}
