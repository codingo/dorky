# dorky
A tool to automate dorking of Github/Shodan and a variety of other sources


# Idea

Dorky works with a variety of sources, contained in `/sources/` and scans, found in `/scans/`. These scan files are in JSON format and take the following properties:

## Sources
- GreyHatWarfare
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

An example scan would be:

```
{  
   "sources":[  
      "gitlab",
      "github"
   ],
   "title":"Direct Credentials",
   "description":"Find credentials against an associated domain name.",
   "severity":1,
   "queries":{"\"_taget_\" password\"",
              "\"_taget_\" pass\""}
}
```

## Command Line

The command line takes the following arguments:

| Argument     | Description                      |
|--------------|----------------------------------|
| -s -severity | The severity level to run        |
| -t           | The target to run a scan against |
| -o           | File to save output to           |

# Examples

Using the above JSON example, and the command line input, an example search would be:

```
dorky -t example.com -severity 1
```

This would run all sources in `/sources/` that have a severity of 1 or higher and output the results:

```
DORKY - By hakluke, sml and codingo
2 Dorks with results:
[Direct Credentials] https://github.com/example-link
[Direct Credentials] https://gitlab.com/example-link
```
