# jdq

JSON Date Query - A CLI tool for querying time-series master data stored in JSON format.

```json
[
  {"user_id":"1001","start_date":"20240401","end_date":"20240531","plan":"premium"},
  {"user_id":"1001","start_date":"20240601","end_date":"20240630","plan":"basic"}
]
```

```bash
$ jdq --date 20240522 1001 master.json
{"user_id":"1001","plan":"premium"}
```

## Features

- **CSS-style Priority Matching**: Specific keys override defaults
- **Value-level Inheritance**: Partial field overrides with fallback to defaults  
- **Input Order Preservation**: Output respects original JSON field ordering
- **Zero Dependencies**: Uses Go standard library only

## Usage

```
Usage: jdq [options] <key> <json-file>
  -date string
    	Query date (YYYY-MM-DD or YYYYMMDD format, defaults to today)
  -version
    	Show version information
```

## Build

```bash
make        # Build binary
make eas    # Run Examples as Specifications
```

## Examples

The [examples/](examples/) directory contains detailed usage examples that also serve as executable specifications. These examples demonstrate CSS-style inheritance, date handling, and value overrides - and you can verify they work with:

```bash
./test_examples.sh                    # Verify all examples work as documented
./test_examples.sh basic_queries      # Verify specific example category
```
