package server

import (
	"flag"
)

type Options struct {
	httpAddr   string
	httpsAddr  string
	tunnelAddr string
	domain     string
	tlsCrt     string
	tlsKey     string
	logto      string
	loglevel   string
	posturl    string
	mqtt       string
}

func parseArgs() *Options {
	httpAddr := flag.String("httpAddr", "", "Public address for HTTP connections, empty string to disable")
	httpsAddr := flag.String("httpsAddr", "", "Public address listening for HTTPS connections, emptry string to disable")
	tunnelAddr := flag.String("tunnelAddr", ":14443", "Public address listening for ngrok client")
	domain := flag.String("domain", "ngrok.com", "Domain where the tunnels are hosted")
	tlsCrt := flag.String("tlsCrt", "", "Path to a TLS certificate file")
	tlsKey := flag.String("tlsKey", "", "Path to a TLS key file")
	logto := flag.String("log", "stdout", "Write log messages to this file. 'stdout' and 'none' have special meanings")
	loglevel := flag.String("log-level", "DEBUG", "The level of messages to log. One of: DEBUG, INFO, WARNING, ERROR")
	posturl := flag.String("posturl", "http://magicwifi.com.cn/v3/api/device/vpn", "The http post url used to send current existence mac-rport-lport pair")
	mqtt := flag.String("mqtt", "tcp://hiweeds.net:1883", "MQTT server/broker this client connect to and publish message")
	flag.Parse()

	return &Options{
		httpAddr:   *httpAddr,
		httpsAddr:  *httpsAddr,
		tunnelAddr: *tunnelAddr,
		domain:     *domain,
		tlsCrt:     *tlsCrt,
		tlsKey:     *tlsKey,
		logto:      *logto,
		loglevel:   *loglevel,
		posturl:    *posturl,
		mqtt:       *mqtt,
	}
}
