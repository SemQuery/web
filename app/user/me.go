package user

import (
    "github.com/semquery/web/app/common"

    "github.com/martini-contrib/render"
)

func MePage(user common.User, r render.Render) {
    data := struct {
        Loggedin bool
        Username string
        Repos []string
        Pagename string
    } {
        Loggedin: user.IsLoggedIn(),
        Username: user.Username(),
        Repos: user.GetIndexed(),
        Pagename: "me",
    }
    r.HTML(200, "me", data)
}
