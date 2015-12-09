package routes

import (
    "github.com/go-martini/martini"

    "github.com/semquery/web/app/home"
    "github.com/semquery/web/app/query"
    "github.com/semquery/web/app/user"
    "github.com/semquery/web/app/cextension"
    "github.com/semquery/web/app/docs"
)

func RegisterRoutes(m *martini.ClassicMartini) {
    m.Get("/", home.HomePage)

    m.Get("/search", query.SearchPage)
    m.Post("/index_source", query.InitiateIndex)

    m.Get("/socket", query.SocketPage)

    m.Get("/me", user.MePage)

    m.Get("/login", user.Login)
    m.Post("/logout", user.LogoutAction)

    m.Get("/githubcallback", user.GithubCallback)

    m.Get("/cextension", cextension.ExtensionPage)

    m.Get("/docs", docs.DocsPage)
}
