package cextension

import (
    "net/http"

    "github.com/martini-contrib/render"
)

func ExtensionPage(w http.ResponseWriter, r *http.Request, re render.Render) {
    re.HTML(200, "cextension", map[string]string {
        "loggedin": "false",
        "q": r.URL.Query().Get("q"),
        "repo": r.URL.Query().Get("repo"),
    })
}
