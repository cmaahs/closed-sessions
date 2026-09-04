# closed-sessions

A small Go CLI for listing `restore-session-*` files by date.

## Config

The config file is stored in `~/.config/closed-sessions/config.yaml` and contains the base directory to search:

```yaml
base_search_path: /Users/you/Work
```

You can set it from the CLI:

```bash
closed-sessions config set /Users/you/Work
```

## Usage

```bash
# Today's matches
closed-sessions

# Exact date
closed-sessions 20250404

# Relative windows
closed-sessions 1d
closed-sessions 4d
closed-sessions 1w
closed-sessions 1m
```

The output strips the configured base path and prints a simple two-column table:

```text
DIRECTORY  FILENAME
nested     restore-session-20250410.txt
```
