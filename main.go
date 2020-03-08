// handle basic HTTP requests

package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	//http.HandleFunc("/_ah/warmup", warmupHandler)
	http.HandleFunc("/", defaultHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func warmupHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if i := strings.Index(host, ":"); i != -1 {
		// Handle host or host:port
		log.Printf("%v -> host %v, port %v", host, host[:i], host[i:])
		host = host[:i]
	}
	domain := strings.Split(host, ".")
	path := r.RequestURI

	// Get subdomain part
	subStart := 0
	subEnd := len(domain) // not included
	// Find last domain part
SUB_END:
	for subEnd > subStart {
		subEnd--
		switch domain[subEnd] {
		case "victorz",
			// "victor1",
			"victor-redirect":
			break SUB_END
		}
	}
	// Strip leading part
SUB_START:
	for subStart < subEnd {
		switch domain[subStart] {
		case "www", "incoming":
			subStart++
		default:
			break SUB_START
		}
	}

	// Get site
	site := "https://victorz.ca"
	if subEnd != subStart { // subEnd-subStart != 0
		// Check subdomain (partial suffix)
		switch domain[subEnd-1] {
		case "acr":
			site = "https://acr.victorz.ca"
			subEnd--
			/*
				if subEnd != subStart {
					switch domain[subEnd-1] {
					case "forum":
						site = "https://forum.acr.victorz.ca"
						subEnd--
					}
				}
			*/
		default:
			if subEnd-subStart == 1 {
				// Check subdomain (exact match)
				switch domain[subStart] {
				case "dunk":
					site = "https://games.victorz.ca/cat/6/dunk"
				case "r":
					scheme := "http"
					if r.TLS != nil {
						scheme = "https"
					}
					site = scheme + "://" + r.Host
					num := 0
					if len(path) >= 1 {
						num, _ = strconv.Atoi(path[1:])
					}
					path = "/" + strconv.Itoa(num+1)
				default:
					goto NOT_EXACT_SUBDOMAIN
				}
				subEnd = subStart
			NOT_EXACT_SUBDOMAIN:
			}
		}

		// Append extra subdomain parts in reverse
		for i := subEnd - 1; i >= subStart; i-- {
			site += "/" + domain[i]
		}
	}

	if path != "/" {
		// Add on the path
		site += path
	}

	log.Printf("%v <- %v <- %v", site, domain[subStart:subEnd], domain)
	http.Redirect(w, r, site, 301)
}
