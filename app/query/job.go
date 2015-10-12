package query

import (
    "github.com/semquery/web/app/common"

    "log"
    "encoding/json"
    "github.com/aws/aws-sdk-go/service/sqs"
)


type IndexingJob struct {
    User common.User

    RepositoryOwner string
    RepositoryName  string
}

func (j *IndexingJob) toSQSJson() string {
    data := map[string]interface{}{
        "user_id": "TODO",
        "repo_owner": j.RepositoryOwner,
        "repo_name": j.RepositoryName,
    }

    encoded, _ := json.Marshal(data)

    return string(encoded)
}

// returns: whether queueing the job was successful
func QueueIndexingJob(job IndexingJob) bool {
    msg := job.toSQSJson()
    input := sqs.SendMessageInput{
        MessageBody: &msg,
        QueueURL: &common.QueueURL,
    }

    _, err := common.Queue.SendMessage(&input)

    if err != nil {
        log.Print(err)
    }
    return err == nil
}
