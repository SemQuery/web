package main

import (
    "github.com/semquery/web/app/common"
    "github.com/semquery/web/app/routes"

    "github.com/go-martini/martini"
    "github.com/martini-contrib/render"
    "github.com/martini-contrib/sessions"

    "encoding/json"
    "log"
    "os"

    "gopkg.in/mgo.v2"
    "gopkg.in/redis.v3"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/sqs"
)

func initDB() {
    session, err := mgo.DialWithInfo(&mgo.DialInfo{
        Addrs:    []string{common.Config.DBAddr},
        Database: common.Config.DBName,
        Username: common.Config.DBUser,
        Password: common.Config.DBPass,
    })
    if err != nil {
        log.Fatal(err)
    }

    common.Database = session.DB(common.Config.DBName)
    log.Print("Database online")

    common.Rds = redis.NewClient(&redis.Options {
        Addr: common.Config.RedisAddr,
        Password: common.Config.RedisPass,
        DB: common.Config.RedisDB,
    })
}

func initQueue() {
    common.Queue = sqs.New(&aws.Config{
        Region: &common.Config.QueueRegion,
    })

    input := sqs.GetQueueUrlInput{
        QueueName: &common.Config.QueueName,
    }
    output, err := common.Queue.GetQueueUrl(&input)
    if err != nil {
        log.Fatal(err)
    }
    common.QueueURL = *output.QueueUrl
}

func main() {
    cfg, err := os.Open("config.json")
    if err != nil {
        log.Fatal(err)
    }
    parser := json.NewDecoder(cfg)
    if err = parser.Decode(common.Config); err != nil {
        log.Fatal(err)
    }

    initDB()
    initQueue()

    m := martini.Classic()
    m.Use(sessions.Sessions("semquery", sessions.NewCookieStore([]byte("secret"))))
    m.Use(render.Renderer(render.Options {
        Layout: "layout",
    }))
    m.Use(common.UserInject)

    routes.RegisterRoutes(m)

    if len(os.Args) > 1 {
        m.RunOnAddr(os.Args[1])
    } else {
        m.Run()
    }
}
