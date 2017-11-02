package beater

import (
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/gothfrid/jolokiabeat/config"
)

// Jolokiabeat Struct
type Jolokiabeat struct {
	done   chan struct{}
	config config.Config
	client publisher.Client
	events chan []common.MapStr
}

// New - creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}
	for i, source := range config.Sources {
		if source.Headers == nil {
			config.Sources[i].Headers = make(map[string]string, 1)
		}
		config.Sources[i].Headers["Content-Type"] = "application/json"
		config.Sources[i].EndPoint = "http://" + source.Host + source.Address
	}
	jb := &Jolokiabeat{
		done:   make(chan struct{}),
		config: config,
		events: make(chan []common.MapStr),
	}
	return jb, nil
}

// Run - endless`` loop, submiting data with defined interval
func (jb *Jolokiabeat) Run(b *beat.Beat) error {

	logp.Info("Jolokiabeat is running! Hit CTRL-C to stop it.")
	jb.client = b.Publisher.Connect()
	fmt.Println(jb.config)
	go jb.handleEvents()
	ticker := time.NewTicker(jb.config.Period)
	for {
		select {
		case <-jb.done:
			return nil
		case <-ticker.C:
		}
		for _, source := range jb.config.Sources {
			go jb.FetchData(&source)
		}
	}

}

func (jb *Jolokiabeat) handleEvents() error {
	for {
		select {
		case <-jb.done:
			return nil
		case events := <-jb.events:
			jb.client.PublishEvents(events)
		}
	}
}

func (jb *Jolokiabeat) Stop() {
	jb.client.Close()
	close(jb.events)
	close(jb.done)
}
