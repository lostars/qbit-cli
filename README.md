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
Available Commands:
  add         Add one or more torrent
  category    Manage torrent category
  delete      Delete torrents
  files       List torrent files by torrent hash
  fp          Set torrent file priority
  list        List torrents
  rename      Rename a torrent
  search      Search torrents through qBittorrent plugins
  tag         Tag management
  update      A bulk of torrent operations, support multiple or all torrents.
```

**search**

You can use `--auto-download=true` `--torrent-regex=batman` to download torrents automatically.
`qbit torrent search -h` for more details.

### rss

```
Available Commands:
  rule        Manage RSS rules
  sub         Manage subscriptions
```

### job

`job [job] -h` for details.

```
Available Commands:
  list        Job list
  run         Run job
```

### plugin

```
Available Commands:
  enable      Manage plugin status
  install     Install plugins
  list        List all plugins
  uninstall   Uninstall plugins. API seems not working...
  update      Update all plugins
```

### jackett

```
Available Commands:
  item        Item management
```

### emby

```
Available Commands:
  item        Item management
```

### app
```
Available Commands:
  info        Show app info
  p           Show app preferences in formated json(if no filter set)
  sp          Update app preferences
```