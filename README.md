# cache-warmer
## purpose
Given a regex for filtering urls and an initial url to start from it will crawl all sites linked to from the initial url.

## Usage
The tool accepts flags to set conditions and define behaviour. 

### Flags
```
  -a string
        UserAgent: Set custom value for the "User-Agent" header. (default "cache-warmer_v2022-05-05")
  -c string
        Curl: Make a curl to the given url before starting warmer. Warmer will not start if response status code is not 200.
  -d int
        Depth: Defines how many steps to follow links found. '1' is just the initial url. (default 2)
  -f string
        Filter: Accepts a regular expression and only visits urls matching that expression. (default "example\.com")
  -i string
        Initial Url: Defines a starting point for the warmer. (default "https://www.example.com/")
  -v    Verbose: Enables extra printouts.
  -w int
        Workers: How many workers to use. (default 1)
```

```
$ go build -o ./build/cache-warmer .
```

## todos:
- list location of erranious links
- alert on error
- follow redirects


## ideas
- list erronious urls from previous run and monitor where they are on the page on subsequent run