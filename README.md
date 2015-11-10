# Git Hound

Git plugin that helps prevent sensitive data from being committed by sniffing potential commits against regular expressions from a local `.githound.yml` file.

## Installation
To install Hound, please use `go get`. If you don't have Go installed, [get it here](https://golang.org/dl/). If you would like to grab a precompiled binary, head over to the [releases](https://github.com/ezekg/git-hound/releases) page. The precompiled Hound binaries have no external dependencies.

```
go get github.com/ezekg/git-hound
```

**Alias `git add` inside `.bash_profile`:** _(optional)_
```bash
alias git='_() { if [[ "$1" == "add" ]]; then git-hound "$@"; else git "$@"; fi }; _'
```

## Usage
```bash
git hound add <files>
git add <files> # When using the optional alias above
```

## Option flags

| Flag           | Type   | Default         | Usage                                      |
| :------------- | :----- | :-------------- | :----------------------------------------- |
| `-no-color`    | bool   | `false`         | Disable color output                       |
| `-config=file` | string | `.githound.yml` | Hound config file                          |
| `-bin=file`    | string | `git`           | Executable binary to use for `git` command |

## Example `.githound.yml`
Please see [Go's regular expression syntax documentation](https://golang.org/pkg/regexp/syntax/) for usage options.

```yaml
# Output warning on match but continue
warn:
  - '(?i)user(name)?\W*[:=,]\W*.+$'
# Fail immediately upon match
fail:
  - '(?i)db_(user(name)?|pass(word)?|name)\W*[:=,]\W*.+$'
  - '(?i)pass(word)?\W*[:=,]\W*.+$'
# Skip on matched filename
skip:
  - '\.example$'
  - '\.sample$'
```
