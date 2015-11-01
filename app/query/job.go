package query

import (
    "github.com/semquery/web/app/common"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/sqs"

    "log"
    "encoding/json"
)

type IndexingJob struct {
    Token string

    RepositoryPath string
}

func (job IndexingJob) toJson() string {
    encode, _ := json.Marshal(job)
    return string(encode)
}

// returns: whether queueing the job was successful
func QueueIndexingJob(job IndexingJob) bool {
    input := sqs.SendMessageInput{
        MessageBody: aws.String(job.toJson()),
        QueueUrl: &common.QueueURL,
    }

    _, err := common.Queue.SendMessage(&input)

    if err != nil {
        log.Print(err)
    }
    return err == nil
}
