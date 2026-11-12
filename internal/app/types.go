package app

// The names this package reads as its own. The document lives in
// internal/workspace and the sending in internal/request; they are aliased
// here so a view says "Request" and "Result" rather than carrying a package
// prefix through every line.

import (
	"github.com/yaffalhakim1/endpoint/internal/request"
	"github.com/yaffalhakim1/endpoint/internal/workspace"
)

type (
	Auth        = workspace.Auth
	AuthType    = workspace.AuthType
	BodyType    = workspace.BodyType
	Document    = workspace.Document
	Environment = workspace.Environment
	Folder      = workspace.Folder
	Pair        = workspace.Pair
	Project     = workspace.Project
	Recent      = workspace.Recent
	Recovery    = workspace.Recovery
	Request     = workspace.Request
	Result      = request.Result
)

const (
	AuthNone    = workspace.AuthNone
	AuthBearer  = workspace.AuthBearer
	AuthBasic   = workspace.AuthBasic
	AuthKey     = workspace.AuthKey
	AuthInherit = workspace.AuthInherit

	BodyNone = workspace.BodyNone
	BodyJSON = workspace.BodyJSON
	BodyForm = workspace.BodyForm
	BodyFile = workspace.BodyFile
)
