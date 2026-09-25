package main

import (
	"errors"
	"flag"
	"strings"
	"time"
)

type Config struct {
	ListenAddress    string
	BackendAddresses []string
	HealthInterval   time.Duration
	DialTimeout      time.Duration
}

func parseConfig() (Config, error) {
	listen := flag.String("listen", ":8080", "address to listen on")
	backends := flag.String("backends", "", "comma seperated list of backend addresses")
	healthInterval := flag.Duration("health-interval", 5*time.Second, "interval between backend health checks")
	dialTimeout := flag.Duration("dial-timeout", 2*time.Second, "backend connection timeout")

	flag.Parse()

	if strings.TrimSpace(*backends) == "" {
		return Config{}, errors.New("at least one backend must be provided")
	}

	backendAddresses := strings.Split(*backends, ",")

	for i := range backendAddresses {
		backendAddresses[i] = strings.TrimSpace(backendAddresses[i])
	}

	return Config{
		ListenAddress:    *listen,
		BackendAddresses: backendAddresses,
		HealthInterval:   *healthInterval,
		DialTimeout:      *dialTimeout,
	}, nil
}
