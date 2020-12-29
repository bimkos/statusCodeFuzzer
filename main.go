package main 

import (
	"log"
	"bufio"
	URL "net/url"
	"net/http"
	"sync"
	"os"
	"crypto/tls"
	"errors"
	"net"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"fmt"
	"time"

	"github.com/alexflint/go-arg"
)

func check(e error) {
	if e != nil {
		log.Println(e)
	}
}

func writeToFile(str string, code int) {
	f, err := os.OpenFile(strconv.Itoa(code) + ".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	check(err)
	defer f.Close()
	_, err = f.WriteString(str + "\n")
	check(err)
}

func statusCodeChecker(client *http.Client, urls chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	var url string
	url = ""
	_ = fmt.Sprintf(url)
	for {
		url := <-urls
		if url == "-1" {
			return
		}
		log.Println(url)

		resp, err := client.Get(url)
		if err != nil {
			if strings.Contains(fmt.Sprintf("%s", err), "Redirect") {
				resp.Body.Close()
				loc, err := resp.Location()
				check(err)
				toWrite := fmt.Sprintf("%s -> %s", url, loc)
				writeToFile(toWrite, resp.StatusCode)
			} else {
				log.Println(err)
			}
		} else {
				resp.Body.Close()
				writeToFile(url, resp.StatusCode)
		}
	}
}

func readFiles(client *http.Client, hosts, suffix string, threads int) {
	urls := make(chan string, threads)
	var wg sync.WaitGroup

	for count := 1; count <= threads; count++ {
		wg.Add(1)
		go statusCodeChecker(client, urls, &wg)
	}
	var suff []string
	if suffix != "" {
		suffixFile, err := os.Open(suffix)
		check(err)
		defer suffixFile.Close()
		suffixScanner := bufio.NewReader(suffixFile)
		for {
			suffixFileLine, err := suffixScanner.ReadString('\n')
			if err != nil && err != io.EOF {
				break
			}
			suffixFileLine = strings.ReplaceAll(suffixFileLine, "\n", "")
			suffixFileLine = strings.ReplaceAll(suffixFileLine, "\r", "")
			suffixFileLine = strings.ReplaceAll(suffixFileLine, " ", "")

			suff = append(suff, suffixFileLine)

			if err != nil {
				break
			}
		}
	}

	hostsFile, err := os.Open(hosts)
	check(err)
	defer hostsFile.Close()
	hostsScanner := bufio.NewReader(hostsFile)
	for {
		hostsFileLine, err := hostsScanner.ReadString('\n')
		if err != nil && err != io.EOF {
			break
		}
		hostsFileLine = strings.ReplaceAll(hostsFileLine, "\n", "")
		hostsFileLine = strings.ReplaceAll(hostsFileLine, "\r", "")
		hostsFileLine = strings.ReplaceAll(hostsFileLine, " ", "")
		if hostsFileLine != "" && hostsFileLine != "\n" {
			if suffix != "" {
				for _, v := range suff {
					if strings.Contains(hostsFileLine, "https://") || strings.Contains(hostsFileLine, "http://") {
						urls <- hostsFileLine + v
					} else {
						urls <- "https://" + hostsFileLine + v
						urls <- "http://" + hostsFileLine + v
					}
				}
			} else {
				if strings.Contains(hostsFileLine, "https://") || strings.Contains(hostsFileLine, "http://") {
					urls <- hostsFileLine
				} else {
					urls <- "https://" + hostsFileLine
					urls <- "http://" + hostsFileLine
				}
			}
		}

		if err != nil {
			break
		}
	}

	for i := 1; i <= threads; i++ {
		urls <- "-1"
	}
	wg.Wait()
}

func main() {
	// Parse flags
	var opts struct {
		Hosts string `arg:"required, -f, --file" help:"file with hosts"`
		Suffix string `arg:"-s, --suffix" help:"file with suffix"`
		Proxy string `arg:"-p, --proxy" help:"ex socks5://127.0.0.1:9050"`
		Threads int `arg:"-t, --threads" default:"1" help:"threads count"`
	}
	arg.MustParse(&opts)

	// Setting up client
	transport := &http.Transport{
		ResponseHeaderTimeout: 5 * time.Second,
		DialContext: (&net.Dialer{   
			Timeout: 5 * time.Second,
		}).DialContext,
		MaxIdleConnsPerHost: -1,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		DisableKeepAlives: true,
	}
	proxyURL, err := URL.Parse(opts.Proxy)
	check(err)

	// Check and enable proxy
	if opts.Proxy != "" {
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	
	// Create Client
	client := &http.Client{
		Transport: transport,
		Timeout: 5 * time.Second,
	}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
        return errors.New("Redirect")
    }

	// Check proxy
	if opts.Proxy != "" {
		resp, err := client.Get("http://ifconfig.io/ip")
		check(err)
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		check(err)
		log.Println("Current IP: ",string(body))
	}
	
	readFiles(client, opts.Hosts, opts.Suffix, opts.Threads)
}
