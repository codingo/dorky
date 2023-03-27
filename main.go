package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v38/github"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"
)

type config struct {
	oFlag        bool
	rFlag        bool
	uFlag        bool
	maxFlag      int
	cFlag        bool
	ghFlag       bool
	glFlag       bool
	sFlag        bool
	vFlag        bool
}

var (
	flags = config{}
	urlRegexp    = regexp.MustCompile(`^https?://(?:www\.)?([^/]+)`)
	spaceRegexp  = regexp.MustCompile(`\s+`)
	wordPatterns = []string{"", "-", ""}
)

func init() {
	flag.BoolVar(&flags.oFlag, "o", false, "search for organization names")
	flag.BoolVar(&flags.rFlag, "r", false, "search for repository names")
	flag.BoolVar(&flags.uFlag, "u", false, "search for username matches")
	flag.IntVar(&flags.maxFlag, "max", 10, "maximum search results per category")
	flag.BoolVar(&flags.cFlag, "c", false, "clean input URLs")
	flag.BoolVar(&flags.ghFlag, "gh", false, "search only GitHub")
	flag.BoolVar(&flags.glFlag, "gl", false, "search only GitLab")
	flag.BoolVar(&flags.sFlag, "s", false, "simple output style for piping to another tool")
	flag.BoolVar(&flags.vFlag, "v", false, "enable verbose mode")
}

func main() {
	flag.Parse()
	validateFlags(flags)

	verbosePrint("Reading and cleaning words...\n")
	words := readAndCleanWords(flags)
	verbosePrint("Words cleaned.\n")

	verbosePrint("Searching platforms...\n")
	searchPlatforms(words, flags)
	verbosePrint("Platform search completed.\n")
}

func validateFlags(cfg config) {
	if !(cfg.oFlag || cfg.rFlag || cfg.uFlag) {
		fmt.Println("At least one search flag (-o, -r, or -u) must be specified")
		os.Exit(1)
	}
	verbosePrint("Flags validated.\n")
}

func verbosePrint(format string, a ...interface{}) {
	if flags.vFlag {
		fmt.Printf(format, a...)
	}
}

func readAndCleanWords(cfg config) map[string]struct{} {
	words := make(map[string]struct{})
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		processWord(word, words, cfg)
	}
	checkScannerError(scanner)

	return words
}

func processWord(word string, words map[string]struct{}, cfg config) {
	if cfg.cFlag {
		word = cleanWord(word)
	}

	addWordToMap(words, word)
	word = removeWhitespace(word)
	wordLines := strings.Split(word, "\n")

	for _, w := range wordLines {
		addWordToMap(words, w)
	}
}

func addWordToMap(words map[string]struct{}, word string) {
	if _, exists := words[word]; !exists {
		words[word] = struct{}{}
	}
}

func checkScannerError(scanner *bufio.Scanner) {
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading stdin: %s\n", err)
		os.Exit(1)
	}
}

func searchPlatforms(words map[string]struct{}, cfg config) {
	for word := range words {
		verbosePrint("Searching GitHub for word: %s\n", word)
		if !cfg.glFlag {
			searchGitHub(word, cfg)
		}

		verbosePrint("Searching GitLab for word: %s\n", word)
		if !cfg.ghFlag {
			searchGitLab(word, cfg)
		}
	}
}

func cleanWord(word string) string {
	match := urlRegexp.FindStringSubmatch(word)
	if len(match) > 1 {
		return match[1]
	}
	return word
}

func removeWhitespace(word string) string {
	removedSpaces := spaceRegexp.ReplaceAllString(word, "")
	withHyphens := spaceRegexp.ReplaceAllString(word, "-")
	return removedSpaces + "\n" + withHyphens
}

func searchGitHub(query string, cfg config) {
	if cfg.oFlag {
		searchGitHubOrganizations(query, cfg.maxFlag)
	}

	if cfg.rFlag {
		searchGitHubRepositories(query, cfg.maxFlag)
	}

	if cfg.uFlag {
		searchGitHubUsers(query, cfg.maxFlag)
	}
}

