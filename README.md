<h1 align="center">dashboard</h1>

<p align="center">Create a badge dashboard from a yaml file.</p>

<p  align="center">
 <!-- <a href="https://dev.azure.com/cugu/dfir/_build?definitionId=3&_a=summary"><img src="https://img.shields.io/azure-devops/build/cugu/dfir/1" alt="build" /></a>
 <a href="https://codecov.io/gh/cugu/fslib"><img src="https://codecov.io/gh/cugu/fslib/branch/master/graph/badge.svg" alt="coverage" /></a> -->
 <a href="https://goreportcard.com/report/github.com/cugu/dashboard"><img src="https://goreportcard.com/badge/github.com/cugu/dashboard" alt="report" /></a>
 <!-- <a href="https://godoc.org/github.com/cugu/dashboard"><img src="https://godoc.org/github.com/cugu/dashboard?status.svg" alt="doc" /></a> -->
</p>

# Installation
``` sh
go get github.com/cugu/dashboard
```

# Usage
``` sh
dashboard example-projects.yml
```

## projects.yml

``` yaml
table:
  - name: icon
    enabled: ["icon"]
  - name: link
    enabled: ["link"]
  - name: build
    enabled: ["azure", "gitlab"]
    disabled: ["travis"]
  - name: version
    enabled: ["github-version"]

categories:
  - name: index
    projects:
    - url: https://github.com/cugu/afro
      enable: ["travis"]
    - url: https://github.com/cugu/apfs.ksy
      disable: ["version", "build"]
```