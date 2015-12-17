package query

import (
    "github.com/semquery/web/app/common"

    "fmt"
    "net/url"
    "net/http"
    "encoding/json"

    "github.com/aws/aws-sdk-go/aws/request"
    "github.com/aws/aws-sdk-go/service/s3"
)

// files are split into 50KB sectors
const FILE_SECTOR_BYTES = 50000

// [ID]/[path]$[start]$[end]
const FILENAME_FORMAT = "%s/%s$%d$%d"

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

// Creates & signs the necessary GET Object requests to
// fetch the file sections associated with a query result.
func CreateS3Requests(qr QueryResult, id string) (reqs []*request.Request, err error) {
    reqs =  []*request.Request{}

    byteStart := FILE_SECTOR_BYTES * (qr.Start / FILE_SECTOR_BYTES)
    for byteStart <= qr.End {
        key := fmt.Sprintf(FILENAME_FORMAT, id, qr.File, byteStart, byteStart + FILE_SECTOR_BYTES)

        input := s3.GetObjectInput{
            Bucket: &common.Config.S3SourceCodeBucket,
            Key: &key,
        }
        req, _ := common.S3SourceCode.GetObjectRequest(&input)
        err = req.Sign()
        if err != nil { return }
        reqs = append(reqs, req)

        byteStart += FILE_SECTOR_BYTES
    }

    return
}
