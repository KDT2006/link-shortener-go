package main

import (
	"encoding/json"
	"fmt"
	"io"
	"link-shortner/data"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/mostafa-asg/ip2country"
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

type BodyLink struct {
	Address string
	Short   string

	// format: [ [ <Address>, <Condition>, <Priority> ].. ]
	Destinations [][]string
}

func createLink(l BodyLink) (*data.Link, error) {
	temp := data.NewLink(l.Address)
	temp.Short = l.Short
	tempDestinations := []data.Destination{}
	for _, destination := range l.Destinations {
		if strings.Contains(destination[1], "NewCountryEquals") {
			country := strings.Split(destination[1], ":")[1]
			parsedPriority, err := strconv.ParseInt(destination[2], 10, 64)
			if err != nil {
				fmt.Printf("Error parsing priority as int: %s\n", err)
				return nil, err
			}
			newDestination := data.NewDestination(destination[0], data.NewCountryEquals(country), int(parsedPriority))
			tempDestinations = append(tempDestinations, newDestination)
		} else if strings.Contains(destination[1], "NewBrowserIn") {
			browsers := strings.Split(destination[1], " ")[1:]
			parsedPriority, err := strconv.ParseInt(destination[2], 10, 64)
			if err != nil {
				fmt.Printf("Error parsing priority as int: %s\n", err)
				return nil, err
			}
			newDestination := data.NewDestination(destination[0], data.NewBrowserIn(browsers...), int(parsedPriority))
			tempDestinations = append(tempDestinations, newDestination)
		}
	}
	temp.Destinations = tempDestinations

	return &temp, nil
}

func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return IPAddress
}

func getPublicIP() string {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(ip)
}

func main() {
	// Create a inMemStore
	store := data.NewInmemLinkStore()

	mux := http.NewServeMux()

	// Setup the IP-Country package
	ip2country.Load("./data/IP2COUNTRY.CSV")

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

		// country := ip2country.GetCountry(ReadUserIP(r)) // In Production!
		country := ip2country.GetCountry(getPublicIP()) // Only while testing!
		println("Country: ", country)
		if country == "" {
			country = "Unknown"
		}
		req := data.NewRequest(r.RemoteAddr, country)
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

	// API Route for adding new links
	mux.HandleFunc("POST /newLink", func(w http.ResponseWriter, r *http.Request) {
		var l BodyLink
		err := json.NewDecoder(r.Body).Decode(&l)
		if err != nil {
			fmt.Printf("Error decoding the body content: %s\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		link, err := createLink(l)
		if err != nil {
			fmt.Printf("Error in the createLink function: %s\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		store.SaveLink(*link)

		fmt.Printf("Link: %+v\n", link)
	})

	// TODO(maybe): Feature to update links

	// API route for deleting links
	mux.HandleFunc("DELETE /{short}", func(w http.ResponseWriter, r *http.Request) {
		short := r.PathValue("short")
		fmt.Printf("Delete request for the short: %s\n", short)

		// Delete the link
		link, err := store.DeactivateLinkByShort(short)
		if err != nil {
			fmt.Printf("Error deleting link by short: %s\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Printf("Deleted Link: %+v\n", link)
	})

	log.Fatal(http.ListenAndServe(":3000", mux))
}
