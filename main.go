package main

import (
	"encoding/json"
	"fmt"
	"io"
	"link-shortner/data"
	"log"
	"net"
	"net/http"
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
	Address      string            `json:"address"`
	Short        string            `json:"short"`
	Destinations []BodyDestination `json:"destinations"`
}

type BodyDestination struct {
	Address string `json:"address"`
	// Used for parsing of raw bytes into json without specific type
	Condition json.RawMessage `json:"condition"`
	Priority  int             `json:"priority"`
}

func parseCondition(condRaw json.RawMessage) (data.Conditioner, error) {
	var condMap map[string]interface{}
	if err := json.Unmarshal(condRaw, &condMap); err != nil {
		return nil, fmt.Errorf("invalid condition format: %v", err)
	}

	condType, ok := condMap["type"].(string)
	if !ok {
		return nil, fmt.Errorf("condition type not specified")
	}

	switch condType {
	case "AND":
		return parseAndCondition(condMap)
	case "CountryEquals":
		return parseCountryEqualsCondition(condMap)
	case "BrowserIn":
		return parseBrowserInCondition(condMap)
	default:
		return nil, fmt.Errorf("unknown condition type: %s", condType)
	}
}

func parseAndCondition(condMap map[string]interface{}) (data.Conditioner, error) {
	childrenRaw, ok := condMap["children"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid AND condition format")
	}

	var children []data.Conditioner
	for _, childRaw := range childrenRaw {
		childJSON, err := json.Marshal(childRaw)
		if err != nil {
			return nil, fmt.Errorf("error marshaling child condition: %v", err)
		}

		child, err := parseCondition(childJSON)
		if err != nil {
			return nil, fmt.Errorf("error parsing child condition: %v", err)
		}

		children = append(children, child)
	}

	return data.NewAnd(children...), nil
}

func parseCountryEqualsCondition(condMap map[string]interface{}) (data.Conditioner, error) {
	country, ok := condMap["country"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid CountryEquals condition format")
	}

	return data.NewCountryEquals(country), nil
}

func parseBrowserInCondition(condMap map[string]interface{}) (data.Conditioner, error) {
	browsersRaw, ok := condMap["browsers"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid BrowserIn condition format")
	}

	browsers := make([]string, len(browsersRaw))
	for i, b := range browsersRaw {
		browsers[i], ok = b.(string)
		if !ok {
			return nil, fmt.Errorf("invalid browser type in BrowserIn condition")
		}
	}

	return data.NewBrowserIn(browsers...), nil
}

func createLink(l BodyLink) (*data.Link, error) {
	link := data.NewLink(l.Address)
	link.Short = l.Short

	for _, dest := range l.Destinations {
		condition, err := parseCondition(dest.Condition)
		if err != nil {
			return nil, fmt.Errorf("error parsing condition for destination %s: %v", dest.Address, err)
		}

		link.Destinations = append(link.Destinations, data.NewDestination(dest.Address, condition, dest.Priority))
	}

	return &link, nil
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

		err = store.SaveLink(*link)
		if err != nil {
			http.Error(w, "Failed to save link", http.StatusInternalServerError)
			return
		}

		fmt.Printf("Link: %+v\n", link)

		response := struct {
			ShortURL string `json:"short_url"`
			Message  string `json:"message"`
		}{
			ShortURL: link.Short,
			Message:  "Link created",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	})

	// TODO(maybe): Feature to update links

	// API route for deleting links
	mux.HandleFunc("DELETE /{short}", func(w http.ResponseWriter, r *http.Request) {
		short := r.PathValue("short")
		fmt.Printf("Delete request for the short: %s\n", short)

		// Delete the link
		link, err := store.ToggleLinkByShort(short)
		if err != nil {
			fmt.Printf("Error deleting link by short: %s\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Printf("Deleted Link: %+v\n", link)
	})

	log.Fatal(http.ListenAndServe(":3000", mux))
}
