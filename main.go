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
}

func (o options) print() {
	if o.curl == "" {
		o.curl = "[no curl]"
	}
	fmt.Printf("----- OPTIONS -----\nInitial url: %s\nRegex filter: %s\nMax depth: %d\nWorkers: %d\nCurl: %s\nUser Agent: %s\n\n", o.initialUrl, o.urlRegex, o.depth, o.workers, o.curl, o.agent)
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
	}
}

func isTerminal() bool {
	o, _ := os.Stdout.Stat()
	return (o.Mode() & os.ModeCharDevice) != 0
}
