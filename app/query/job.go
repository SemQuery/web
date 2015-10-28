package query

import (
    "github.com/semquery/web/app/common"

    "log"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/sqs"
)

type IndexingJob struct {
    Token string

    RepositoryPath string
}

// returns: whether queueing the job was successful
func QueueIndexingJob(job IndexingJob) bool {
    input := sqs.SendMessageInput{
        MessageAttributes: map[string]*sqs.MessageAttributeValue {
            "path": {
                DataType: aws.String("String"),
                StringValue: aws.String(job.RepositoryPath),
            },
            "token": {
                DataType: aws.String("String"),
                StringValue: aws.String(job.Token),
            },

        },
        QueueUrl: &common.QueueURL,
    }

    _, err := common.Queue.SendMessage(&input)

    if err != nil {
        log.Print(err)
    }
    return err == nil
}
