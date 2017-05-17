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
	posturl_en int
	posturl    string
	mqtt_en    int
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
	posturl_en := flag.Int("posturl-enable", 0, "Enable post to url assigned by posturl, Default: disable")
	mqtt_en := flag.Int("mqtt-enable", 0, "Enable send mqtt message to mqtt server, Default: disable")
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
		mqtt_en:    *mqtt_en,
		posturl_en: *posturl_en,
	}
}
