<p align="center">
    <img alt="orchid logo" src="./assets/logo/orchid.png">
</p>
<p align="center">
    <a alt="GoReport" href="https://goreportcard.com/report/github.com/isutton/orchid">
        <img alt="GoReport" src="https://goreportcard.com/badge/github.com/isutton/orchid">
    </a>
    <a alt="Code Coverage" href="https://codecov.io/gh/isutton/orchid">
        <img alt="Code Coverage" src="https://codecov.io/gh/isutton/orchid/branch/master/graph/badge.svg">
    </a>
    <a href="https://godoc.org/github.com/isutton/orchid/pkg/orchid">
        <img alt="GoDoc Reference" src="https://godoc.org/github.com/isutton/orchid/pkg/orchid?status.svg">
    </a>
    <a alt="CI Status" href="https://travis-ci.com/isutton/orchid">
        <img alt="CI Status" src="https://travis-ci.com/isutton/orchid.svg?branch=master">
    </a>
<!--
    <a alt="Docker-Cloud Build Status" href="https://hub.docker.com/r/isutton/orchid">
        <img alt="Docker-Cloud Build Status" src="https://img.shields.io/docker/cloud/build/isutton/orchid.svg">
    </a>
  -->
</p>

# `orchid`

`orchid` is an experiment on building a Kubernetes compatible API server, offering
CustomResourceDefinition as basic primitive. The database of choice to store data associated with
those definitions is Postgres, which will be responsible not only for managing CRUD operations
related to those resources but, more importantly, a transactional engine that can be used to leverage
a transactional API yet to be defined.

As a design goal, this experiment implements only the basics of an API server to prove the concept.
Eventual implementation might require a SIG to be formed to coordinate shared components used by this
project and Kubernetes, in the case `orchid` catches the attention of the community.
