# RESTful Couchbase

An experiment with using [Couchbase](http://docs.couchbase.com/home/) as
a drop-in replacement for PostgreSQL.

This builds on my [RESTful-Recipes](https://github.com/mramshaw/RESTful-Recipes) repo,
which stores data in [PostgreSQL](https://www.postgresql.org/).

All dependencies are handled via [Docker](https://www.docker.com/products/docker) and __docker-compose__.

TDD (Test-Driven Development) is implemented; the build will fail if the tests do not pass.

Likewise the build will fail if either __golint__ or __go vet__ fails.

Enforces industry-standard __gofmt__ code formatting.

All testing can be done with [curl](CURLs.txt).


## Features

- uses [Gorilla MUX](http://github.com/Gorilla/mux)
- uses [Pure Go couchbase driver](http://blog.couchbase.com/go-sdk-1.0-ga/)


## Couchbase

Couchbase is a document-oriented database based on the JSON document model.

What would be referred to as a __table__ or __row__ in a relational database
is referred to as a __document__ in Couchbase. Like other NoSQL databases,
documents may have embedded elements. In this regard, Couchbase documents are
more akin to object storage than relational database rows or records.

What would be referred to as a __database__ in a relational database seems
to be referred to as a __bucket__ in Couchbase.

As with other NoSQL databases (such as DynamoDB), schemas are flexible.

Unusually, features master-master replication (all nodes are identical).

N1QL (the SQL-like Couchbase query language) operates on JSON documents,
returning JSON documents.

Similiar to __redis__ and __Cassandra__, data may be assigned arbitrary
expiry times.

Couchbase is packaged with an Admin Console GUI.


## Getting familiar with Couchbase

Couchbase has a nice introduction:

    http://hub.docker.com/r/couchbase/server

We will start off with the __Community Edition__ (6.0.0 as of the time of writing):

```bash
$ docker run --name db -p 8091-8094:8091-8094 -p 11210:11210 --rm couchbase:community-6.0.0
Unable to find image 'couchbase:community-6.0.0' locally
community-6.0.0: Pulling from library/couchbase
7b722c1070cd: Pull complete 
5fbf74db61f1: Pull complete 
ed41cb72e5c9: Pull complete 
7ea47a67709e: Pull complete 
ca04de705515: Pull complete 
90771350bcab: Pull complete 
79af75d1044c: Pull complete 
41c3df01c635: Pull complete 
e6eb0512d813: Pull complete 
3d5ef856364c: Pull complete 
52d068d8593c: Pull complete 
ed268ff62c2b: Pull complete 
61cb7b758139: Pull complete 
Digest: sha256:5aa8172f1ef8fa78bd3d0b54caefa0c691496eb3f2eceb6fac053b900aba8fca
Status: Downloaded newer image for couchbase:community-6.0.0
Starting Couchbase Server -- Web UI available at http://<ip>:8091
and logs available in /opt/couchbase/var/lib/couchbase/logs
<...>
```

[This may take some time, depending upon download speed.]

This makes the Admin UI for our Couchbase server available at:

    http://localhost:8091

It should look as follows:

![Couchbase Server 1](images/Couchbase_Server_1.png)

We will click <kbd>Setup New Cluster</kbd>.

![Couchbase Server 2](images/Couchbase_Server_2.png)

We will add values as shown (the password is `admin123`) and click <kbd>Next: Accept Terms</kbd>.

![Couchbase Server 3](images/Couchbase_Server_3.png)

We will accept the terms and conditions as shown and click <kbd>Configure Disk, Memory, Services</kbd>.

![Couchbase Server 4](images/Couchbase_Server_4.png)

We will accept the default values as shown and click <kbd>Save & Finish</kbd>.

Which should give us this spiffy dashboard:

![Couchbase Dashboard](images/Couchbase_Dashboard.png)

We will click on the __sample bucket__ link, select the ___beer sample___ option and click <kbd>Load Sample Data</kbd>.

This will give rise to the following warning screen:

![Couchbase Bucket warning](images/Couchbase_Bucket_warning.png)

We will click on the __Security__ tab, and then click <kbd>ADD USER</kbd>:

![Couchbase Add User](images/Couchbase_Add_User.png)

And the following screen will be displayed:

![Couchbase User](images/Couchbase_User.png)

We will add values as shown (the password is `test123`) and click <kbd>Add User</kbd>.

And now we can query our database
<kbd>SELECT name FROM `beer-sample` WHERE brewery_id ="mishawaka_brewing";</kbd>:

![Couchbase Query](images/Couchbase_Query.png)

Note that the bucket is surrounded by backticks (`) and the result set is provided as [JSON](http://en.wikipedia.org/wiki/JSON).

However, we can also display our result set as a __Table__ or a __Tree__. We can also ___export___ our results as JSON.

[Unusually, __Ctrl-C__ / __Ctrl-D__ will not stop our Couchbase server. We will need to kill it from a new terminal.]

## To Run

The command to run:

    $ docker-compose up -d

For the first run, there will be a warning as `mramshaw4docs/golang-couchbase:1.11` must be built.

This image will contain all of the Go dependencies and should only need to be built once.

For the very first run, `golang` may fail as it takes `couchbase` some time to ramp up.

A successful `golang` startup should show the following as the last line of `docker-compose logs golang`:

    golang_1    | 2018/02/24 18:38:01 Now serving recipes ...

In this line does not appear, repeat the `docker-compose up -d` command (there is no penalty for this).


## To Build:

The command to run:

    $ docker-compose up -d

Once `make` indicates that `restful_couchbase` has been built, can change `docker-compose.yml` as follows:

1) Comment `command: make`

2) Uncomment `command: ./restful_couchbase`


## For testing:

[Optional] Start couchbase:

    $ docker-compose up -d couchbase

Start golang [if couchbase is not running, this step will start it]:

    $ docker-compose run --publish 80:8100 golang make

Successful startup will be indicated as follows:

    2018/02/24 16:27:10 Now serving recipes ...

This should make the web service available at:

    http://localhost/v1/recipes

Once the service is running, it is possible to `curl` it. Check `CURLs.txt` for examples.


## See what's running:

The command to run:

    $ docker ps


## View the build and/or execution logs

The command to run:

    $ docker-compose logs golang


## To Shutdown:

The command to run:

    $ docker-compose down


## Clean Up

The command to run:

    $ docker-compose run golang make clean


## Results


## To Do

- [x] Learn [N1QL](http://docs.couchbase.com/server/6.0/getting-started/try-a-query.html)
- [ ] Update build process to `vgo`
- [ ] Add tests for data TTL
