# Konawalls
> Mini CLI Tool for quick wallpapers installing from https://konachan.com/

`config.json` schema:
```jsonc
{
    "tags": ["example_art_tag"],
    "limit": 100, // count of arts from konachan 
    "savePath": "somepath",
    "executeAfter": "wallpaper installation cmd(may be null)"
}
```

build using `go build -o konawalls` and `ln -s (absolute path to builded executable) /usr/bin/konawalls`

### todos
 - [ ] add excluding of tags
 - [ ] add placeholders in executeAfter(for example, to save arts and for not deleting them - symlinking em)
