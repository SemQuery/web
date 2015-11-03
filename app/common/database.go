package common

import (
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

var Database *mgo.Database

const ReposColl = "repositories"

type RepoStatus string

const (
    RepoStatusNotFound = "none"
    RepoStatusWorking  = "working"
    RepoStatusDone     = "done"
)

func RepositoryStatus(r *Repository) RepoStatus {
    doc := bson.M{
        "user": r.User,
        "name": r.Name,
    }
    var res bson.M
    err := Database.C(ReposColl).Find(doc).One(&res)

    if err == nil {
        return res["status"].(RepoStatus)
    } else {
        return RepoStatusNotFound
    }
}
