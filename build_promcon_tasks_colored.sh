#!/bin/sh

[ -r github.token ] || echo 'file github.token does not exist.' && exit 1
[ -r github.token ] && export GITHUB_TOKEN=$(cat github.token)
[ -z GITHUB_TOKEN ] && echo 'set $GITHUB_TOKEN or put it into github.token' && exit 2
[ -z GOPATH ]       && echo 'set $GOPATH' && exit 3

$GOPATH/bin/tasks-state-graph
dot -Tsvg promcon_tasks_colored.dot > promcon_tasks_colored.svg
