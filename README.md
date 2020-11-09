# RESTful Couchbase

[![Build status](http://travis-ci.org/mramshaw/RESTful-Couchbase.svg?branch=master)](http://travis-ci.org/mramshaw/RESTful-Couchbase)
[![Coverage Status](http://codecov.io/github/mramshaw/RESTful-Couchbase/coverage.svg?branch=master)](http://codecov.io/github/mramshaw/RESTful-Couchbase?branch=master)
[![Go Report Card](http://goreportcard.com/badge/github.com/mramshaw/RESTful-Couchbase?style=flat-square)](http://goreportcard.com/report/github.com/mramshaw/RESTful-Couchbase)
[![GitHub release](http://img.shields.io/github/v/release/mramshaw/RESTful-Couchbase?style=flat-square)](http://github.com/mramshaw/RESTful-Couchbase/releases)

An experiment with using [Couchbase](http://docs.couchbase.com/home/) as
a drop-in replacement for PostgreSQL.

## Contents

The contents are as follows:

* [Rationale](#rationale)
* [Features](#features)
* [Couchbase](#couchbase)
    * [Eventually Consistent](#eventually-consistent)
    * [Views](#views)
    * [Caveats](#caveats)
    * [Transactions, Sagas and locking](#transactions-sagas-and-locking)
    * [Getting familiar with Couchbase](02-Couchbase-Introduction.md)
* [Couchbase Performance Tips](03-Couchbase-Performance-Tips.md)
    * [Query by KEYS rather than by id](03-Couchbase-Performance-Tips.md#query-by-keys-rather-than-by-id)
    * [Specify "AdHoc(false)" to cache queries](03-Couchbase-Performance-Tips.md#specify-adhocfalse-to-cache-queries)
    * [Track prepared statement performance](03-Couchbase-Performance-Tips.md#track-prepared-statement-performance)
* [Operations](#operations)
    * [To Build](#to-build)
    * [To Run](#to-run)
    * [For testing](#for-testing)
    * [See what's running](#see-whats-running)
    * [View the build and/or execution logs](#view-the-build-andor-execution-logs)
    * [To Shutdown](#to-shutdown)
    * [Clean up](#clean-up)
    * [Results](#results)
* [Versions](#versions)
* [Reference](#reference)
    * [Couchbase BLOG](#couchbase-blog)
* [To Do](#to-do)

## Rationale

This builds on my [RESTful-Recipes](http://github.com/mramshaw/RESTful-Recipes) repo,
which stores data in [PostgreSQL](http://www.postgresql.org/).

All dependencies are handled via [Docker](http://www.docker.com/products/docker) and [docker-compose](https://github.com/docker/compose).

TDD (Test-Driven Development) is implemented; the build will fail if the tests do not pass.

Likewise the build will fail if either __golint__ or __go vet__ fails.

Enforces industry-standard __gofmt__ code formatting.

All testing can be done with [curl](CURLs.txt).

We will use parameterized N1QL to prevent SQL injection.

## Features

- uses [Gorilla MUX](http://github.com/Gorilla/mux)
- uses [Go couchbase driver](http://blog.couchbase.com/go-sdk-1.0-ga/)

## Couchbase

Couchbase is a document-oriented database based on the JSON document model.

What would be referred to as a __table__ or __row__ in a relational database
is referred to as a __document__ in Couchbase. Like other NoSQL databases,
documents may have embedded elements. In this regard, Couchbase documents are
more akin to object storage than relational database rows or records.

What would be referred to as a __database__ in a relational database seems
to be referred to as a __bucket__ in Couchbase.

Likewise, what would normally be referred to as a __server__ seems to be
referred to as a __cluster__ in Couchbase.

As with other NoSQL databases (such as DynamoDB), schemas are flexible.

Unusually, features master-master replication (all nodes are identical).

N1QL (the SQL-like Couchbase query language) operates on JSON documents,
returning JSON documents.

Similiar to __redis__ and __Cassandra__, data may be assigned arbitrary
expiry times.

Unlike __Cassandra__, Couchbase has support for __joins__.

Couchbase is packaged with an Admin Console GUI. Other NoSQL solutions
(such as __MongoDB__ and __Cassandra__) apparently are not packaged with
administrative consoles (although third-party consoles are available).

#### Eventually Consistent

The Couchbase documentation seem to be somewhat inconsistent. For instance:

> The basic storage and indexing sequence is:
>
> 1. A document is stored within the cluster. Initially the document is stored only in RAM.
>
> 2. The document is communicated to the indexer through replication to be indexed by views.
>
> This sequence means that the view results are eventually consistent with what is stored in
> memory based on the latency in replication of the change to the indexer. It is possible to
> write a document to the cluster and access the index without the newly written document
> appearing in the generated view.
>
> Conversely, documents that have been stored with an expiry may continue to be included
> within the view until the document has been removed from the database by the expiry pager.
>
> Couchbase Server supports the Observe command, which enables the current state of a document
> and whether the document has been replicated to the indexer or whether it has been considered
> for inclusion in an index.
>
> When accessing a view, the contents of the view are asynchronous to the stored documents.
> In addition, the creation and updating of the view is subject to the `stale` parameter.
> This controls how and when the view is updated when the view content is queried.

    http://docs.couchbase.com/server/6.0/learn/views/views-store-data.html#document-storage-and-indexing-sequence

[Note that the above makes no mention of the whether or not the document has been persisted to disk.]

For more on the `stale` parameter:

    http://docs.couchbase.com/server/6.0/learn/views/views-operation.html#index-stale

Note the following:

> * stale=false
>
> The index is updated before you execute the query, making sure that any documents updated and
> persisted to disk are included in the view. The client will wait until the index has been
> updated before the query has executed and, therefore, the response will be delayed until the
> updated index is available.

[This means that it is possible to request __only__ data that has been persisted to disk.]

#### Views

Couchbase __views__ are described as "eventually consistent".

> Views are eventually consistent compared to the underlying stored documents.
> Documents are included in views when the document data is persisted to disk.

    http://docs.couchbase.com/server/6.0/learn/views/views-intro.html

Contrast the above with the following:

> Views are updated as the document data is updated in memory. There may be a delay between the
> document being created or updated and the document being updated within the view depending
> on the client-side query parameters.

    http://docs.couchbase.com/server/6.0/learn/views/views-operation.html

[Presumably this reflects how the `stale` parameter is set.]

What this ___may___ mean in practice is that Couchbase views are not updated
until the underlying documents they reference are persisted to disk - and ___not___
when the document updates are acknowledged and stored in memory (see [Caveats](#caveats)
below). [Presumably this is also when the indexes the views reference that track
these documents are updated.]

The exception to this is __DCP__ (Database Change Protocol) - which is available for stream-based views:

> With DCP, data does not need to be persisted to disk before retrieving it with a view query.

    http://docs.couchbase.com/server/6.0/learn/views/views-streaming.html

#### Caveats

Couchbase has a __memory-first__ architecture, which means that the results
of write operations are acknowledged when stored in memory (they are then
queued to be asynchronously written to disk and/or then replicated to another
node). So if an operation is written to memory, and the system shuts down
immediately afterwards, then that operation may not persist.

This is the __default__ behaviour, and potentially violates the __D__ of
ACID transactions. However, this behaviour can be over-ridden (at a small
performance cost) if durability requirements are critical.

Couchbase has a maximum capacity of 20 MB per document (probably not an
issue, but worth bearing in mind).

Document ids must be unique for the bucket in Couchbase (different document
types must have unique document ids). Also, each Document id (key) must be
a string in Couchbase (i.e. `"1"` instead of `1`). For this reason, it
seems to be normal practice to use a compound id (such as `"user::5"`); a
variation on this seems to be to embed a __type__ field within the document
(such as `"type": "user"`).

[The document id may be accessed via the document metadata: `META().id`.
 If using N1QL this value may be returned, however it is unclear to me
 how this value may be accessed from the Couchbase GOCB driver API.]

#### Transactions, Sagas and locking

Couchbase provides ACID transactions for single document operations, but multi-document transactions (or __sagas__) are not yet
natively supported (alpha third-party solutions are available).

Couchbase offers both [optimistic and pessimistic locking](http://docs.couchbase.com/go-sdk/1.5/concurrent-mutations-cluster.html).

Locking in Couchbase is implemented by CAS (Compare And Swap). This is basically a hash or digest, and will indicate if the document
in question has been mutated (if it __has__ and the CAS has been supplied, then the update will fail). Alternatively, there are
__lock__ and __unlock__ primitives (as well as __get\_and\_lock__). The lock time may be specified. Mutating the document will
also serve to unlock it.

#### Getting familiar with Couchbase

Refer to [Couchbase Introduction](02-Couchbase-Introduction.md) for a quick guide to getting started with Couchbase.

## Couchbase Performance Tips

Refer to [Couchbase Performance Tips](03-Couchbase-Performance-Tips.md) for some general performance tips.

## Operations

We will use __Docker__ and __docker-compose__ to build and test our application.

#### To Build:

The command to run:

    $ docker-compose up

This should look as follows:

```bash
$ docker-compose up
Creating network "restfulcouchbase_couchnet" with the default driver
Creating restfulcouchbase_couchbase_1_d6da7719d982 ... done
Creating restfulcouchbase_golang_1_cb1241403038    ... done
Attaching to restfulcouchbase_couchbase_1_84f81eb2a871, restfulcouchbase_golang_1_6b5fa034bf5a
couchbase_1_84f81eb2a871 | + set -m
couchbase_1_84f81eb2a871 | + sleep 10
couchbase_1_84f81eb2a871 | + /entrypoint.sh couchbase-server
couchbase_1_84f81eb2a871 | Starting Couchbase Server -- Web UI available at http://<ip>:8091
couchbase_1_84f81eb2a871 | and logs available in /opt/couchbase/var/lib/couchbase/logs
couchbase_1_84f81eb2a871 | + /opt/couchbase/bin/couchbase-cli cluster-init -c localhost --cluster-username halcouch --cluster-password couchpass --services data,index,query
couchbase_1_84f81eb2a871 | SUCCESS: Cluster initialized
couchbase_1_84f81eb2a871 | + /opt/couchbase/bin/couchbase-cli bucket-create -c localhost --username halcouch --password couchpass --bucket recipes --bucket-type couchbase --bucket-ramsize 100 --enable-flush=1
couchbase_1_84f81eb2a871 | SUCCESS: Bucket created
couchbase_1_84f81eb2a871 | + fg 1
couchbase_1_84f81eb2a871 | /entrypoint.sh couchbase-server
golang_1_6b5fa034bf5a | GOPATH=/go GOOS=linux GOARCH=amd64 gofmt -d -e -s -w *.go
golang_1_6b5fa034bf5a | GOPATH=/go GOOS=linux GOARCH=amd64 gofmt -d -e -s -w application/*.go
golang_1_6b5fa034bf5a | GOPATH=/go GOOS=linux GOARCH=amd64 gofmt -d -e -s -w recipes/*.go
golang_1_6b5fa034bf5a | GOPATH=/go GOOS=linux GOARCH=amd64 gofmt -d -e -s -w test/*.go
golang_1_6b5fa034bf5a | GOPATH=/go GOOS=linux GOARCH=amd64 golint -set_exit_status *.go
golang_1_6b5fa034bf5a | GOPATH=/go GOOS=linux GOARCH=amd64 golint -set_exit_status ./...
golang_1_6b5fa034bf5a | GOPATH=/go GOOS=linux GOARCH=amd64 go tool vet *.go
golang_1_6b5fa034bf5a | GOPATH=/go GOOS=linux GOARCH=amd64 go tool vet application/*.go
golang_1_6b5fa034bf5a | GOPATH=/go GOOS=linux GOARCH=amd64 go tool vet recipes/*.go
golang_1_6b5fa034bf5a | GOPATH=/go GOOS=linux GOARCH=amd64 go tool vet test/*.go
golang_1_6b5fa034bf5a | GOPATH=/go GOOS=linux GOARCH=amd64 go test -v test
golang_1_6b5fa034bf5a | === RUN   TestEmptyTables
golang_1_6b5fa034bf5a | --- PASS: TestEmptyTables (0.55s)
golang_1_6b5fa034bf5a | === RUN   TestGetNonExistentRecipe
golang_1_6b5fa034bf5a | --- PASS: TestGetNonExistentRecipe (0.41s)
golang_1_6b5fa034bf5a | === RUN   TestCreateRecipe
golang_1_6b5fa034bf5a | --- PASS: TestCreateRecipe (0.41s)
golang_1_6b5fa034bf5a | === RUN   TestGetRecipe
golang_1_6b5fa034bf5a | --- PASS: TestGetRecipe (0.39s)
golang_1_6b5fa034bf5a | === RUN   TestUpdatePutRecipe
golang_1_6b5fa034bf5a | --- PASS: TestUpdatePutRecipe (0.37s)
golang_1_6b5fa034bf5a | === RUN   TestUpdatePatchRecipe
golang_1_6b5fa034bf5a | --- PASS: TestUpdatePatchRecipe (0.34s)
golang_1_6b5fa034bf5a | === RUN   TestDeleteRecipe
golang_1_6b5fa034bf5a | --- PASS: TestDeleteRecipe (0.37s)
golang_1_6b5fa034bf5a | === RUN   TestAddRating
golang_1_6b5fa034bf5a | --- PASS: TestAddRating (0.39s)
golang_1_6b5fa034bf5a | === RUN   TestSearch
golang_1_6b5fa034bf5a | --- PASS: TestSearch (4.43s)
golang_1_6b5fa034bf5a | PASS
golang_1_6b5fa034bf5a | ok  	test	9.691s
golang_1_6b5fa034bf5a | GOPATH=/go GOOS=linux GOARCH=amd64 go build -v -o restful_couchbase main.go
golang_1_6b5fa034bf5a | restful_couchbase has been compiled
golang_1_6b5fa034bf5a | ./restful_couchbase
golang_1_6b5fa034bf5a | type 'make serve' to run
```

Once `make` indicates that `restful_couchbase` has been compiled, can change [docker-compose.yml](docker-compose.yml) as follows:

1) Comment `command: bash -c "sleep 15; make"`

2) Uncomment `#command: bash -c "sleep 15; ./restful_couchbase"`

#### To Run

The command to run:

    $ docker-compose up -d

For the first run, there will be a warning if `mramshaw4docs/golang-couchbase:1.15.4` has not been built.

This image will contain all of the Go dependencies and should only need to be built once.

A successful `golang` startup should show the following as the last line of `docker-compose logs golang`:

    golang_1    | 2019/03/01 19:05:05 Now serving recipes ...

If this line does not appear, repeat the `docker-compose up -d` command (there is no penalty for this).

#### For testing:

[Optional] Start couchbase:

    $ docker-compose up -d couchbase

Start golang [if couchbase is not running, this step will start it]:

    $ docker-compose run --publish 80:8100 golang make

Successful startup will be indicated as follows:

    2019/03/01 19:05:05 Now serving recipes ...

This should make the web service available at:

    http://localhost/v1/recipes

Once the service is running, it is possible to `curl` it. Check [CURLs.txt](CURLs.txt) for examples.

#### See what's running:

The command to run:

    $ docker ps

#### View the build and/or execution logs

The command to run:

    $ docker-compose logs golang

#### To Shutdown:

The command to run:

    $ docker-compose down

Clean up docker volumes as follows:

	$ docker volume prune

#### Clean Up

The command to run:

    $ docker-compose run golang make clean

Clean up docker image as follows:

	$ docker rmi mramshaw4docs/golang-couchbase:1.15.4

#### Results

Due to the ad-hoc nature of NoSQL documents, the code is somewhat more complicated
than would be the case with relational databases; however as I am using Couchbase
as a drop-in replacement for PostgreSQL this is hardly a fair comparison. But for
learning the ins and outs of Couchbase, it's been a worthwhile exercise.

## Versions

* Couchbase - Community Edition - version __6.0.0__
* Docker __18.09.7__
* Docker-Compose __1.25.4__
* Go __1.15.4__

More recent versions of these components should be fine.

## Reference

Query Optimization Using Prepared (Optimized) Statements:

    http://docs.couchbase.com/go-sdk/1.5/n1ql-query.html#prepare-stmts

[My feeling is that using ___named___ parameters is more self-documenting (and may
 therefore result in fewer bugs) than using ___positional___ parameters.]

Concurrent Document Mutations:

    http://docs.couchbase.com/go-sdk/1.5/concurrent-mutations-cluster.html

Views:

    http://docs.couchbase.com/server/6.0/learn/views/views-intro.html

#### Couchbase BLOG

For general articles on Couchbase, their [BLOG](http://blog.couchbase.com/) is the place to start.

I found the following articles useful:

    http://blog.couchbase.com/moving-from-sql-server-to-couchbase-part-1-data-modeling/

[To a large extent, the use of JSON in Couchbase removes the need for OR/M tooling.]

    http://blog.couchbase.com/sql-server-couchbase-data-migration/

[The article states there is no __date__ primitive in JSON. While this may be technically
 true, for most uses the __time__ primitive will suffice instead.]

    http://blog.couchbase.com/moving-sql-server-couchbase-app-migration/

[The article focuses on migrating from SQL Server but is useful for other databases.]

    http://blog.couchbase.com/comparing-couchbase-views-couchbase-n1ql-indexing/

[Compares and contrasts Couchbase Views with Couchbase N1QL & Indexing using GSI.]

## To Do

- [x] Learn [N1QL](http://docs.couchbase.com/server/6.0/getting-started/try-a-query.html)
- [ ] Investigate the use of UUIDs as well as replication consensus procedures
- [ ] Investigate using views to enforce constraints (indexes are a performance nightmare)
- [x] Upgrade to latest release of Golang (__1.15.4__ as of the time of writing)
- [x] Upgrade `release` badge to conform to new Shields.io standards
- [ ] Investigate the use of `n1qlResp.Metrics.MutationCount`
- [x] Add Travis CI build process & code coverage reporting
- [x] Add pessimistic locking to updates
- [ ] Update build process to `vgo`
- [ ] Add tests for data TTL
