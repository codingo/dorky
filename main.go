package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v38/github"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"
)

var (
	oFlag     = flag.Bool("o", false, "search for organization names")
	rFlag     = flag.Bool("r", false, "search for repository names")
	uFlag     = flag.Bool("u", false, "search for username matches")
	maxFlag   = flag.Int("max", 10, "maximum search results per category")
	cFlag     = flag.Bool("c", false, "clean input URLs")
	ghFlag    = flag.Bool("gh", false, "search only GitHub")
	glFlag    = flag.Bool("gl", false, "search only GitLab")
	sFlag     = flag.Bool("s", false, "simple output style for piping to another tool")
	urlRegexp = regexp.MustCompile(`^https?://(?:www\.)?([^/]+)`)
)

func main() {
	flag.Parse()

	// Ensure at least one search flag is specified
	if !(*oFlag || *rFlag || *uFlag) {
		fmt.Println("At least one search flag (-o, -r, or -u) must be specified")
		os.Exit(1)
	}

	// Read input words from stdin and store them in a map to avoid duplicates
	words := make(map[string]struct{})
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if *cFlag {
			word = cleanWord(word)
		}

		if _, exists := words[word]; !exists {
			words[word] = struct{}{}
		}
	}

	// Check for errors while reading stdin
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading stdin: %s\n", err)
		os.Exit(1)
	}

	// Perform searches on GitHub and/or GitLab for each input word
	for word := range words {
		if !*glFlag {
			searchGitHub(word)
		}

		if !*ghFlag {
			searchGitLab(word)
		}
	}
}

// cleanWord removes the URL scheme and www. prefix from the input word
func cleanWord(word string) string {
	match := urlRegexp.FindStringSubmatch(word)
	if len(match) > 1 {
		return match[1]
	}
	return word
}

// searchGitHub performs searches on GitHub based on the specified flags
func searchGitHub(query string) {
	if *oFlag {
		searchGitHubOrganizations(query, *maxFlag)
	}

	if *rFlag {
		searchGitHubRepositories(query, *maxFlag)
	}

	if *uFlag {
		searchGitHubUsers(query, *maxFlag)
	}
}

// searchGitLab performs searches on GitLab based on the specified flags
func searchGitLab(query string) {
	if *oFlag || *uFlag {
		searchGitLabGroupsAndUsers(query, *maxFlag)
	}

	if *rFlag {
		searchGitLabProjects(query, *maxFlag)
	}
}

// searchGitHubOrganizations searches for organizations on GitHub with the given query
func searchGitHubOrganizations(query string, maxResults int) {
	ctx := context.Background()
	client, err := createGithubClient(ctx)
	if err != nil {
		fmt.Printf("Error creating GitHub client: %s\n", err)
		return
	}

	opt := &github.SearchOptions{ListOptions: github.ListOptions{PerPage: maxResults}}
	results, _, err := client.Search.Organizations(ctx, query, opt)
	if err != nil {
		fmt.Printf("Error searching organizations: %s\n", err)
		return
	}

	orgLogins := make([]string, len(results.Organizations))
	for i, org := range results.Organizations {
		orgLogins[i] = *org.Login
	}

	printResults(fmt.Sprintf("GitHub organizations matching '%s'", query), orgLogins)
}

// searchGitHubRepositories searches for repositories on GitHub with the given query
func searchGitHubRepositories(query string, maxResults int) {
	ctx := context.Background()
	client, err := createGithubClient(ctx)
	if err != nil {
		fmt.Printf("Error creating GitHub client: %s\n", err)
		return
	}

	opt := &github.SearchOptions{ListOptions: github.ListOptions{PerPage: maxResults}}
	results, _, err := client.Search.Repositories(ctx, query, opt)
	if err != nil {
		fmt.Printf("Error searching repositories: %s\n", err)
		return
	}

	repoNames := make([]string, len(results.Repositories))
	for i, repo := range results.Repositories {
		repoNames[i] = *repo.FullName
	}

	printResults(fmt.Sprintf("GitHub repositories matching '%s'", query), repoNames)
}

