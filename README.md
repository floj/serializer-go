# serializer-go

This is a cheap clone of [serializer.io](https://serializer.io/) that does just do the Hackernews part of it and has less features.

Here is the original readme:

> serializer collects links from Hacker news & others and lists them in
> **sequential** order, some might even say it *serializes* them. It also
> collects a few other sources, not all of which are voting based, e.g. Ars
> Technica.

## Why?
At some point in time, HN blocked serializer.io from scraping for a couple of days, but I was so used to get the frontpage stories in a _serialized_ way, that I decided to make a quick and dirty clone and run it in my homelab.

## Build & run
To build the executable
```sh
# install required code-gen tools
go install github.com/a-h/templ/cmd/templ@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# generate models and templates
sqlc generate
templ generate

# build the actual app
go build -o serializer-go .
```

A Postgres DB is required, you can either run it locally (eg. via [docker](https://hub.docker.com/_/postgres)) or get a SaaS one (eg. for free from [neon.tech](https://neon.tech)).

```sh
# start DB locally (in an extra terminal)
docker run -ti --rm -p 5432:5432 -e POSTGRES_DB=serializer -e POSTGRES_PASSWORD=secret postgres:15

# run the app
./serializer-go -db-uri "postgresql://postgres:secret@localhost/serializer?sslmode=disable"
```

## Credits
All credit goes to [charlieegan3](https://github.com/charlieegan3) for building such an awesome service and providing it for free.