package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/awalterschulze/gographviz"
	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
)

const titleAttribute = "label"

func createIssue(ctx context.Context, client *github.Client, owner string, repo string, node *gographviz.Node) error {
	title := strings.Trim(node.Attrs[titleAttribute], `"`)
	ir := &github.IssueRequest{
		Title:  &title,
		Body:   github.String("[node-id=" + node.Name + "]"),
		Labels: &[]string{"promcon-2019"},
	}
	_, _, err := client.Issues.Create(ctx, "prometheus", "promcon", ir)
	return err
}

func main() {
	dotFile := flag.String("dotfile", "promcon_tasks.dot", "The path to the dot file containing the issues to create.")
	githubOwner := flag.String("github-org", "prometheus", "The GitHub org which contains the issues.")
	githubRepo := flag.String("github-repo", "promcon", "The repo within the GitHub org which contains the issues.")

	flag.Parse()

	buf, err := ioutil.ReadFile(*dotFile)
	if err != nil {
		log.Fatal("Error reading dot file:", err)
	}
	graph, err := gographviz.Read(buf)
	if err != nil {
		log.Fatal("Error parsing dot file:", err)
	}

	for _, n := range graph.Nodes.Nodes {
		if n.Attrs[titleAttribute] == "" {
			log.Fatalf("Node %q is missing the %q attribute", n.Name, titleAttribute)
		}
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("Please set the GITHUB_TOKEN environment variable")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	for _, n := range graph.Nodes.Nodes {
		fmt.Printf("Creating issue for %q...\n", n.Name)
		createIssue(ctx, client, *githubOwner, *githubRepo, n)
	}
}