// searchGitHubUsers searches for users on GitHub with the given query
func searchGitHubUsers(query string, maxResults int) {
	ctx := context.Background()
	client, err := createGithubClient(ctx)
	if err != nil {
		fmt.Printf("Error creating GitHub client: %s\n", err)
		return
	}

	opt := &github.SearchOptions{ListOptions: github.ListOptions{PerPage: maxResults}}
	results, _, err := client.Search.Users(ctx, query, opt)
	if err != nil {
		fmt.Printf("Error searching users: %s\n", err)
		return
	}

	userLogins := make([]string, len(results.Users))
	for i, user := range results.Users {
		userLogins[i] = *user.Login
	}

	printResults(fmt.Sprintf("GitHub users matching '%s'", query), userLogins)
}

// createGithubClient creates a new GitHub client with the given context and access token from the environment variable
func createGithubClient(ctx context.Context) (*github.Client, error) {
	token := os.Getenv("GITHUB_ACCESS_TOKEN")
	if token == "" {
		return nil, errors.New("GITHUB_ACCESS_TOKEN environment variable is not set")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	client.Client.Transport = &rateLimitedTransport{
		transport: tc.Transport,
		limiter:   rate.NewLimiter(rate.Every(10), 10),
	}

	return client, nil
}

// rateLimitedTransport is a custom transport that adds rate limiting to the GitHub API requests
type rateLimitedTransport struct {
	transport http.RoundTripper
	limiter   *rate.Limiter
}

func (t *rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.limiter.Wait(context.Background()); err != nil {
		return nil, err
	}

	return t.transport.RoundTrip(req)
}

// searchGitLabGroupsAndUsers searches for groups and/or users on GitLab with the given query
func searchGitLabGroupsAndUsers(query string, maxResults int) {
	client, err := createGitLabClient()
	if err != nil {
		fmt.Printf("Error creating GitLab client: %s\n", err)
		return
	}

	opt := &gitlab.SearchOptions{PerPage: maxResults}
	groups, _, err := client.Search.GroupSearch(query, opt)
	if err != nil {
		fmt.Printf("Error searching GitLab groups: %s\n", err)
		return
	}

	if *oFlag {
		groupFullPaths := make([]string, len(groups))
		for i, group := range groups {
			groupFullPaths[i] = group.FullPath
		}

		printResults(fmt.Sprintf("GitLab groups matching '%s'", query), groupFullPaths)
	}

	users, _, err := client.Search.UserSearch(query, opt)
	if err != nil {
		fmt.Printf("Error searching GitLab users: %s\n", err)
		return
	}

	if *uFlag {
		userUsernames := make([]string, len(users))
		for i, user := range users {
			userUsernames[i] = user.Username
		}

		printResults(fmt.Sprintf("GitLab users matching '%s'", query), userUsernames)
	}
}

// searchGitLabProjects searches for projects on GitLab with the given query
func searchGitLabProjects(query string, maxResults int) {
	client, err := createGitLabClient()
	if err != nil {
		fmt.Printf("Error creating GitLab client: %s\n", err)
		return
	}

	opt := &gitlab.SearchOptions{PerPage: maxResults}
	projects, _, err := client.Search.ProjectSearch(query, opt)
	if err != nil {
		fmt.Printf("Error searching GitLab projects: %s\n", err)
		return
	}

	projectFullPaths := make([]string, len(projects))
	for i, project := range projects {
		projectFullPaths[i] = project.PathWithNamespace
	}

	printResults(fmt.Sprintf("GitLab projects matching '%s'", query), projectFullPaths)
}

// createGitLabClient creates a new GitLab client with the access token from the environment variable
func createGitLabClient() (*gitlab.Client, error) {
	token := os.Getenv("GITLAB_ACCESS_TOKEN")
	if token == "" {
		return nil, errors.New("GITLAB_ACCESS_TOKEN environment variable is not set")
	}

	client, err := gitlab.NewClient(token)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// printResults displays the search results with a header or in a simple format based on the sFlag
func printResults(header string, results []string) {
	if *sFlag {
		for _, result := range results {
			fmt.Println(result)
		}
	} else {
		fmt.Printf("\n%s:\n", header)
		for _, result := range results {
			fmt.Printf("- %s\n", result)
		}
	}
}