func searchGitLab(query string, cfg config) {
	if cfg.oFlag || cfg.uFlag {
		searchGitLabGroupsAndUsers(query, cfg.maxFlag)
	}

	if cfg.rFlag {
		searchGitLabProjects(query, cfg.maxFlag)
	}
}

func searchGitHubOrganizations(query string, maxResults int) {
	ctx := context.Background()
	client, err := createGithubClient(ctx)
	if err != nil {
		fmt.Printf("Error creating GitHub client: %s\n", err)
		return
	}

	opt := &github.SearchOptions{ListOptions: github.ListOptions{PerPage: maxResults}}
	results, _, err := client.Search.Users(ctx, "type:org "+query, opt)
	if err != nil {
		fmt.Printf("Error searching organizations: %s\n", err)
		return
	}

	orgLogins := make([]string, len(results.Users))
	for i, org := range results.Users {
		orgLogins[i] = *org.Login
	}

	printResults(fmt.Sprintf("GitHub organizations matching '%s'", query), orgLogins)
}

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

func searchGitHubUsers(query string, maxResults int) {
	ctx := context.Background()
	client, err := createGithubClient(ctx)
	if err != nil {
		fmt.Printf("Error creating GitHub client: %s\n", err)
		return
	}

	opt := &github.SearchOptions{ListOptions: github.ListOptions{PerPage: maxResults}}
	results, _, err := client.Search.Users(ctx, "type:user "+query, opt)
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

func createGithubClient(ctx context.Context) (*github.Client, error) {
	token := os.Getenv("GITHUB_ACCESS_TOKEN")
	if token == "" {
		return nil, errors.New("GITHUB_ACCESS_TOKEN environment variable is not set")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	tc.Transport = &rateLimitedTransport{
		transport: tc.Transport,
		limiter:   rate.NewLimiter(rate.Every(10), 10),
	}

	client := github.NewClient(tc)

	return client, nil
}

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

func searchGitLabGroupsAndUsers(query string, maxResults int) {
	client, err := createGitLabClient()
	if err != nil {
		fmt.Printf("Error creating GitLab client: %s\n", err)
		return
	}

	opt := &gitlab.ListGroupsOptions{Search: gitlab.String(query), ListOptions: gitlab.ListOptions{PerPage: maxResults}}
	groups, _, err := client.Groups.ListGroups(opt)
	if err != nil {
		fmt.Printf("Error searching GitLab groups: %s\n", err)
		return
	}

	if flags.oFlag {
		groupFullPaths := make([]string, len(groups))
		for i, group := range groups {
			groupFullPaths[i] = group.FullPath
		}

		printResults(fmt.Sprintf("GitLab groups matching '%s'", query), groupFullPaths)
	}

	users, _, err := client.Users.ListUsers(&gitlab.ListUsersOptions{Search: gitlab.String(query), ListOptions: gitlab.ListOptions{PerPage: maxResults}})
	if err != nil {
		fmt.Printf("Error searching GitLab users: %s\n", err)
		return
	}

	if flags.uFlag {
		userUsernames := make([]string, len(users))
		for i, user := range users {
			userUsernames[i] = user.Username
		}

		printResults(fmt.Sprintf("GitLab users matching '%s'", query), userUsernames)
	}
}

func searchGitLabProjects(query string, maxResults int) {
	client, err := createGitLabClient()
	if err != nil {
		fmt.Printf("Error creating GitLab client: %s\n", err)
		return
	}

	opt := &gitlab.ListProjectsOptions{Search: gitlab.String(query), ListOptions: gitlab.ListOptions{PerPage: maxResults}}
	projects, _, err := client.Projects.ListProjects(opt)
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

func printResults(header string, results []string) {
	if flags.sFlag {
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
