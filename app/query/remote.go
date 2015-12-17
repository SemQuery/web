package query

import (
    "github.com/semquery/web/app/common"

    "net/url"
    "net/http"
    "encoding/json"

    "github.com/aws/aws-sdk-go/aws/request"
    "github.com/aws/aws-sdk-go/service/s3"
)

// Handles querying over HTTP

type QueryResult struct {
    File  string `json:"file"`
    Start int    `json:"start"`
    End   int    `json:"end"`
}

func ExecuteQuery(query, sourceID string) ([]QueryResult, error) {
    resp, err := http.PostForm(
        common.Config.QueryAddr,
        url.Values{"query": {query}, "source": {sourceID}})

    if err != nil {
        return nil, err
    }

    decoder := json.NewDecoder(resp.Body)

    var results []QueryResult
    err = decoder.Decode(results)
    if err != nil {
        return nil, err
    }

    return results, nil
}

// Creates & signs a GET Object request
func CreateS3Request(qr QueryResult, id string) (req *request.Request, err error) {
    input := s3.GetObjectInput{
        Bucket: &common.Config.S3SourceCodeBucket,
        Key: &qr.File,
    }
    req, _ = common.S3SourceCode.GetObjectRequest(&input)
    err = req.Sign()

    return
}
