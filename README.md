![GitHub Repo stars](https://img.shields.io/github/stars/skanehira/db-data-generator?style=social)
![GitHub](https://img.shields.io/github/license/skanehira/db-data-generator)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/skanehira/db-data-generator)
![GitHub all releases](https://img.shields.io/github/downloads/skanehira/db-data-generator/total)
![GitHub CI Status](https://img.shields.io/github/workflow/status/skanehira/db-data-generator/ci?label=CI)
![GitHub Release Status](https://img.shields.io/github/workflow/status/skanehira/db-data-generator/Release?label=release)

# db-data-generator
Generate test data to database.

## How to use
```sh
$ db-data-generator -h
Usage:
  db-data-generator [flags]
  db-data-generator [command]

Available Commands:
  version
  completion  generate the autocompletion script for the specified shell
  help        Help about any command

Flags:
  -D, --database string   database name
  -f, --file string       definition file
  -h, --help              help for db-data-generator
      --host string       database host
  -l, --limit string      generate table rows limit
  -p, --password string   password
  -P, --port string       database port
  -u, --user string       user name

Use "db-data-generator [command] --help" for more information about a command.

$ db-data-generator -f definition.yaml -l 1000000 -u root -p test -D test
```

```yaml
# kind of database
# current available value is only mysql
database: mysql
tables:
  - name: users
    columns:
      - name: id
        value: xid
      - name: name
        value: # if value is array, value will picked random
          - gorilla
          - godzilla
          - dog
          - cat
          - human
      - name: age
        value:
          - 10
          - 15
          - 33
```

## Author
skanehira
