package common

import (
    "gopkg.in/mgo.v2"

    "github.com/aws/aws-sdk-go/service/sqs"
)

type config struct {
    WebAddr string `json:"web_addr"`

    DBAddr string `json:"db_addr"`
    DBName string `json:"db_name"`
    DBUser string `json:"db_user"`
    DBPass string `json:"db_pass"`

    OAuth2Client_ID string `json:"github_id"`
    OAuth2Client_Secret string `json:"github_secret"`

    EngineExecutable string `json:"engine_executable"`

    QueueName   string `json:"sqs_name"`
    QueueRegion string `json:"sqs_region"`
}

var Config *config = &config{}

var Database *mgo.Database

var Queue *sqs.SQS
var QueueURL string
