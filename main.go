package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	port           int
	password       string
	chain          string
	maxRetries     int
	retryCount     map[string]int
	retryCountLock sync.Mutex
)

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>IP Whitelist Service</title>
</head>
<body>
    <h1>Welcome</h1>
    <p>Your IP address is: {{.IP}}</p>
    <form method="post">
        <input type="password" name="password" placeholder="Enter password">
        <input type="submit" value="Add to whitelist">
    </form>
    {{if .Message}}
    <p>{{.Message}}</p>
    {{end}}
</body>
</html>
`

type PageData struct {
	IP      string
	Message string
}

func checkIPInChain(ip string) (bool, error) {
	cmd := exec.Command("iptables", "-C", chain, "-s", ip, "-j", "ACCEPT")
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func addToWhitelist(ip string) error {
	exists, err := checkIPInChain(ip)
	if err != nil {
		return fmt.Errorf("error checking IP in chain: %v", err)
	}
	if exists {
		return fmt.Errorf("IP %s already exists in chain %s", ip, chain)
	}
	cmd := exec.Command("iptables", "-I", chain, "-s", ip, "-j", "ACCEPT")
	return cmd.Run()
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Error getting IP", http.StatusInternalServerError)
		log.Printf("Error getting IP: %v", err)
		return
	}

	data := PageData{IP: ip}

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			log.Printf("Error parsing form: %v", err)
			return
		}

		retryCountLock.Lock()
		count := retryCount[ip]
		retryCountLock.Unlock()

		if count >= maxRetries {
			data.Message = "Maximum retry limit reached. Please contact the administrator."
			log.Printf("IP %s blocked due to exceeding maximum retry limit", ip)
		} else if r.FormValue("password") == password {
			err := addToWhitelist(ip)
			if err != nil {
				if strings.Contains(err.Error(), "already exists") {
					data.Message = fmt.Sprintf("Your IP (%s) is already in the whitelist.", ip)
					log.Printf("IP %s already in whitelist", ip)
				} else {
					data.Message = "Failed to add your IP to the whitelist. Please try again."
					log.Printf("Failed to add IP %s to whitelist: %v", ip, err)
				}
			} else {
				data.Message = "Your IP has been added to the whitelist."
				log.Printf("Successfully added IP %s to whitelist", ip)
			}
			retryCountLock.Lock()
			delete(retryCount, ip)
			retryCountLock.Unlock()
		} else {
			retryCountLock.Lock()
			retryCount[ip]++
			count = retryCount[ip]
			retryCountLock.Unlock()
			data.Message = fmt.Sprintf("Incorrect password. %d attempts remaining.", maxRetries-count)
			log.Printf("Incorrect password attempt from IP %s. Attempts remaining: %d", ip, maxRetries-count)
		}
	}

	tmpl, err := template.New("page").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, "Error creating template", http.StatusInternalServerError)
		log.Printf("Error creating template: %v", err)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Error rendering template: %v", err)
	}
}

func main() {
	flag.IntVar(&port, "port", 8080, "Port to run the server on")
	flag.StringVar(&password, "password", "", "Password for adding IP to whitelist")
	flag.StringVar(&chain, "chain", "DOCKER-USER", "Iptables chain name")
	flag.IntVar(&maxRetries, "max-retries", 5, "Maximum number of password retry attempts")
	flag.Parse()

	if password == "" {
		fmt.Println("Password must be provided")
		os.Exit(1)
	}

	if os.Geteuid() != 0 {
		fmt.Println("This program must be run as root. Please use sudo.")
		os.Exit(1)
	}

	retryCount = make(map[string]int)

	http.HandleFunc("/", handleRequest)

	log.Printf("Server is running on port %d...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}