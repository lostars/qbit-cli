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

You can use `-c /path/to/config.yaml` or `--config=/path/to/config.yaml`.

Default config location is `~/.config/qbit-cli/config.yaml` or same as executable file named `config.yaml`.

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
jackett:
  host: ""
  api-key: ""
```

## Commands

You can use `qbit -h` `qbit [command] -h` for details.

### torrent
```
list    # torrent list
add     # add torrents
files   # list torrent files
search  # search torrents through qBittorrent plugins and automatically download
```

**search**
You can use `--auto-download=true` `--torrent-regex=batman` to download torrents automatically.
`qbit torrent search -h` for more details.

### rss

```
feed    # add feed
rule    # rule -h for details
sub     # sub -h for details
```

### rename

```
jp      # auto rename your JP torrents rename jp -h for more details
```

**jp**

Supports: 
* `4k` tag
* Emby `cd1` file parts
* `-C` Chinese subtitle tag
* single file or `test/test.mp4`

### plugin

```
list    # plugin list
```

### jackett

```
feed    # add jackett feed to qBittorrent
```