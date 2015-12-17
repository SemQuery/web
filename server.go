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
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/sqs"
    "github.com/aws/aws-sdk-go/service/s3"
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
    common.UsersColl = common.Database.C("users")
    common.CodeSourceColl = common.Database.C("sources")
    log.Print("Database online")

    common.Rds = redis.NewClient(&redis.Options {
        Addr: common.Config.RedisAddr,
        Password: common.Config.RedisPass,
        DB: common.Config.RedisDB,
    })
}

func initQueue() {
    common.Queue = sqs.New(session.New(), &aws.Config{
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

func initS3() {
    common.S3SourceCode = s3.New(session.New(), &aws.Config{
        Region: &common.Config.S3SourceCodeRegion,
    })
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
    initS3()

    m := martini.Classic()

    store := sessions.NewCookieStore([]byte(common.Config.SessionsSecret))
    m.Use(sessions.Sessions("semquery", store))

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
