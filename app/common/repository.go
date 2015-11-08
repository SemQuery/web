package common

import (
    "gopkg.in/mgo.v2/bson"

    "net/url"
)

const CodeSourceColl = "sources"

const (
    CodeSourceGitHub = "github"
    CodeSourceLink   = "link"
)

type CodeSourceStatus string

const (
    CodeSourceStatusNotFound = "none"
    CodeSourceStatusWorking  = "working"
    CodeSourceStatusDone     = "done"
)

type CodeSource interface {
    ToQuery() bson.M
}

type RepositorySource struct {
    User string
    Name string
}

func (r *RepositorySource) ToQuery() bson.M {
    return bson.M{
        "type": "github_repo",
        "repo_user": r.User,
        "repo_name": r.Name,
    }
}

type LinkSource struct {
    URL *url.URL
}

func (l *LinkSource) ToQuery() bson.M {
    return bson.M{
        "type": "link",
        "link_url": l.URL.String(),
    }
}

func GetCodeSourceStatus(src CodeSource) CodeSourceStatus {
    doc := src.ToQuery()

    var res bson.M
    err := Database.C(CodeSourceColl).Find(doc).One(&res)

    if err == nil {
        return CodeSourceStatus(res["status"].(string))
    } else {
        return CodeSourceStatusNotFound
    }
}
