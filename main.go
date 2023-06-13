package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strconv"

	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
	"golang.org/x/exp/slog"
)

var fNgrokAPIToken = flag.String("ngrok", "", "Ngrok API token")
var fPrefix = flag.String("prefix", "192.168.0.0/16", "subnet to use for local server")
var fPort = flag.String("port", "8080", "local port to serve")
var fLogLevel = flag.String("log", "", "log level: info|debug")

func main() {
	flag.Parse()
	setupLoggers()

	dir := die(os.Getwd())
	slog.Info("Serving directory", "dir", dir)

	fshandler := http.FileServer(http.Dir(dir))

	if *fNgrokAPIToken != "" {
		launchNgrokServer(fshandler, *fNgrokAPIToken)
	} else {
		launchLocalServer(fshandler)
	}
}

func setupLoggers() {
	var logger *slog.Logger

	if *fLogLevel != "" {
		var opts *slog.HandlerOptions

		switch *fLogLevel {
		case "debug":
			opts = &slog.HandlerOptions{Level: slog.LevelDebug}
		case "info":
			opts = &slog.HandlerOptions{Level: slog.LevelInfo}
		}

		logger = slog.New(slog.NewTextHandler(os.Stderr, opts))
	} else {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	slog.SetDefault(logger)
}

func launchNgrokServer(handler http.Handler, token string) {
	ctx := context.Background()
	l, err := ngrok.Listen(ctx,
		config.HTTPEndpoint(),
		ngrok.WithAuthtoken(token),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ngrok ingress url: https://%s?v=%s\n", l.Addr(), generateUniqueParam())
	http.Serve(l, handler)
}

var b = make([]byte, 4)

func generateUniqueParam() string {
	rand.Reader.Read(b)
	return hex.EncodeToString(b)
}

func launchLocalServer(handler http.Handler) {
	portNum := die(strconv.Atoi(*fPort))
	address := die(getOutboundIP())

	http.Handle("/", handler)

	for i := 0; i < 20; i++ {
		port := fmt.Sprintf(":%d", portNum+i)
		u := fmt.Sprintf("http://%s%s?v=%s\n", address, port, generateUniqueParam())
		slog.Info("Serving file server", "url", u)

		fmt.Println(u)

		err := http.ListenAndServe(port, nil)
		if err != nil {
			slog.Error("Error starting server trying again", "err", err)
		}
	}
}

func getOutboundIP() (string, error) {
	prefix := die(netip.ParsePrefix(*fPrefix))
	prefix = prefix.Masked()

	ifaces := die(net.Interfaces())
	for _, i := range ifaces {
		addrs := die(i.Addrs())
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if ip4 := v.IP.To4(); ip4 != nil {
					ip := die(netip.ParseAddr(v.IP.String()))
					slog.Debug("Found IP", "ip", ip, "prefix", prefix)
					if prefix.Contains(ip) {
						return ip.String(), nil
					}
				}
			case *net.IPAddr:
				if ip4 := v.IP.To4(); ip4 != nil {
					ip := die(netip.ParseAddr(v.IP.String()))
					slog.Debug("Found IP", "ip", ip, "prefix", prefix)
					if prefix.Contains(ip) {
						return ip.String(), nil
					}
				}
			}
		}
	}

	return "", errors.New("not found")
}

func die[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
