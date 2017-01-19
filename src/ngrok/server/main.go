package server

import (
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"ngrok/conn"
	log "ngrok/log"
	"ngrok/msg"
	"ngrok/util"
	"os"
	"runtime/debug"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	registryCacheSize uint64        = 1024 * 1024 // 1 MB
	connReadTimeout   time.Duration = 10 * time.Second
)

// GLOBALS
var (
	tunnelRegistry  *TunnelRegistry
	controlRegistry *ControlRegistry

	// XXX: kill these global variables - they're only used in tunnel.go for constructing forwarding URLs
	opts      *Options
	listeners map[string]*conn.Listener
	mqttc     *MQTT.Client
)

var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("Topic: %s\n", msg.Topic())
	fmt.Printf("%s\n", msg.Payload())
}

func NewMQTTClient(srv string) *MQTT.Client {
	opts := MQTT.NewClientOptions().AddBroker(srv)
	opts.SetClientID("magicwifi-176")
	opts.SetDefaultPublishHandler(f)
	opts.SetUsername("admin")
	opts.SetPassword("password")

	c := MQTT.NewClient(opts)

	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := c.Subscribe("ngrok-server/info", 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	return &c
}

func NewProxy(pxyConn conn.Conn, regPxy *msg.RegProxy) {
	// fail gracefully if the proxy connection fails to register
	defer func() {
		if r := recover(); r != nil {
			pxyConn.Warn("Failed with error: %v", r)
			pxyConn.Close()
		}
	}()

	// set logging prefix
	pxyConn.SetType("pxy")

	// look up the control connection for this proxy
	pxyConn.Info("Registering new proxy for %s", regPxy.ClientId)
	ctl := controlRegistry.Get(regPxy.ClientId)

	if ctl == nil {
		panic("No client found for identifier: " + regPxy.ClientId)
	}

	ctl.RegisterProxy(pxyConn)
}

// Listen for incoming control and proxy connections
// We listen for incoming control and proxy connections on the same port
// for ease of deployment. The hope is that by running on port 443, using
// TLS and running all connections over the same port, we can bust through
// restrictive firewalls.
func tunnelListener(addr string, tlsConfig *tls.Config) {
	// listen for incoming connections
	listener, err := conn.Listen(addr, "tun", tlsConfig)
	if err != nil {
		panic(err)
	}

	log.Info("Listening for control and proxy connections on %s", listener.Addr.String())
	for c := range listener.Conns {
		go func(tunnelConn conn.Conn) {
			// don't crash on panics
			defer func() {
				if r := recover(); r != nil {
					tunnelConn.Info("tunnelListener failed with error %v: %s", r, debug.Stack())
				}
			}()

			tunnelConn.SetReadDeadline(time.Now().Add(connReadTimeout))
			var rawMsg msg.Message
			if rawMsg, err = msg.ReadMsg(tunnelConn); err != nil {
				tunnelConn.Warn("Failed to read message: %v", err)
				tunnelConn.Close()
				return
			}

			// don't timeout after the initial read, tunnel heartbeating will kill
			// dead connections
			tunnelConn.SetReadDeadline(time.Time{})

			switch m := rawMsg.(type) {
			case *msg.Auth:
				NewControl(tunnelConn, m)

			case *msg.RegProxy:
				NewProxy(tunnelConn, m)

			default:
				tunnelConn.Close()
			}
		}(c)
	}
}

func MQTTtunnel() {
	for {
		defer (*mqttc).Disconnect(250)

		var s string
		for k := range controlRegistry.controls {
			control := controlRegistry.Get(k)
			if control.tunnels != nil {
				s += fmt.Sprintf("mac:%s rport:%d,lport:%d\n", control.auth.Mac, control.tunnels[0].req.RemotePort, control.tunnels[0].req.LocalPort)
			}
		}
		//fmt.Println(s)
		token := (*mqttc).Publish("ngrok-server/info", 0, false, s)
		token.Wait()

		time.Sleep(time.Second * 5)

		// t := controlRegistry.Get("11:22:33:44:55:66")
		// if t != nil {
		// 	fmt.Println("clientId=", t.id)
		// 	fmt.Println("port:", t.tunnels[0].req.RemotePort)

		// 	time.Sleep(time.Second * 5)
		// }
	}
}

func HTTPPost(url string) {
	for {
		var err error
		defer func(url string) {

			if r := recover(); r != nil {

				fmt.Println("Recovered in testPanic2Error", r)

				//check exactly what the panic was and create error.
				switch x := r.(type) {
				case string:
					err = errors.New(x)
				case error:
					err = x
				default:
					err = errors.New("Unknow panic")
				}
			}
			fmt.Println(err)
			HTTPPost(url)

		}(url)

		var s string
		for k := range controlRegistry.controls {
			control := controlRegistry.Get(k)
			if control.tunnels != nil {
				s += fmt.Sprintf("mac|%s,rport|%d,lport|%d\n", control.auth.Mac, control.tunnels[0].req.RemotePort, control.tunnels[0].req.LocalPort)
			}
		}

		//Http post string to server
		resp, err := http.Post(url,
			"application/x-www-form-urlencoded",
			strings.NewReader(s))

		if err != nil {
			fmt.Println(err)
		}

		resp.Body.Close()

		time.Sleep(time.Second * 5)
	}
}

func Main() {
	// parse options
	opts = parseArgs()

	// init logging
	log.LogTo(opts.logto, opts.loglevel)

	// seed random number generator
	seed, err := util.RandomSeed()
	if err != nil {
		panic(err)
	}
	rand.Seed(seed)

	// init tunnel/control registry
	registryCacheFile := os.Getenv("REGISTRY_CACHE_FILE")
	tunnelRegistry = NewTunnelRegistry(registryCacheSize, registryCacheFile)
	controlRegistry = NewControlRegistry()
	mqttc = NewMQTTClient(opts.mqtt)
	log.Info("MQTT Server %s", opts.mqtt)
	go MQTTtunnel()

	log.Info("HTTP Post url: %s", opts.posturl)
	go HTTPPost(opts.posturl)
	// start listeners
	listeners = make(map[string]*conn.Listener)

	// load tls configuration
	tlsConfig, err := LoadTLSConfig(opts.tlsCrt, opts.tlsKey)
	if err != nil {
		panic(err)
	}

	// listen for http
	if opts.httpAddr != "" {
		listeners["http"] = startHttpListener(opts.httpAddr, nil)
	}

	// listen for https
	if opts.httpsAddr != "" {
		listeners["https"] = startHttpListener(opts.httpsAddr, tlsConfig)
	}

	// ngrok clients
	tunnelListener(opts.tunnelAddr, tlsConfig)
}
