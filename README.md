# Dela dead simple web TODO list

## About

Dela is a web TODO list application which can be hosted on your server and accessed via browser. Current capabilities include: 

- Password protected account system
- TODO creation
- TODO completion

## Build

```
make
```

or manually compile with `go build` inside the `src` directory, move the binary and put it alongside `pages`, `scripts` and `static` (although the path to the contents can be specified via configuration file).

```
make portable
```

will create a `dela.zip` archive with all the content directories and the binary which are to be freely transported to other machines (of the same OS and architecture, obviously) 

## Use

After the first run a configuration file will be put alongside the executable, upon the second launch server will start and the service will be accessible on the specified port.