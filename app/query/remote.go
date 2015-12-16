package query

import (
    "github.com/semquery/web/app/common"

    "net/url"
    "net/http"
    "encoding/json"
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
