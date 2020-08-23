package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	filename       = flag.String("f", "/var/log/auth.log", "ssh log file source")
	prometheusPort = flag.String("m", ":2112", "prometheus port")
	geolocate      = flag.Bool("g", true, "geolocation service enabled/disabled")
	debug          = flag.Bool("d", false, "debug mode enabled")
)

const apiStackKey = "SSH2PROM_IPSTACK_ACCESSKEY"

func main() {
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	// Setup geolocation services
	geolocationServices := []Geolocator{ipAPI{}}

	apiStackAccessKey := os.Getenv(apiStackKey)
	if apiStackAccessKey != "" {
		geolocationServices = append(geolocationServices, apiStack{AccessKey: apiStackAccessKey})
	}

	geolocator := NewGeolocationProvider(geolocationServices...)
	locatorOpt := geolocateOption{geolocator}

	// Setup parser
	parser := NewFailedConnEventParser()

	// Setup reader
	readerOpts := []ReaderOption{}
	if *geolocate {
		readerOpts = append(readerOpts, locatorOpt)
	}

	respChan := make(chan FailedConnEvent, 100)
	errorChan := make(chan error, 100)

	reader := NewFileReader(*filename, parser, respChan, errorChan, readerOpts...)
	go reader.Start()
	defer reader.Stop()
	defer close(respChan)
	defer close(errorChan)

	// Setup prometheus reporter
	rep := prometheusReporter{}
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(*prometheusPort, nil)

	// Setup shutdown from OS signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case ev := <-respChan:
			rep.Report(ev)
			log.Debugf("Reported %v", ev)
		case err := <-errorChan:
			log.Debugf("Error %v", err)
		case _ = <-sigs:
			log.Debugf("Shutting down")
			os.Exit(0)
		}
	}
}
