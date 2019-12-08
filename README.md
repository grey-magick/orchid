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

# `orchid` - Kubernetes inspired Object Store

`orchid` is an experiment on building a **Kubernetes inspired Object Store**, offering an equivalent 
of Kubernetes *CustomResourceDefinition* as basic primitive and using Postgres to store user defined 
resources. Postgres will not only be responsible for managing CRUD operations related to those 
resources but, more importantly, a transactional engine that can be used to leverage a transactional 
API yet to be defined.

As a design goal, `orchid` should support user interactions through `kubectl`; this is important to
leverage already existing knowledge and tooling.

It is expected that any serious implementation might require a work group or SIG to be formed in 
order to coordinate the development of shared components used by this project and Kubernetes, in the 
case `orchid` catches the community's attention.
