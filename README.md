# Git Hound
[![Travis](https://img.shields.io/travis/ezekg/git-hound.svg?style=flat-square)](https://travis-ci.org/ezekg/git-hound)
[![Code Climate](https://img.shields.io/codeclimate/github/ezekg/git-hound.svg?style=flat-square)](https://codeclimate.com/github/ezekg/git-hound)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/ezekg/git-hound)

Hound is a Git plugin that helps prevent sensitive data from being committed into a repository by sniffing potential commits against PCRE regular expressions.

## How does it work?
Upon commit, it runs the output of `git diff -U0 --staged` through the Hound, which matches every _added_ or _modified_ line against your provided list of regular expressions from a local `.githound.yml` file.

## Installation
To install Hound, please use `go get`. If you don't have Go installed, [get it here](https://golang.org/dl/). If you would like to grab a precompiled binary, head over to the [releases](https://github.com/ezekg/git-hound/releases) page. The precompiled Hound binaries have no external dependencies.

```
go get github.com/ezekg/git-hound
```

## Compiling
To compile for your operating system, simply run the following from the root of the project directory:
```bash
go install
```

To compile for all platforms using [`goxc`](https://github.com/laher/goxc), run the following:
```bash
goxc -pv={VERSION} -d=releases/
```

## Usage

#### Commit
```bash
# Scan changes since last commit and pass to git-commit when clean
git hound commit …
```

#### Sniff
```bash
# Scan changes since last commit
git hound sniff HEAD

# Scan entire history of repository
git hound sniff
```

## Option flags

| Flag           | Type   | Default         | Usage                                      |
|:---------------|:-------|:----------------|:-------------------------------------------|
| `-no-color`    | bool   | `false`         | Disable color output                       |
| `-config=file` | string | `.githound.yml` | Hound config file                          |
| `-bin=file`    | string | `git`           | Executable binary to use for `git` command |

## Example `.githound.yml`

```yaml
# Output warning on match but continue
warn:
  - '(?i)user(name)?\W*[:=,]\W*.+$'
  - '\/Users\/\w+\/'
# Fail immediately upon match
fail:
  - '(?!.*[\s])(?=.*[A-Za-z])(?=.*[0-9])(?=.*[!@#$&*])?.{16,}'
  - '(?i)db_(user(name)?|pass(word)?|name)\W*[:=,]\W*.+$'
  - '(?i)pass(word)?\W*[:=,]\W*.+$'
# Skip on matched filename
skip:
  - '\.example$'
  - '\.sample$'
```
