package main

import (
    "github.com/semquery/web/app/common"

    "github.com/go-martini/martini"
    "github.com/martini-contrib/sessions"
    "github.com/martini-contrib/render"

    "log"
    "os"
    "encoding/json"

    "gopkg.in/mgo.v2"
)

var config struct {
    WebAddr string `json:"web_addr"`

    DBAddr string `json:"db_addr"`
    DBName string `json:"db_name"`
    DBUser string `json:"db_user"`
    DBPass string `json:"db_pass"`

    OAuth2Client_ID string `json:"github_id"`
    OAuth2Client_Secret string `json:"github_secret"`
}

var database *mgo.Database

func main() {
    cfg, err := os.Open("config.json")
    if err != nil {
        log.Fatal(err)
    }
    parser := json.NewDecoder(cfg)
    if err = parser.Decode(&config); err != nil {
        log.Fatal("Bad json")
    }

    session, err := mgo.DialWithInfo(&mgo.DialInfo{
        Addrs: []string{config.DBAddr},
        Database: config.DBName,
        Username: config.DBUser,
        Password: config.DBPass,
    })
    if err != nil {
        log.Fatal(err)
    }

    database = session.DB(config.DBName)
    log.Print("Database online")

    m := martini.Classic()
    m.Use(sessions.Sessions("semquery", sessions.NewCookieStore([]byte("secret"))))
    m.Use(render.Renderer(render.Options {
        Layout: "layout",
    }))
    m.Use(common.UserInject)

    RegisterHandlers(m)

    m.Run()
}

