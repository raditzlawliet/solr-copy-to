# Colr Copy To
Solr copy collection to Solr/MongoDB using Solr API. This project already using dep as package manager.

## Available Source 
- Solr 6.6

## Available Target 
- Solr 6.6
- MongoDB 3.6

## Build
Dont forget to `dep ensure` to getting all vendor 
```bash
go build
```

## Command
- Run : Execute Copy Solr collection to Solr/mongodb

```bash
PS D:\Go Project\src\raditzlawliet\solr-copy-to> go run main.go run --help
Copy Solr collection to Solr/mongodb

Usage:
  solr-to-mgo run [flags]

Flags:
  -h, --help                   help for run
  -m, --max int                Maximum data to be copied, -1/0 for all (default -1)
  -s, --source string          Source Collection
      --source-cursor string   Solr Source Cursor (default "*")
      --source-host string     Solr Source Full URL (with /solr) (default "http://127.0.0.1:8983/solr/")
  -q, --source-query string    Solr Source Query (default "*:*&sort=id+desc")
      --source-rows int        Solr Source Rows Fetch each Query (default 10000)
  -t, --target string          Target Collection
      --target-commit          Commit after Post (Solr) (default true)
      --target-db string       Database (Mongo)
      --target-host string     Mongo (Mongo) | Solr Source Full URL (with /solr) (default "127.0.0.1")
      --target-pass string     Password Database (Mongo)
      --target-type string     Target Collection Type (default "mongo")
      --target-user string     Username Database (Mongo)
```

## Run without Build
Dont forget to `dep ensure` to getting all vendor.

Copy from Solr searchLog collection into mongoDB searchLog collection with max data 1m and fetch/insert each 10k data.
```bash
go run main.go run -s searchLog -t searchLog --source-host http://192.168.0.230:8983/solr/ --target-host localhost:27017 --target-type mongo --target-db melon --source-rows 10000 -m 1000000
```

Copy from Solr searchLog collection into Solr searchLog collection with max data 1m and fetch/insert each 10k data, auto commit true.
```bash
go run main.go run -s searchLog -t searchLog --source-host http://192.168.0.230:8983/solr/ --target-host http://192.168.0.230:8983/solr/ --target-type solr --source-rows 10000 -m 1000000
```
