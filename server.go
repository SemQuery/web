package main

import (
    "github.com/semquery/web/app/common"
    "github.com/semquery/web/app/routes"

    "github.com/go-martini/martini"
    "github.com/martini-contrib/sessions"
    "github.com/martini-contrib/render"

    "log"
    "os"
    "encoding/json"

    "gopkg.in/mgo.v2"
)

func main() {
    cfg, err := os.Open("config.json")
    if err != nil {
        log.Fatal(err)
    }
    parser := json.NewDecoder(cfg)
    if err = parser.Decode(common.Config); err != nil {
        log.Fatal(err)
    }

    session, err := mgo.DialWithInfo(&mgo.DialInfo{
        Addrs: []string{common.Config.DBAddr},
        Database: common.Config.DBName,
        Username: common.Config.DBUser,
        Password: common.Config.DBPass,
    })
    if err != nil {
        log.Fatal(err)
    }

    common.Database = session.DB(common.Config.DBName)
    log.Print("Database online")

    m := martini.Classic()
    m.Use(sessions.Sessions("semquery", sessions.NewCookieStore([]byte("secret"))))
    m.Use(render.Renderer(render.Options {
        Layout: "layout",
    }))
    m.Use(common.UserInject)

    routes.RegisterRoutes(m)

    m.Run()
}

