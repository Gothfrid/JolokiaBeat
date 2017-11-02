// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"time"
)

// Source data structure
type Source struct {
	Host      string            `config:"host"`      // Host of data source
	Address   string            `config:"address"`   // Address of data source
	FetchOnly bool              `config:"fetchOnly"` //Fetch all or only listed domains
	Domains   []string          `config:"domains"`   // Included or Excluded Java Domains
	Bean      []string          `config:"beans"`     // Included or Excluded Java Beans
	Headers   map[string]string `config:"headers"`   // Auth or other required headers
	EndPoint  string
}

// Config data structure
type Config struct {
	Period  time.Duration `config:"period"`  // Fetch interval
	Sources []Source      `config:"sources"` // Sources to Fetch
}

// DefaultConfig set default values
var DefaultConfig = Config{
	Period: 5 * time.Second,
	Sources: []Source{
		Source{
			Host:    "localhost:8080",
			Address: "/jolokia",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			FetchOnly: false,
			Domains: []string{
				"JMImplementation::0",
				"jolokia::0",
				"jmx4perl::0",
				"Catalina::3",
				"com.sun.management::0",
				"java.util.logging::0",
			},
		},
	},
}
