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

Clone the source and run `make build` or download the binary from release page.

## Configuration

You can use `-c /path/to/config.yaml` or `--config=/path/to/config.yaml`.

Default config location is `~/.config/qbit-cli/config.yaml` or same as executable file named `config.yaml`.

**Attention:
Take good care of your password and configure the part that related to the command you use.**
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
emby:
  host: ""
  api-key: ""
  user: ""
```

**Notice:**

Emby user must be provided to use `/emby/Users/{user}/Items/{item}` api 
which is used by `emby item info <item>` command.

Be aware of your user permissions, `emby` command use `/emby/Users/**` apis.

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
rule    # rule -h for details
sub     # sub -h for details
```

### job

`job [job] -h` for details.

```
list      # job list
run       # run job by name
```

### plugin

```
list    # plugin list
```

### jackett

```
feed    # add jackett feed to qBittorrent
```

### emby

```
item    # item management `item -h` for more information
```