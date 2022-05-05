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
const version = "v2022-05-05"

type warmerError struct {
	response *colly.Response
	err      error
}

func (wr warmerError) Error() string {
	return wr.err.Error()
}

func main() {
	log.Printf("CACHE WARMER START")
	opts := generateOptions()
	opts.print()
	if opts.curl != "" {
		resp, err := http.Get(opts.curl)
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
		fmt.Printf("\n-------- Curl OK --------\nstatus: %s\n\n", resp.Status)

	}

	numResponses, errs := warmCache(opts)

	fmt.Printf("\n-------- responses: %d --------\n", numResponses)

	if len(errs) > 0 {
		fmt.Printf("\n-------- ERRORS: %d --------\n", len(errs))
	}
	for _, err := range errs {
		fmt.Printf("%d: %s, %s\n", err.response.StatusCode, err.response.Request.URL, err.Error())
	}
	log.Printf("CACHE WARMER END")

}

func warmCache(o options) (int, []warmerError) {

	var errs []warmerError
	numResponses := 0

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

	if !o.verbose {
		fmt.Println("'.':  response, '4': code 4xx, '5': code 5xx, 'e': other error")
	}

	c.OnError(func(r *colly.Response, e error) {
		if 400 <= r.StatusCode && r.StatusCode < 500 {
			fmt.Print("4")
		} else if 500 <= r.StatusCode && r.StatusCode < 600 {
			fmt.Print("5")
		} else {
			fmt.Print("e")
		}
		errs = append(errs, warmerError{response: r, err: e})
	})
	c.OnResponse(func(r *colly.Response) {
		numResponses++
		if o.verbose {
			fmt.Printf("- %s\n", r.Request.URL)
		} else {
			fmt.Print(".")
		}
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
	return numResponses, errs
}

type options struct {
	urlRegex   *regexp.Regexp
	initialUrl string
	workers    int
	depth      int
	verbose    bool
	curl       string
	agent      string
}

func (o options) print() {
	if o.curl == "" {
		o.curl = "[no curl]"
	}
	fmt.Printf("----- OPTIONS -----\nInitial url: %s\nRegex filter: %s\nMax depth: %d\nWorkers: %d\nCurl: %s\n\n", o.initialUrl, o.urlRegex, o.depth, o.workers, o.curl)
}

func generateOptions() options {
	// FLAGS
	flagWorkers := flag.Int("w", 1, "Workers: How many workers to use.")
	flagFilter := flag.String("f", `example\.com`, "Filter: Accepts a regular expression and only visits urls matching that expression.")
	flagInitialUrl := flag.String("i", "https://www.example.com/", "Initial Url: Defines a starting point for the warmer.")
	flagDepth := flag.Int("d", 2, "Depth: Defines how many steps to follow links found. '1' is just the initial url.")
	flagVerbose := flag.Bool("v", false, "Verbose: Enables extra printouts.")
	flagCurl := flag.String("c", "", "Curl: Make a curl to the given url before starting warmer. Warmer will not start if response status code is not 200.")
	flagAgent := flag.String("a", "cache-warmer_"+version, "UserAgent: Set custom value for the \"User-Agent\" header.")
	flag.Parse()

	return options{
		urlRegex:   regexp.MustCompile(*flagFilter),
		initialUrl: *flagInitialUrl,
		workers:    *flagWorkers,
		depth:      *flagDepth,
		verbose:    *flagVerbose,
		curl:       *flagCurl,
		agent:      *flagAgent,
	}
}
