# Dorky

[![License](https://img.shields.io/badge/license-GPL3-_red.svg)](https://www.gnu.org/licenses/gpl-3.0.en.html) [![Twitter](https://img.shields.io/badge/twitter-@codingo__-blue.svg)](https://twitter.com/codingo_)

Dorky is a command-line tool that searches GitHub and GitLab for matches in organization names, repository names, and usernames based on a list of input words. This tool can be helpful in identifying potential targets for security assessments, finding interesting projects, and discovering new organizations and users on GitHub and GitLab.

## Example

```bash
echo "codingo\ncodingo dot com" | dorky -o -r -u -c
```

This will search for organization names, repository names, and usernames on both GitHub and GitLab based on the cleaned input words:

```
codingo
codingodotcom
codingo-dot-com
```

## Installation

1. Clone the repository:

```bash
git clone https://github.com/codingo/dorky.git
```

2. Set your GitHub and/or GitLab access tokens as environment variables:

```bash
export GITHUB_ACCESS_TOKEN=your-github-access-token
export GITLAB_ACCESS_TOKEN=your-gitlab-access-token
```

3. Build the Dorky tool:

```bash
go build -o dorky
```

## Docker Instructions

### Requirements

- Docker
- GITHUB_ACCESS_TOKEN and GITLAB_ACCESS_TOKEN environment variables

1. Build the Docker image:

   ```bash
   docker build -t dorky .
   ```

2. Run the Docker container:

   ```bash
   docker run --rm -it -e GITHUB_ACCESS_TOKEN=your-github-token -e GITLAB_ACCESS_TOKEN=your-gitlab-token dorky
   ```

   Replace `your-github-token` and `your-gitlab-token` with your GitHub and GitLab access tokens, respectively.

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
- `-s`: Simple output style for piping to another tool
- `-v`: Enable verbose mode for more detailed output

By default, the tool searches both GitHub and GitLab based on the provided access tokens. If both tokens are set, both platforms will be searched. If only one token is set, only that platform will be searched.

## Dependencies

- google/go-github/v38
- xanzy/go-gitlab
- golang.org/x/oauth2
- golang.org/x/time/rate
