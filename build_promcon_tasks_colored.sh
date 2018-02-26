#!/bin/sh

[ -r github.token ] && export GITHUB_TOKEN=$(cat github.token)
[ -z GITHUB_TOKEN ] && echo 'set $GITHUB_TOKEN or put it into github.token' && exit 1
[ -z GOPATH ]       && echo 'set $GOPATH' && exit 2

$GOPATH/bin/tasks-state-graph
dot -Tsvg promcon_tasks_colored.dot > promcon_tasks_colored.svg
