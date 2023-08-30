
## Usage

### Importer

There is a tool designed to import the data, which assumes you have a series of parquet files in a particular folder structure:

```
  │
  ├── Experiment1
  │    │
  │    ├── Taxa1
  │    │    │
  │    │    ├── res_10123_7.parquet
  │    │    ├── res_356734_7.parquet
  │    │    ├── ...
  │    │    └── res_9971_7.parquet
  │    │
  │    ├── Taxa2
  │    │    │
  ...
```
Each experiment has a subfolder for taxa (mammals, birds, etc.) and then in each is a result file from `h3calculator.py` in [persistence-calculator](https://github.com/carboncredits/persistence-calculator/), where the name contains the IUCN species ID (e.g., 10123) and the H3 tile zoom level (e.g., 7) in the name.

You pass one experiment folder to the import tool like thus:

```
$ ./bin/importer /maps/results/Experiment1
```

You will need to provide the tool with the PostGIS database credentials in the environment like this:

```
$ export PSERVER_IMPORT_DSN="host=localhost user=postgres password=mysecretpassword dbname=postgres port=5432 sslmode=disable"
```

The provided user must have the correct permissions to write to the database and modify tables (for migrations).


### Server

The JSON data server will need another environment variable set like so:

```
$ export PSERVER_DSN="host=localhost user=postgres password=mysecretpassword dbname=postgres port=5432 sslmode=disable"
```

It uses a different variable name so that you can optionally provide just a read only end point to the server.

You then run the server like this:

```
$ ./bin/pserver
```
