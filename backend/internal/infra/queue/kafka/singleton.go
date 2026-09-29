package kafka

import (
	"context"
	"sync"

	kafkago "github.com/segmentio/kafka-go"
)

var (
	globalBrokers []string
	brokerMu      sync.RWMutex
)

func SetBrokers(brokers []string) {
	brokerMu.Lock()
	defer brokerMu.Unlock()
	globalBrokers = brokers
}

func GetBrokers() []string {
	brokerMu.RLock()
	defer brokerMu.RUnlock()
	return globalBrokers
}

func Ping(ctx context.Context) error {
	brokerMu.RLock()
	brokers := globalBrokers
	brokerMu.RUnlock()

	if len(brokers) == 0 {
		return nil // not configured — treated as not_configured, not an error
	}

	conn, err := kafkago.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return err
	}
	return conn.Close()
}
