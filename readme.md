# Konawalls
> Mini CLI Tool for quick wallpapers installing

`config.json` schema:
```json
{
    "tags": ["example_art_tag"],
    "savePath": "somepath",
    "executeAfter": "wallpaper installation cmd(may be null)" 
}
```

build using `go build -o konawalls` and `ln -s ./konawalls /usr/bin/konawalls`
