package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/awalterschulze/gographviz"
	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
)

var nodeIDRe = regexp.MustCompile(`\[node-id=(.*?)\]`)

func getIssues(ctx context.Context, client *github.Client, owner string, repo string, label string) ([]*github.Issue, error) {
	issues, _, err := client.Issues.ListByRepo(ctx, owner, repo, &github.IssueListByRepoOptions{
		Labels: []string{label},
		State:  "all",
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 10000,
		},
	})
	return issues, err
}

func main() {
	inDotFile := flag.String("dotfile-in", "promcon_tasks.dot", "The path to the input dot file.")
	outDotFile := flag.String("dotfile-out", "promcon_tasks_colored.dot", "The path to the output dot file.")
	githubOwner := flag.String("github-org", "prometheus", "The GitHub org which contains the issues.")
	githubRepo := flag.String("github-repo", "promcon", "The repo within the GitHub org which contains the issues.")
	githubLabel := flag.String("github-label", "promcon-2019", "The label by which to select issues.")

	flag.Parse()

	buf, err := ioutil.ReadFile(*inDotFile)
	if err != nil {
		log.Fatal("Error reading dot file:", err)
	}
	graph, err := gographviz.Read(buf)
	if err != nil {
		log.Fatal("Error parsing dot file:", err)
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

	log.Println("Fetching issues from GitHub...")
	issues, err := getIssues(ctx, client, *githubOwner, *githubRepo, *githubLabel)
	if err != nil {
		log.Fatal("Error fetching issues from GitHub:", err)
	}
	log.Printf("Found %d issues.", len(issues))

	nodeIDToIssue := map[string]*github.Issue{}
	for _, issue := range issues {
		m := nodeIDRe.FindStringSubmatch(issue.GetBody())
		if len(m) < 2 {
			log.Printf("Skipping issue %s with missing ID marker", issue.GetHTMLURL())
			continue
		}

		nodeIDToIssue[m[1]] = issue
	}

	log.Println("Building output graph...")
	for _, n := range graph.Nodes.Nodes {
		issue, ok := nodeIDToIssue[n.Name]
		if !ok {
			// TODO: Mark the node as having a missing issue somehow!
			continue
		}
		color := "coral"
		if issue.GetState() == "closed" {
			color = "chartreuse"
		}

		n.Attrs.Add("style", "filled")
		n.Attrs.Add("fillcolor", color)
		n.Attrs.Add("URL", `"`+issue.GetHTMLURL()+`"`) // The graphViz library doesn't quote this correctly.
	}

	log.Println("Writing output dot file...")
	ioutil.WriteFile(*outDotFile, []byte(graph.String()), 0666)
}
