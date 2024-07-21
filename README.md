# FeedMash

FeedMash combines multiple input web feeds into one.
You specify input feeds (RSS/Atom/JSON),
FeedMash monitors and downloads them periodically.
Then it combines all of them into one single Atom feed
and serves this final combined feed via HTTP.
You can use any feed reader to subscribe to this resulting feed.

FeedMash also can be used to subscribe to YouTube channels (even without a YouTube account).
YouTube channel feeds are a little different:
the links to them are hidden and also require some post-processing to be properly displayed in a feed reader.
FeedMash uses YouTube channel URLs to generate proper web feed URLs,
and then it does the needed processing automatically.


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

- Ubuntu 24.04 (x86_64)
- Windows 11 (x86_64)
- macOS 14 Catalina (x86_64)


## License

AGPLv3
