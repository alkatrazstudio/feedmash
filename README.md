# FeedMash

Monitors multiple RSS/Atom/JSON feeds, 
combines them into one Atom feed
and serves the resulting feed via HTTP.


## Usage

```
Usage:
  feedmash <config-file>

Examples:
  1) Get an example config (which also contains further instructions):

    feedmash --print-example-config

  2) Use that config to create your own config file and then pass it to FeedMash:

    feedmash /path/to/your/config.yaml

Flags:
  -h, --help                   help for feedmash
      --print-example-config   print an example config file
  -v, --version                version for feedmash
```

The rest of the documentation is in the example config file.
Get it with `feedmash --print-example-config` or see it
[here](https://github.com/alkatrazstudio/feedmash/blob/master/config.yaml).

## Minimum system requirements

- Ubuntu 20.04 (x86_64)
- Windows 10 version 1909 (x86_64)
- macOS 10.15 Catalina (x86_64)


## License

AGPLv3
