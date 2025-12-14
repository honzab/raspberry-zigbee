package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/lib/pq"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// {"battery":100,"contact":false,"linkquality":204,"update":{"installed_version":16777241,"latest_version":16777241,"state":"idle"}}
type MainDoorMessage struct {
	Battery     int32 `json:"battery"`
	Contact     bool  `json:"contact"`
	LinkQuality int32 `json:"linkquality"`
	Update      struct {
		InstalledVersion int32  `json:"installed_version"`
		LatestVersion    int32  `json:"latest_version"`
		State            string `json:"state"`
	} `json:"update"`
}

type Metric struct {
	Metric string            `json:"metric"`
	Value  float64           `json:"value"`
	Tags   map[string]string `json:"tags"`
}

func connect(clientId string, uri *url.URL) mqtt.Client {
	opts := createClientOptions(clientId, uri)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
	return client
}

func createClientOptions(clientId string, uri *url.URL) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", uri.Host))
	opts.SetUsername(uri.User.Username())
	password, _ := uri.User.Password()
	opts.SetPassword(password)
	opts.SetClientID(clientId)
	return opts
}

func simpleListen(client mqtt.Client, uri *url.URL, topic string, name string) {
	client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		//log.Printf("P1IB: [%s] %s", msg.Topic(), msg.Payload())
		value, err := strconv.ParseFloat(string(msg.Payload()), 64)
		if err != nil {
			log.Printf("Could not parse %s", msg.Payload())
			return
		}
		m := Metric{
			Metric: name,
			Value:  value,
			Tags:   map[string]string{},
		}
		marshalled, err := json.Marshal(m)
		if err != nil {
			log.Printf("Could not marshal the tsdb put")
			return
		}
		//fmt.Printf("Posting: %v+\n", m)
		// # TODO(janbrucek)(20250117) Configurable!
		http.Post("http://192.168.1.102:4242/api/put", "application/json", bytes.NewReader(marshalled))
	})
}

func listen(client mqtt.Client, uri *url.URL, topic string) {
	client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		var parsedPayload MainDoorMessage

		err := json.Unmarshal(msg.Payload(), &parsedPayload)
		if err != nil {
			log.Printf("Could not unmarshal %s", msg.Payload())
			return
		}
		value := float64(0)
		if parsedPayload.Contact {
			value = float64(1)
		}

		m := Metric{
			Metric: "door_closed",
			Value:  value,
			Tags:   map[string]string{"sensor": msg.Topic()},
		}
		marshalled, err := json.Marshal(m)
		if err != nil {
			log.Printf("Could not marshal the tsdb put")
			return
		}
		//fmt.Printf("Posting: %v+\n", m)
		// # TODO(janbrucek)(20250117) Configurable!
		http.Post("http://192.168.1.102:4242/api/put", "application/json", bytes.NewReader(marshalled))

		storeInPostgres(parsedPayload.Contact)

		if !parsedPayload.Contact {
			pushoverDoorPush()
		}

		// fmt.Printf("* [%s] %s\n", msg.Topic(), string(msg.Payload()))
	})
}

type PushoverNotification struct {
	Token    string `json:"token"`
	User     string `json:"user"`
	Message  string `json:"message"`
	Priority int16  `json:"priority"`
}

func pushoverDoorPush() {
	marshalledNotification, err := json.Marshal(PushoverNotification{
		Token:    "",
		User:     "",
		Message:  "Door open!",
		Priority: 0,
	})
	if err != nil {
		log.Printf("Could not marshal the pushover notification")
		return
	}
	r, err := http.Post("https://api.pushover.net/1/messages.json", "application/json", bytes.NewReader(marshalledNotification))
	if err != nil {
		log.Printf("Could not post to pushover")
		return
	}
	log.Printf("%v+\n, ", r)
}

func loadLatestFromPostgres() bool {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", DbHost, DbPort, DbUser, DbPassword, DbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Could not connect to the database: %v", err)
		return true // Assume the door is closed
	}
	defer db.Close()

	var doorClosed bool
	err = db.QueryRow("SELECT metric_value FROM metrics WHERE metric_name = $1 ORDER BY created_at DESC LIMIT 1", "door_closed").Scan(&doorClosed)
	if err != nil {
		log.Printf("Could not insert into the database: %v", err)
		return true // Assume the door is closed
	}
	return doorClosed
}

func storeInPostgres(doorClosed bool) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", DbHost, DbPort, DbUser, DbPassword, DbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Could not connect to the database: %v", err)
		return
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO metrics (created_at, metric_name, metric_value) VALUES (NOW(), 'door_closed', $1)", doorClosed)
	if err != nil {
		log.Printf("Could not insert into the database: %v", err)
		return
	}
}

const (
	DbHost     = "localhost"
	DbPort     = 5432
	DbUser     = "mqtt2tsdb"
	DbPassword = ""
	DbName     = "mqtt2tsdb"
)

type metrics struct {
	doorClosedState prometheus.Gauge
}

func NewMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		// Create a summary to track fictional inter service RPC latencies for three
		// distinct services with different latency distributions. These services are
		// differentiated via a "service" label.
		doorClosedState: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "door_closed",
				Help: "Whether the door is closed or not",
			},
		),
	}
	reg.MustRegister(m.doorClosedState)
	return m
}

func main() {
	var wg sync.WaitGroup
	// TODO(janbrucek)(20250117) Make this configurable better
	// $ export CLOUDMQTT_URL=mqtt://localhost:1883/zigbee2mqtt/maindoor
	// $ ./mqtt2tsdb

	uri, err := url.Parse(os.Getenv("CLOUDMQTT_URL"))

	client := connect("sub", uri)

	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Using URL: %s", uri)
	topic := uri.Path[1:len(uri.Path)]
	if topic == "" {
		topic = "test"
	}

	log.Printf("Listening to %s", topic)
	wg.Add(1)
	go listen(client, uri, topic)

	topicsAndNames := map[string]string{
		"p1ib/p1ib_h_active_imp_q1_q4/state":   "energy_import",
		"p1ib/p1ib_active_power_p_q1_q4/state": "momentary_power_import",
		"p1ib/p1ib_voltage_l1/state":           "voltage",
		"p1ib/p1ib_current_l1/state":           "current",
		"p1ib/p1ib_rssi/state":                 "wifi_rssi",
	}
	for topic := range topicsAndNames {
		wg.Add(1)
		log.Printf("Listening to %s as %s", topic, topicsAndNames[topic])
		go simpleListen(client, uri, topic, topicsAndNames[topic])
	}

	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	initState := loadLatestFromPostgres()
	if initState {
		log.Printf("Door is closed")
		m.doorClosedState.Set(1)
	} else {
		log.Printf("Door is open")
		m.doorClosedState.Set(0)
	}

	// Add Go module build info.
	reg.MustRegister(collectors.NewBuildInfoCollector())

	http.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
			// Pass custom registry
			Registry: reg,
		},
	))
	log.Fatal(http.ListenAndServe(":8081", nil))

	wg.Wait()
}
