package routes

import (
    "github.com/go-martini/martini"

    "github.com/semquery/web/app/home"
    "github.com/semquery/web/app/query"
    "github.com/semquery/web/app/user"
    "github.com/semquery/web/app/cextension"
)

func RegisterRoutes(m *martini.ClassicMartini) {
    m.Get("/", home.HomePage)

    m.Post("/query", query.QueryPage)
    m.Get("/socket", query.SocketPage)

    m.Get("/login", user.AlreadyLogRedirect, user.LoginPage)
    m.Post("/login", user.AlreadyLogRedirect, user.LoginAction)
    m.Get("/register", user.AlreadyLogRedirect, user.RegisterPage)
    m.Post("/register", user.AlreadyLogRedirect, user.RegisterAction)
    m.Get("/logout", user.LogoutAction)

    m.Get("/githubauth", user.AlreadyLogRedirect, user.GithubAuth)
    m.Get("/githubcallback", user.AlreadyLogRedirect, user.GithubCallback)

    m.Get("/cextension", cextension.ExtensionPage)
}
