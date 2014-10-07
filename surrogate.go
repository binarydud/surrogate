package main

import (
    "fmt"
    "log"
    "flag"
    "time"
    "net/http"
    "github.com/BurntSushi/toml"
    "github.com/mailgun/vulcan"
    "github.com/mailgun/vulcan/endpoint"
    "github.com/mailgun/vulcan/route/pathroute"
    "github.com/mailgun/vulcan/location/httploc"
    "github.com/mailgun/vulcan/loadbalance/roundrobin"
    st "github.com/binarydud/surrogate/types"
)
func Log(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        clientIP := r.RemoteAddr
	    if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
		    clientIP = clientIP[:colon]
	    }
        requestLine := fmt.Sprintf("%s %s %s", r.Method, r.RequestURI, r.Proto)
        log.Printf("%s - %s", r.RemoteAddr, r.Method, r.URL)
        handler.ServeHTTP(w, r)
    })
}

func main() {
    var config_file = flag.String("config", "sample.toml", "dub config file path")
    flag.Parse()
    var config st.Config
    if _, err := toml.DecodeFile(*config_file, &config); err != nil {
        fmt.Println(err)
        log.Fatalf("Error: %s", err)
    }
    log.Println(config.Frontends)
    exit_chan := make(chan int)
    count := 0
    for serverName, server := range config.Frontends {
        go func (name string, serverConfig st.Frontend, backendMap map[string]st.Backend) {
            defer func() {
                if err := recover(); err != nil {
                    log.Println("work failed:", err)
                }
            }()
            log.Println("Starting CreateServer")
            // find backends attach to server
            // for each backend
            log.Println("Creating Path Router")
            router := pathroute.NewPathRouter()
            for _, name := range(serverConfig.Backends) {
                backendConfig, ok := backendMap[name]
                if !ok {
                    log.Fatalf("Error: %s", ok)
                }
                // create load balancer (aka, Round Robin)
                log.Println("Creating Round Robin Load Balancer")
                rr, err := roundrobin.NewRoundRobin()
                if err != nil {
                    log.Fatalf("Error: %s", err)
                }
                // add hosts from backend
                for _, host := range(backendConfig.Hosts){
                    log.Println("Adding following host to balancer", endpoint.MustParseUrl(host))
                    rr.AddEndpoint(endpoint.MustParseUrl(host))
                }
                log.Println("Creating Location")
                loc, err := httploc.NewLocation(name, rr)
                // create location
                if backendConfig.Path == "" {
                    backendConfig.Path = "/"
                }
                router.AddLocation(backendConfig.Path, loc)
            }
            // create proxy
            proxy, err := vulcan.NewProxy(router)
            if err != nil {
                log.Fatalf("Error: %s", err)
            }
            server := &http.Server{
                Addr:           serverConfig.Bind,
                Handler:        Log(proxy),
                ReadTimeout:    10 * time.Second,
                WriteTimeout:   10 * time.Second,
                MaxHeaderBytes: 1 << 20,
            }
            log.Println("Listening on", serverConfig.Bind)
            if err := server.ListenAndServe(); err != nil {
                panic(err)
                log.Fatalf("Error: %s", err)
            }
            exit_chan <- 1

        }(serverName, server, config.Backends)
        count++
        log.Println(count)
    }
    for i := 0; i < count; i++ {
        <-exit_chan
    }
}
