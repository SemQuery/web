package user

import (
    "github.com/semquery/web/app/common"

    "github.com/martini-contrib/render"
)

func MePage(user common.User, r render.Render) {
    data := common.CreateData(user)
    r.HTML(200, "me", data)
}
