# Another qBittorrent CLI

**This tool pays more attention about automation.**

**If you really need a full usage CLI of qBittorrent, 
[qBittorrent-cli](https://github.com/ludviglundgren/qbittorrent-cli) may suits u better.**

**All the qBittorrent operations based on webui api v2.11.3(qBittorrent 5.0+), 
you can find official docs [here](https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)).**

**This CLI may works on other version of qBittorrent.
And you may experience lots of bugs cause it hasn't been fully tested.
You can create an issue when you encounter any bug.**

## Installation

Clone the source and run `make build`.
Or you can download the binary from release page if exists.

## Configuration

You can use `-c /path/to/config.yaml` or `--config=/path/to/config.yaml` 
or just place the `config.yaml` next to the executable file.

**Attention: take good care of your password**
```yaml
server:
  host: "https://xx.com:8080"
  username: "test"
  password: "test"
torrent:
  default-save-path: "/media"
  default-save-category: "movie"
#  different tags separated by ,
  default-save-tags: "act,love"
```

## Commands

You can use `qbit -h` for details, README file may update slowly.

### [torrent]

```
Usage:
  qbit torrent [command]

Available Commands:
  add         add torrent, you can add one or more torrents seperated by blank space
  files       List torrent files by torrent hash
  list        List torrents
```

`qbit torrent [command] -h` for more information

### [search]

**Attention: auto download is disabled by default.**

```
Usage:
  qbit search <keyword> [flags]

Examples:
qbit search <keyword> --category=movie --plugins=bt4g

Flags:
      --auto-download          Attention: if true, it will auto download all the torrents that filter by torrent-regex
      --auto-management        whether enable torrent auto management default is true, valid only when auto download enabled (default true)
      --category string        category of plugin(define by plugin) (default "all")
  -h, --help                   help for search
      --plugins string         plugins a|b|c, all and enabled also supported (default "enabled")
      --save-category string   torrent save category, valid only when auto download enabled
      --save-path string       torrent save path, valid only when auto download enabled
      --save-tags string       torrent save tags, valid only when auto download enabled
      --torrent-regex string   torrent file name filter
```

### [rename]

**jp sub command**

Automatically rename your JP torrent files. 
`qbit rename jp -h` for more details.