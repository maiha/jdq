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
- **Configurable Field Names**: Specify custom key and date field names
- **Input Order Preservation**: Output respects original JSON field ordering
- **Zero Dependencies**: Uses Go standard library only

## Usage

```
Usage: jdq [options] <key> <json-file>

Options:
  -d, --date string          Query date (YYYY-MM-DD or YYYYMMDD format, defaults to today)
  -k, --key-field string     Primary key field name (defaults to first field)
  -s, --start-field string   Start date field name (default "start_date")
  -e, --end-field string     End date field name (default "end_date")
  -E, --exit-status          Exit with non-zero status when no record found
  -v, --version              Show version information
```

### Examples

```bash
# Basic query with default field names
jdq --date 20240522 1001 data.json

# Custom field names (short form)
jdq -k account_id -s from_date -e to_date -d 20240522 1001 data.json

# Exit with non-zero status if not found (useful in scripts)
jdq -E -d 20240522 9999 data.json || echo "Record not found"
```

For detailed use cases, see [examples/](examples/) directory.

You can verify all examples work as documented:

```bash
./test_examples.sh                    # Verify all examples
./test_examples.sh basic_queries      # Verify examples/basic_queries/spec.txt
```

## Build

```bash
make        # Build binary
make eas    # Run Examples as Specifications
```

## License

MIT License - see [LICENSE](LICENSE) file for details
