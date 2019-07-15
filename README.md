# dorky
A tool to automate dorking of Github/Shodan and a variety of other sources


# Idea

Dorky works with a variety of sources, contained in `/sources/` and scans, found in `/scans/`. These scan files are in JSON format and take the following properties:

## Sources
- Greymatter
- Shodan
- Github Search
- Gitlab Search

## Scans
Scans are stored as JSON, they contain the following properties:

| Item     | Details                                                                                             |
|----------|-----------------------------------------------------------------------------------------------------|
| sources  | array of sources that this module will run against                                                  |
| query    | the query to use for this source, where _target_ is replaced by the target name and _port_ the port |
| severity | the severity level for findings from informational to high                                          |
