package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/gocolly/colly"
)

// Version number is used in userAgent string
const version = "v2022-09-14"

type warmerError struct {
	response *colly.Response
	err      error
}

func (wr warmerError) Error() string {
	return wr.err.Error()
}

func main() {
	log.Printf("-START-")
	defer log.Printf("-END-")
	defer fmt.Println("")
	opts := newOptions()
	opts.print()
	if opts.curl != "" {
		mustCurl(opts)
	}

	resultState := warmCache(opts)
	if opts.verbose || !isTerminal() {
		resultState.print(opts)
	}

	resultState.printErrors(opts)

}

func warmCache(o options) result {
	fmt.Printf("-------- Warming --------\n")

	res := result{}

	c := colly.NewCollector(
		colly.URLFilters(o.urlRegex),
		colly.MaxDepth(o.depth),
		colly.UserAgent(o.agent),
		colly.Async(true),
		// colly.CacheDir("./.cache"),
		// colly.Debugger(&debug.LogDebugger{}),
	)
	// c.DisableCookies()
	// c1 := &http.Cookie{Name: "Bobby", Value: "Draper"}
	// c.SetCookies("https://www.dpj.se", []*http.Cookie{c1})
	// c.OnRequest(
	// func(r *colly.Request) {
	// r.Headers.Set("Cookie", "PrestaShop-2b300779f285f16fb7e7cb0b5c8d604d=def502008d2c9e60ebc983342589dbb3277d338ae6254a6e0a7889b6711e8491b939b9df83587dd1cba670c9c3b75ef8335403d5ea1a75ea7e5ba0a3bf05451f65d06c1c9cc13402ff5d029457648ef436293dc2583dd681492f5b9ebba9081d9c9ad45a78d32fbff14ddba1551471d524f5e34e737c31f50c367fee344066de754d41a19468384d971b9861ba9ba3ad999ea381262d333f63d6372c1c0a696a85c97bc234dfef1086959607929719bde6a9a432d2f58e2208b9ac9428f9a9b7d5cc20d7f50405208f11eff5231b2c235b819af57025d86b0c2be2a17475d2d36002bde0a785dfbcb17cbbfa9ecdcb0cb960340ce1d9b69242c170c50fe1cf3674b7b3e823f6668055a8017c76dbcaa7d8119d45055a964037fe3f8cf997347d9cded4dfcfdef41a680ea85d3c59640a59c3bd4c8b188abd12e60ca2757c861a85aea44f16d839dea8831ba9e06956f7d5512d853e8399f40e8a71cae314015658cd2d277d7a1b783f1a549e2110548dcb530c9dd76aac5717f77c928605ef53757391b34354604d2c7d63f3989390c9a99eac24acc8ce52ec01fec686a978fbdb9696948285bd657e15ec639eeed9d348fc33ff47b58058584b096bca335d30b6f5f264c09d598110ca8ad9cad3b347109661a01fac369d4f4201958a996c183febfd63329cfbc64a0e39f894a9337dbf0b69e4b3a41135f8ee5a8f3815a31ebe3df348e0420efe1733f494d17342799f81446e012dbd11a9adbdae")
	// fmt.Println("Cookie", string(r.Headers.Get("Cookie")))
	// })

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: o.workers,
	})

	c.OnError(func(r *colly.Response, e error) {
		res.addError(o, warmerError{response: r, err: e})
	})
	c.OnResponse(func(r *colly.Response) {
		res.addResponse(o, r)
	})
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})
	c.OnXML("//sitemap/loc", func(e *colly.XMLElement) {
		e.Request.Visit(e.Text)
	})
	c.OnXML("//url/loc", func(e *colly.XMLElement) {
		e.Request.Visit(e.Text)
	})
	c.Visit(o.initialUrl)

	c.Wait()
	return res
}

func mustCurl(o options) {
	resp, err := http.Get(o.curl)
	if err != nil {
		fmt.Printf(" -------- Curl ERROR --------\nCachwarmer is Halting\n%s\n", err)
		log.Printf("CACHE WARMER HALTED ON CURL")
		os.Exit(1)
	}
	if resp.StatusCode != 200 {
		fmt.Printf(" -------- Curl ERROR --------\nCachwarmer is Halting\nResponse Status was %s\n", resp.Status)
		log.Printf("CACHE WARMER HALTED ON CURL")
		os.Exit(1)
	}
	fmt.Printf("-------- Curl OK --------\nstatus: %s\n\n", resp.Status)
}

type options struct {
	urlRegex   *regexp.Regexp
	initialUrl string
	workers    int
	depth      int
	curl       string
	agent      string
	verbose    bool
	isTerminal bool
	cookie     string
}

func (o options) print() {
	if o.curl == "" {
		o.curl = "[no curl]"
	}
	fmt.Printf("----- OPTIONS -----\nInitial url: %s\nRegex filter: %s\nMax depth: %d\nWorkers: %d\nCurl: %s\nUser Agent: %s\nCookie: %s\n\n", o.initialUrl, o.urlRegex, o.depth, o.workers, o.curl, o.agent, o.cookie)
}

func newOptions() options {
	// FLAGS
	flagWorkers := flag.Int("w", 1, "Workers: How many workers to use.")
	flagFilter := flag.String("f", `example\.com`, "Filter: Accepts a regular expression and only visits urls matching that expression.")
	flagInitialUrl := flag.String("i", "https://www.example.com/", "Initial Url: Defines a starting point for the warmer.")
	flagDepth := flag.Int("d", 2, "Depth: Defines how many steps to follow links found. '1' is just the initial url.")
	flagCurl := flag.String("c", "", "Curl: Make a curl to the given url before starting warmer. Warmer will not start if response status code is not 200.")
	flagAgent := flag.String("a", "cache-warmer_"+version, "UserAgent: Set custom value for the \"User-Agent\" header.")
	flagVerbose := flag.Bool("v", false, "Verbose: Print verbose output.")
	flagCookie := flag.String("cookie", "", "Cookie: Set custom value for the \"Cookie\" header.")
	flag.Parse()

	return options{
		urlRegex:   regexp.MustCompile(*flagFilter),
		initialUrl: *flagInitialUrl,
		workers:    *flagWorkers,
		depth:      *flagDepth,
		curl:       *flagCurl,
		agent:      *flagAgent,
		verbose:    *flagVerbose,
		isTerminal: isTerminal(),
		cookie:     *flagCookie,
	}
}

func isTerminal() bool {
	o, _ := os.Stdout.Stat()
	return (o.Mode() & os.ModeCharDevice) != 0
}
