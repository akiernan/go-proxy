package agent

import (
	"log"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-proxy/api"
)

// Agent interface.
type WavefrontAgent interface {
	InitAgent()
}

type DefaultAgent struct {
	AgentID    string
	ApiService api.WavefrontAPI
	LocalAgent bool
	PushAgent  bool
	Ephemeral  bool
	ServerURL  string
}

func (a *DefaultAgent) InitAgent() {
	// register agent GC and memory usage statistics
	// buildAgentMetrics() updates these stats every minute
	metrics.RegisterRuntimeMemStats(metrics.DefaultRegistry)

	// fetch configuration once per minute
	checkinTicker := time.NewTicker(time.Minute * time.Duration(1))
	go a.checkin(checkinTicker)
}

func (a *DefaultAgent) checkin(ticker *time.Ticker) {
	for range ticker.C {
		a.doCheckin()
	}
}

func (a *DefaultAgent) doCheckin() {
	log.Println("Fetching configuration from", a.ServerURL)

	agentMetrics, err := buildAgentMetrics()
	if err != nil {
		log.Println("buildAgentMetrics error", err)
		return
	}

	currentTime := getCurrentTime()
	agentConfig, err := a.ApiService.Checkin(currentTime, a.LocalAgent, a.PushAgent, a.Ephemeral, agentMetrics)
	if err != nil {
		log.Println("Checkin error", err)
		return
	}

	//TODO: update forwarder based on fetched configuration
	log.Println("AgentConfig", *agentConfig)
	err = a.ApiService.AgentConfigProcessed()
	if err != nil {
		log.Println("AgentConfigProcessed error", err)
	}
}

func getCurrentTime() int64 {
	return time.Now().UnixNano() / 1000000
}
