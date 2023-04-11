## Bootstrap github repository

[![release](https://img.shields.io/github/release/DoodleScheduling/gh-bootstrap-repository/all.svg)](https://github.com/DoodleScheduling/gh-bootstrap-repository/releases)
[![release](https://github.com/doodlescheduling/gh-bootstrap-repository/actions/workflows/release.yaml/badge.svg)](https://github.com/doodlescheduling/gh-bootstrap-repository/actions/workflows/release.yaml)
[![report](https://goreportcard.com/badge/github.com/DoodleScheduling/gh-bootstrap-repository)](https://goreportcard.com/report/github.com/DoodleScheduling/gh-bootstrap-repository)
[![Coverage Status](https://coveralls.io/repos/github/DoodleScheduling/gh-bootstrap-repository/badge.svg?branch=master)](https://coveralls.io/github/DoodleScheduling/gh-bootstrap-repository?branch=master)
[![license](https://img.shields.io/github/license/DoodleScheduling/gh-bootstrap-repository.svg)](https://github.com/DoodleScheduling/gh-bootstrap-repository/blob/master/LICENSE)

Github does not provide an easy way to create a repository with predefined settings from another repository.
A template repository for instance is only useful to copy the contents. No repository settings are copied.

This gh cli extension creates a new repository based on another one (A normal or template one).

It copies the following:

* The content with a single commit
* Basic repository settings
* Branch protections
* Organization teams (collaborators)
* Topics

### Install
```
gh extension install DoodleScheduling/gh-bootstrap-repository
```

### Usage
The following command will attempt to create a repository called `myorg/my-repository`
with all the settings and content taken from `myorg/from-repository`.

```
gh bootstrap-repository myorg/my-repository myorg/from-repository
```

Command line help:
```
expects [repository-name] [origin-repository] as arguments

Usage:
  gh-bootstrap-repository [repository-name] [origin-repository] [flags]

Flags:
  -h, --help   help for gh-bootstrap-repository

```
