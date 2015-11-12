# Git Hound
[![Travis](https://img.shields.io/travis/ezekg/git-hound.svg?style=flat-square)](https://travis-ci.org/ezekg/git-hound)
[![Code Climate](https://img.shields.io/codeclimate/github/ezekg/git-hound.svg?style=flat-square)](https://codeclimate.com/github/ezekg/git-hound)

Git plugin that helps prevent sensitive data from being committed by sniffing potential commits against regular expressions from a local `.githound.yml` file.

## How does it work?
Upon commit, it runs the output of `git diff -U0 --staged` through the Hound, which matches every _added_ or _modified_ line against your provided list of regular expressions. This runs in O(m*n) time (where m is the number of lines and n is the number of patterns), so be sure to commit often. But you should be doing that anyway, right?

## Installation
To install Hound, please use `go get`. If you don't have Go installed, [get it here](https://golang.org/dl/). If you would like to grab a precompiled binary, head over to the [releases](https://github.com/ezekg/git-hound/releases) page. The precompiled Hound binaries have no external dependencies.

```
go get github.com/ezekg/git-hound
```

**Alias `git commit` inside `~/.bash(rc|_profile)`:** _(optional)_
```bash
alias git='_() { if [[ "$1" == "commit" ]]; then git-hound "$@"; else git "$@"; fi }; _'
```

## Usage
```bash
git hound commit ...
git commit ... # When using the optional alias above
```

## Option flags
These flags should be included inside of the `git` alias, if used.

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
