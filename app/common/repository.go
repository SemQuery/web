package common

import (
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"

    "net/url"
    "strings"
)

var CodeSourceColl *mgo.Collection

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

type GitSource struct {
    URL *url.URL
}

// determines if this git source is also a
// GitHub repository
func (g *GitSource) IsGitHub() bool {
    if g.URL.Host != "github.com" { return false }

    path := strings.Split(g.URL.Path, "/")
    if len(path) != 3 { return false }

    return true
}

func (g *GitSource) GitHubFormat() (string, string) {
    if !g.IsGitHub() { return "", "" }

    path := strings.Split(g.URL.Path, "/")

    return path[1], strings.TrimSuffix(path[2], ".git")
}

func (g *GitSource) GitHubUser() string {
    user, _ := g.GitHubFormat()
    return user
}

func (g *GitSource) GitHubRepo() string {
    _, repo := g.GitHubFormat()
    return repo
}

func CreateGitHubSource(user, repo string) *CodeSource {
    url := &url.URL{
        Scheme: "https",
        Host: "github.com",
        Path: "/" + user + "/" + repo + ".git",
    }

    return &CodeSource{&GitSource{url}}
}

func CreateGitSource(url *url.URL) *CodeSource {
    return &CodeSource{&GitSource{url}}
}

type CodeSource struct {
    Git *GitSource
}

func (cs *CodeSource) ToBson() bson.M {
    b := bson.M {}
    if cs.Git != nil {
        g := bson.M{
            "url": cs.Git.URL.String(),
        }
        if cs.Git.IsGitHub() {
            user, repo := cs.Git.GitHubFormat()
            g["github_user"] = user
            g["github_repo"] = repo
        }
        b["git"] = g
    }
    return b
}

func ToFlatDocument(doc bson.M) (flat bson.M) {
    flat = bson.M{}
    for k, v := range doc {
        switch v.(type) {
        case bson.M:
            toFlatDocument(flat, v.(bson.M), k)
        default:
            flat[k] = v
        }
    }
    return
}

func toFlatDocument(root, current bson.M, path string) {
    for k, v := range current {
        switch v.(type) {
        case bson.M:
            toFlatDocument(root, v.(bson.M), k)
        default:
            root[path + "." + k] = v
        }
    }
}

func GetCodeSourceStatus(src *CodeSource) CodeSourceStatus {
    doc := ToFlatDocument(src.ToBson())

    var res bson.M
    err := CodeSourceColl.Find(doc).One(&res)

    if err == nil {
        return CodeSourceStatus(res["status"].(string))
    } else {
        return CodeSourceStatusNotFound
    }
}

func InsertSource(src *CodeSource, status CodeSourceStatus) (bson.ObjectId, error) {
    doc  := src.ToBson()
    doc["status"] = status

    id := bson.NewObjectId()
    doc["_id"] = id

    err :=  CodeSourceColl.Insert(doc)

    return id, err
}

func UpdateStatus(id bson.ObjectId, status CodeSourceStatus) error {
    update := bson.M{
        "$set": bson.M{"status": status},
    }

    return CodeSourceColl.UpdateId(id, update)
}
