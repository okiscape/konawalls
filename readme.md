# Konawalls
> CLI Tool for quick wallpapers installation

`config.json` schema:
```jsonc
{
    "tags": ["konata_izumi"],
    "limit": 100, // count of arts from the provider
    "savePath": "~/Pictures/",
    "executeAfter": null, // wallpaper installation cmd
    "defaultProvider": "konachan", // or a list: ["konachan", ...]
    "providers": [
        {
            "danbooru": { // for example (danbooru is already added in defaults providers, no need to add this in your config)
                "tags": ["rating:g"], // tags whichs will added ESPECIALLY for this provider, they'll added to main tags list(for example: filters against nsfw)
                
                "baseUrl": "https://danbooru.donmai.us", // domain without / at the end
                "apiPath": "/posts.json", // path to posts API
                "apiKey": null, // "..." if your provider requires auth
                "apiUser": null,

                // settings of a provider API structure, data keys and etc
                "imageField": "file_url", // post image-url key name
                "idField": "id", // post id key name
                "tagsField": "tag_string", // tags http arg name
                
                "nesting": null, // "..." if posts info contained in additional key (api gives not list but an object with smt like {"posts": []})
                "postUrlTemplate": "/posts/{id}", // post link template
                "limit": 150
            }, 
            "yandere": {
                "tags": ["s"] // and you can customize settings even a little bit
            }
        }
    ]
}
```

build using `go build -o konawalls` and `ln -s (absolute path to builded executable) /usr/bin/konawalls`

### todos
 - [ ] add placeholders in executeAfter(for example, to save arts and for not deleting them - symlinking em)
