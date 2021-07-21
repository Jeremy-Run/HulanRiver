package gateway

import (
	"HulanRiver/src/config"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

var serverPool ServerPool

func lb(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
	}

	peer := serverPool.GetNextPeer()
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func GetTokensByConfig() (tokens []string) {
	confFile := "./src/config/config.cfg"
	config.InitConfig(confFile)
	appConfig := config.AppConfigManager.Config.Load().(*config.AppConfig)
	tokens = strings.Split(appConfig.ListenServer, ",")
	return
}

func Run(port int) {
	tokens := GetTokensByConfig()
	for _, tok := range tokens {
		serverUrl, err := url.Parse(tok)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
			log.Printf("[%s] %s\n", serverUrl.Host, e.Error())
			retries := GetRetryFromContext(request)
			if retries < 3 {
				select {
				case <-time.After(10 * time.Millisecond):
					ctx := context.WithValue(request.Context(), Retry, retries+1)
					proxy.ServeHTTP(writer, request.WithContext(ctx))
				}
				return
			}

			serverPool.MarkBackendStatus(serverUrl, false)

			attempts := GetAttemptsFromContext(request)
			log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
			ctx := context.WithValue(request.Context(), Attempts, attempts+1)
			lb(writer, request.WithContext(ctx))
		}

		serverPool.AddBackend(&Backend{
			URL:          serverUrl,
			Alive:        true,
			ReverseProxy: proxy,
		})
		log.Printf("Configured server: %s\n", serverUrl)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(lb),
	}

	go HealthCheck()

	log.Printf("Load Balancer started at: %d\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// 重启的思路
// 将信号量捕捉到，关闭原进程
//func signalHandler() {
//	ch := make(chan os.Signal, 1)
//	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
//	for {
//		sig := <-ch
//		log.Printf("signal: %v", sig)
//
//		// timeout context for shutdown
//		ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
//		switch sig {
//		case syscall.SIGINT, syscall.SIGTERM:
//			// stop
//			log.Printf("stop")
//			signal.Stop(ch)
//			server.Shutdown(ctx)
//			log.Printf("graceful shutdown")
//			return
//		case syscall.SIGUSR2:
//			// reload
//			log.Printf("reload")
//			err := reload()
//			if err != nil {
//				log.Fatalf("graceful restart error: %v", err)
//			}
//			server.Shutdown(ctx)
//			log.Printf("graceful reload")
//			return
//		}
//	}
//}
