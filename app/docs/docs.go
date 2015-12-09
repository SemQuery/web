package docs

import (
    "github.com/semquery/web/app/common"

    "github.com/martini-contrib/render"
)


func DocsPage(user common.User, r render.Render) {
    data := struct {
        common.User
        Pagename string
        Theme string
    } {user, "docs", "standard"}
    r.HTML(200, "docs", data)
}
