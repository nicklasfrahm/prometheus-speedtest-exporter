package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/showwin/speedtest-go/speedtest"
)

var version = "dev"
var logger = log.New(os.Stdout, "inf: ", log.LUTC)

type Metrics struct {
	downloadSpeed prometheus.Gauge
	uploadSpeed   prometheus.Gauge
	ping          prometheus.Gauge
	jitter        prometheus.Gauge
	resultValid   prometheus.Gauge
	testDuration  prometheus.Gauge
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		downloadSpeed: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "download_speed_bps",
			Help: "Download speed (bit/s)",
		}),
		uploadSpeed: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "upload_speed_bps",
				Help: "Upload speed (bit/s)",
			},
		),
		ping: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "ping_seconds",
				Help: "Latency (seconds)",
			},
		),
		jitter: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "jitter_seconds",
				Help: "Jitter (seconds)",
			},
		),
		resultValid: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "result_valid",
				Help: "Indicates if the result is logical given UL and DL speed",
			},
		),
		testDuration: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "test_duration_seconds",
				Help: "Duration of the test (seconds)",
			},
		),
	}

	reg.MustRegister(m.downloadSpeed)
	reg.MustRegister(m.uploadSpeed)
	reg.MustRegister(m.ping)
	reg.MustRegister(m.jitter)
	reg.MustRegister(m.resultValid)
	reg.MustRegister(m.testDuration)

	return m
}

func main() {
	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port = ":9516"
	}

	// Create a non-global registry.
	reg := prometheus.NewRegistry()

	// Create new metrics and register them using the custom registry.
	metrics := NewMetrics(reg)

	prometheusHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		speedtestClient := speedtest.New()

		userInfo, err := speedtestClient.FetchUserInfo()
		if err != nil {
			handleErr(w, fmt.Errorf("failed to fetch user info: %w", err))
			return
		}

		serverList, err := speedtestClient.FetchServers(userInfo)
		if err != nil {
			handleErr(w, fmt.Errorf("failed to fetch server list: %w", err))
			return
		}

		targets, err := serverList.Available().FindServer([]int{})
		if len(targets) == 0 {
			err = fmt.Errorf("no available server found")
		}
		if err != nil {
			handleErr(w, fmt.Errorf("failed to find available server: %w", err))
			return
		}

		var minDistance float64 = math.MaxFloat64
		var target *speedtest.Server

		for _, t := range targets {
			if t.Distance < minDistance {
				minDistance = t.Distance
				target = t
			}
		}

		start := time.Now()

		// TODO: Implement multi-server test.
		err = target.PingTest()
		if err != nil {
			handleErr(w, fmt.Errorf("failed to run ping test: %w", err))
			return
		}

		err = target.DownloadTest()
		if err != nil {
			handleErr(w, fmt.Errorf("failed to run download test: %w", err))
			return
		}

		err = target.UploadTest()
		if err != nil {
			handleErr(w, fmt.Errorf("failed to run upload test: %w", err))
			return
		}

		elapsed := time.Since(start)

		// TODO: Add metrics with server labels for multi-server test.
		metrics.ping.Set(float64(target.Latency))
		metrics.jitter.Set(float64(target.Jitter))
		metrics.downloadSpeed.Set(target.DLSpeed * 1e6)
		metrics.uploadSpeed.Set(target.ULSpeed * 1e6)
		if target.CheckResultValid() {
			metrics.resultValid.Set(1)
		} else {
			metrics.resultValid.Set(0)
		}
		metrics.testDuration.Set(elapsed.Seconds())

		prometheusHandler.ServeHTTP(w, r)
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(http.StatusText(http.StatusOK)))
		w.WriteHeader(http.StatusOK)
	})

	logger.Printf("prometheus-speedtest-exporter: %s\n", version)
	logger.Printf("starting server: http://0.0.0.0%s\n", port)
	logger.Fatal(http.ListenAndServe(port, nil))
}

func handleErr(w http.ResponseWriter, err error) {
	msg := err.Error()
	w.Write([]byte(msg))
	logger.SetPrefix("err: ")
	logger.Print(msg)
	w.WriteHeader(http.StatusInternalServerError)
}
