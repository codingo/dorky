# Dorky

Dorky is a command-line tool that searches GitHub and GitLab for matches in organization names, repository names, and usernames based on a list of input words.

## Installation

1. Clone the repository:

```
git clone https://github.com/username/dorky.git
```

2. Set your GitHub and/or GitLab access tokens as environment variables:

```
export GITHUB_ACCESS_TOKEN=your-github-access-token
export GITLAB_ACCESS_TOKEN=your-gitlab-access-token
```

3. Build the Dorky tool:

```
go build -o dorky
```

## Usage

Pipe a list of words to the Dorky tool and use the appropriate flags to specify the search categories and platforms:

```
cat wordlist.txt | ./dorky -uro -gh
```

Available flags:

- `-o`: Search for organization names (or groups in GitLab)
- `-r`: Search for repository names (or projects in GitLab)
- `-u`: Search for username matches
- `-max`: Set the maximum number of search results per category (default: 10)
- `-c`: Clean input URLs, turning them into words before performing searches
- `-gh`: Search only GitHub
- `-gl`: Search only GitLab

By default, the tool searches both GitHub and GitLab based on the provided access tokens. If both tokens are set, both platforms will be searched. If only one token is set, only that platform will be searched.
```

With these updates, the Dorky tool now supports cleaning URLs, de-duplication, and searching both GitHub and GitLab. The error handling and code patterns have also been improved. Feel free to use this version as a starting point for your own implementation and make any additional changes as needed.
