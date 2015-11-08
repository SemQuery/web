package user

import (
    "github.com/semquery/web/app/common"

    "github.com/martini-contrib/render"
)

func MePage(user common.User, r render.Render) {
    data := struct {
        Loggedin bool
        Usrname string
        Repos []string
    } {
        Loggedin: user.IsLoggedIn(),
        Usrname: user.Username(),
        Repos: user.GetIndexed(),
    }
    r.HTML(200, "me", data)
}
