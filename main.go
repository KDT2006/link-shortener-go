package main

import (
	"fmt"
	"link-shortner/data"
	"log"
	"net/http"
	"strings"
)

func detectBrowser(userAgent string) string {
	switch {
	case strings.Contains(userAgent, "Edg"):
		return "edge"
	case strings.Contains(userAgent, "Firefox"):
		return "firefox"
	case strings.Contains(userAgent, "Chrome"):
		return "chrome"
	case strings.Contains(userAgent, "Safari"):
		return "safari"
	default:
		return "unknown"
	}
}

func main() {
	// Create a inMemStore
	store := data.NewInmemLinkStore()

	mux := http.NewServeMux()

	mux.HandleFunc("/{short}", func(w http.ResponseWriter, r *http.Request) {
		short := r.PathValue("short")
		fmt.Printf("You have hit the short %s\n", short)

		// Get the link by short
		link, err := store.GetLinkByShort(short)
		if err != nil {
			fmt.Printf("Error getting link by short: %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO: Country is being set manually, gotta change that
		req := data.NewRequest(r.RemoteAddr, "US")
		req.Browser = detectBrowser(r.UserAgent())

		// Get the proper destination from the shortener
		dest, err := link.PickDestination(req)
		if err != nil {
			fmt.Printf("Error getting destination address: %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, dest.URL, http.StatusTemporaryRedirect)
	})

	// TODO: Feature to add new links

	// TODO: Feature to update links

	// TODO: Feature to delete links

	log.Fatal(http.ListenAndServe(":3000", mux))
}
