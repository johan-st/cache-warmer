package main

import (
	"fmt"

	"github.com/gocolly/colly"
)

type result struct {
	resp   []*colly.Response
	e4xx   int
	e5xx   int
	eOther int
	errs   []warmerError
}

func (r *result) addResponse(o options, resp *colly.Response) {
	r.resp = append(r.resp, resp)
	if o.verbose {
		fmt.Printf("%s\n", resp.Request.URL)
	} else {
		r.printTerminal(o)
	}
}

func (r *result) addError(o options, err warmerError) {
	r.errs = append(r.errs, err)
	if 400 <= err.response.StatusCode && err.response.StatusCode < 500 {
		r.e4xx++
		if o.verbose {
			fmt.Print("4")
		}
	} else if 500 <= err.response.StatusCode && err.response.StatusCode < 600 {
		r.e5xx++
		if o.verbose {
			fmt.Print("5")
		}
	} else {
		r.eOther++
		if o.verbose {
			fmt.Print("e")
		}
	}
	if !o.verbose && o.isTerminal {
		r.printTerminal(o)
	}
}

func (r *result) print(o options) {
	fmt.Printf("responses: %d, 4xx: %d, 5xx: %d, other: %d", len(r.resp), r.e4xx, r.e5xx, r.eOther)
}

func (r *result) printTerminal(o options) {
	fmt.Printf("\rresponses: %d, 4xx: %d, 5xx: %d, other: %d", len(r.resp), r.e4xx, r.e5xx, r.eOther)
}

func (r *result) printErrors(o options) {
	for _, err := range r.errs {
		fmt.Printf("\n%d: %s, %s", err.response.StatusCode, err.response.Request.URL, err.Error())
	}
}
