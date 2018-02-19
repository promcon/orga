package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/awalterschulze/gographviz"
)

var nodeIDRe = regexp.MustCompile(`\[node-id=(.*?)\]`)

type issue struct {
	State string `json:"state"`
	Body  string `json:"body"`
	URL   string `json:"html_url"`
}

func getIssues(token, owner, repo, params string) (map[string]issue, error) {
	client := &http.Client{}

	page := 1
	allIssues := map[string]issue{}
	for {
		log.Printf("Getting issues page %d...", page)
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?per_page=100&page=%d&state=&%s", owner, repo, page, params)
		req, err := http.NewRequest("GET", url, nil)
		req.Header.Add("Authorization", "token "+token)
		req.Header.Add("Accept", "application/vnd.github.v3.star+json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("bad response code from GitHub: %s", resp.Status)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()

		issues := []issue{}
		err = json.Unmarshal(body, &issues)
		if len(issues) == 0 {
			break
		}
		log.Printf("Got %d issues", len(issues))
		for _, issue := range issues {
			m := nodeIDRe.FindStringSubmatch(issue.Body)
			if len(m) < 2 {
				log.Printf("Skipping issue %s with missing ID marker", issue.URL)
				continue
			}
			allIssues[m[1]] = issue
		}
		page++
	}
	return allIssues, nil
}

func main() {
	inDotFile := flag.String("dotfile-in", "promcon_tasks.dot", "The path to the input dot file.")
	outDotFile := flag.String("dotfile-out", "promcon_tasks_colored.dot", "The path to the output dot file.")
	githubOwner := flag.String("github-org", "prometheus", "The GitHub org which contains the issues.")
	githubRepo := flag.String("github-repo", "promcon", "The repo within the GitHub org which contains the issues.")
	githubParams := flag.String("github-http-params", "labels=promcon-2018", "Extra HTTP params to add verbatim as a string to the GitHub issues URL.")

	flag.Parse()

	buf, err := ioutil.ReadFile(*inDotFile)
	if err != nil {
		log.Fatal("Error reading dot file:", err)
	}
	graph, err := gographviz.Read(buf)
	if err != nil {
		log.Fatal("Error parsing dot file:", err)
	}

	issues, err := getIssues(os.Getenv("GITHUB_TOKEN"), *githubOwner, *githubRepo, *githubParams)
	if err != nil {
		log.Fatal("Error fetching issues from GitHub:", err)
	}

	for _, n := range graph.Nodes.Nodes {
		issue, ok := issues[n.Name]
		if !ok {
			// TODO: Mark the node as having a missing issue somehow!
			continue
		}
		color := "chartreuse"
		if issue.State == "closed" {
			color = "coral"
		}

		n.Attrs.Add("style", "filled")
		n.Attrs.Add("fillcolor", color)
		n.Attrs.Add("URL", `"`+issue.URL+`"`) // The graphViz library doesn't quote this correctly.
	}

	ioutil.WriteFile(*outDotFile, []byte(graph.String()), 0666)
}
