package logger

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// LogData represents the structured log format
type LogData struct {
	Timestamp     string `json:"timestamp"`
	Level         string `json:"level"`
	Message       string `json:"message,omitempty"`
	IPAddress     string `json:"ip_address,omitempty"`
	AppName       string `json:"appname"`
	Hostname      string `json:"hostname,omitempty"`
	TransactionID string `json:"tr_id,omitempty"`
	Channel       string `json:"channel,omitempty"`
	BankCode      string `json:"bank_code,omitempty"`
	ReferenceID   string `json:"reference_id,omitempty"`
	RRN           string `json:"rrn,omitempty"`
	PublishID     string `json:"publish_id,omitempty"`
	CFTrID        string `json:"cf_trid,omitempty"`
	DeviceInfo    string `json:"device_info,omitempty"`
	ParamA        string `json:"param_a,omitempty"`
	ParamB        string `json:"param_b,omitempty"`
	ParamC        string `json:"param_c,omitempty"`
}

// Logger struct
type Logger struct {
	logger      *logrus.Logger
	GraylogHost string
	GraylogPort string
	Protocol    string // "udp" or "tcp"
}

// NewLogger initializes a new logger with the chosen protocol
func NewLogger(graylogHost, graylogPort, protocol string) *Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})
	l.SetOutput(os.Stdout)

	// Validate protocol
	if protocol != "udp" && protocol != "tcp" {
		fmt.Println("Invalid protocol! Defaulting to UDP.")
		protocol = "udp"
	}

	return &Logger{
		logger:      l,
		GraylogHost: graylogHost,
		GraylogPort: graylogPort,
		Protocol:    protocol,
	}
}

// Log logs a message and sends it to Graylog
func (l *Logger) Log(level, message string, data LogData) {
	// Automatically set timestamp, hostname, and IP dynamically
	data.Timestamp = time.Now().UTC().Format(time.RFC3339)
	data.Message = message

	// Set dynamic hostname and IP if not already provided

	hostname, err := os.Hostname()
	if err == nil {
		data.Hostname = hostname
	} else {
		data.Hostname = "Unknown"
	}

	data.IPAddress = GetLocalIP()

	jsonData, _ := json.Marshal(data)

	// Log locally
	switch level {
	case "INFO":
		l.logger.Info(string(jsonData))
	case "ERROR":
		l.logger.Error(string(jsonData))
	case "DEBUG":
		l.logger.Debug(string(jsonData))
	default:
		l.logger.Warn(string(jsonData))
	}

	// Send to Graylog using the chosen protocol
	l.sendToGraylog(jsonData)
}

// sendToGraylog sends log data to Graylog using the selected protocol
func (l *Logger) sendToGraylog(logData []byte) {
	address := fmt.Sprintf("%s:%s", l.GraylogHost, l.GraylogPort)

	if l.Protocol == "udp" {
		err := sendUDP(address, logData)
		if err != nil {
			fmt.Println("Failed to send log via UDP:", err)
		} else {
			fmt.Println("Log sent successfully to Graylog via UDP!")
		}
	} else {
		err := sendTCP(address, logData)
		if err != nil {
			fmt.Println("Failed to send log via TCP:", err)
		} else {
			fmt.Println("Log sent successfully to Graylog via TCP!")
		}
	}
}

// sendUDP sends log data over UDP
func sendUDP(address string, data []byte) error {
	conn, err := net.Dial("udp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(data)
	return err
}

// sendTCP sends log data over TCP
func sendTCP(address string, data []byte) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(append(data, '\n')) // GELF messages should end with a newline
	return err
}

// GetLocalIP retrieves the local machine's IP address.
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "Unknown"
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return "Unknown"
}
