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

func (r *CodeSource) ToBson() bson.M {
    b := bson.M {}
    if r.Git != nil {
        g := bson.M{
            "url": r.Git.URL.String(),
        }
        if r.Git.IsGitHub() {
            user, repo := r.Git.GitHubFormat()
            g["github_user"] = user
            g["github_repo"] = repo
        }
        b["git"] = g
    }
    return b
}

func GetCodeSourceStatus(src *CodeSource) CodeSourceStatus {
    doc := src.ToBson()

    var res bson.M
    err := CodeSourceColl.Find(doc).One(&res)

    if err == nil {
        return CodeSourceStatus(res["status"].(string))
    } else {
        return CodeSourceStatusNotFound
    }
}

func UpdateStatus(src *CodeSource, status CodeSourceStatus) *bson.ObjectId {
    query  := src.ToBson()
    update := bson.M { "$set": bson.M { "status": status } }

    info, _ :=  CodeSourceColl.Upsert(query, update)
    if info != nil {
        id := info.UpsertedId.(bson.ObjectId)
        return &id
    } else {
        return nil
    }
}
