![Dela](static/images/android-chrome-192x192.png "Dela logo")

# Dela - dead simple web TODO list

## About
Dela is a web TODO list application which can be hosted on your server and accessed via browser. Current capabilities include: 

- Password protected account system
- TODO creation
- TODO completion
- TODO deletion
- TODO drag-and-drop on categories
- Due date selection on TODO creation

## Build
Dela is written in Go, so you need to have a Go compiler.

For ease of compilation it is also recommended to have `make`.

To automatically compile Dela and have a ready-to-go solution, run:

```
make
```

a portable `/bin` directory should appear where binary and base contents are located.


### Manual build
Or you can manually compile with `go build` inside the `src` directory. Move the binary and put it alongside `pages`, `scripts` and `static` (or specify the path to the contents directory via configuration file which is created automatically upon first launch).


### Portable build
To get a portable archive run:

```
make portable
```

which will create `dela.zip` archive with all the content directories and the binary which can be freely transported to other machines (of the same OS and architecture) 

### Cross compilation
For cross-platform compilation there is:

```
make cross
```

which cross compiles the project for linux, freebsd, windows and darwin systems. Portable solutions will be put in `/bin` directory in the corresponding subdirectories.

## Use

### Configuration file
After the first run a configuration file will be put alongside the executable, upon the second launch the server will start and the service will be accessible on the specified port.

Currently configuration file contains these filelds:
```json
{
 "port": 8080,
 "cert_file_path": "",
 "key_file_path": "",
 "base_content_dir": ".",
 "production_db_name": "dela.db"
}
```

| Field | Description |
| --- | ----------- |
| port | port on which the service will run |
| cert_file_path | path to the SSL certificate file |
| key_file_path | path to the SSL certificate key file |
| base_content_dir | path to the directory with `pages`, `scripts` and `static` subdirectories |
| production_db_name | SQLite3 database file path |


### SSL certificates
If you intend to use SSL certificates - there are corresponding fields in the configuration file.

## License
Dela is licensed under AGPL