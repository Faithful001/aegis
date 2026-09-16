package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn	*amqp.Connection
	channel	*amqp.Channel

	config	RabbitMQConfig
	mu		sync.RWMutex

	ctx		context.Context
	cancel	context.CancelFunc
}

type RabbitMQConfig struct {
	URL 				string
	ReconnectionDelay 	time.Duration
	MaxReconnect		int
}

type QueueConfig struct {
	Name 		string
	Durable      bool
	AutoDelete   bool
	Exclusive    bool
	NoWait       bool
	Args         amqp.Table
}

type PublishConfig struct {
	Exchange	string
	RoutingKey	string
	Mandatory	bool
	Immediate	bool
	Message		amqp.Publishing
}

type ConsumeConfig struct {
	Queue		string
	Consumer		string
	AutoAck		bool
	Exclusive	bool
	NoLocal		bool
	NoWait		bool
	Args		amqp.Table
}

func DefaultConfig() RabbitMQConfig {
	return RabbitMQConfig{
		URL: "amqp://admin:admin123@127.0.0.1:5672/",
		ReconnectionDelay: 5 * time.Second,
		MaxReconnect:  5,
	}
}


func NewRabbitMQ(config RabbitMQConfig) (*RabbitMQ, error) {
	if config.URL == ""{
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())
	r := &RabbitMQ{
		config: config,
		ctx: ctx,
		cancel: cancel,
	}

	if err := r.Connect(); err != nil {
		cancel()
		return nil, fmt.Errorf("Failed to connect to rabbitmq %w", err)
	}

	go r.MonitorConnection()
	
	return r, nil
}

func (r *RabbitMQ) Connect() error {
	var lastErr error
	// iterate through the max reconnect
	for i := 0; i <= r.config.MaxReconnect; i++ {
		if i > 0 {
			time.Sleep(r.config.ReconnectionDelay)
			log.Printf("RabbitMQ reconnect attempt %d/%d", i, r.config.MaxReconnect)
		}
		conn, err := amqp.Dial(r.config.URL)

		if err != nil {
			lastErr = err
			continue
		}

		ch, err := conn.Channel()

		if err != nil {
			lastErr = err
			continue
		}

		r.mu.Lock()
		r.conn = conn
		r.channel = ch
		r.mu.Unlock()

		return nil
	} 
	return fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %v", r.config.MaxReconnect, lastErr)
}


func (r *RabbitMQ) DeclareQueue(cfg QueueConfig) (amqp.Queue, error) {
	r.mu.Lock()
	ch := r.channel
	r.mu.Unlock()

	if ch == nil {
		return amqp.Queue{}, errors.New("rabbitmq not connected")
	}

	q, err := ch.QueueDeclare(
		cfg.Name,
		cfg.Durable,
		cfg.AutoDelete,
		cfg.Exclusive,
		cfg.NoWait,
		cfg.Args,
	)

	if err != nil {
		return amqp.Queue{}, err
	}

	return q, nil
}


func (r *RabbitMQ) DeclareExchange(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	if ch == nil {
		return errors.New("rabbitmq not connected")
	}

	err := ch.ExchangeDeclare(
		name,
		kind,
		durable,
		autoDelete,
		internal,
		noWait,
		args,
	)
	if err != nil {
		return fmt.Errorf("exchange declare failed: %w", err)
	}
	return nil
}

func (r *RabbitMQ) BindQueue(queueName, routingKey, exchangeName string, noWait bool, args amqp.Table) error {
	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	if ch == nil {
		return errors.New("rabbitmq not connected")
	}

	err := ch.QueueBind(
		queueName,
		routingKey,
		exchangeName,
		noWait,
		args,
	)
	if err != nil {
		return fmt.Errorf("queue bind failed: %w", err)
	}
	return nil
}

func (r *RabbitMQ) Publish(ctx context.Context, cfg PublishConfig) error {
	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	if ch == nil {
		return errors.New("rabbitmq not connected")
	}

	err := ch.PublishWithContext(
		ctx,
		cfg.Exchange,
		cfg.RoutingKey,
		cfg.Mandatory,
		cfg.Immediate,
		cfg.Message,
	)

	if err != nil {
		return fmt.Errorf("publish failed: %w", err)
	}

	return nil
}

func (r *RabbitMQ) Consume(cfg ConsumeConfig) (<- chan amqp.Delivery, error) {
	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	if ch == nil {
		return nil, errors.New("rabbitmq not connected")
	}

	msgs, err := ch.Consume(
		cfg.Queue,
		cfg.Consumer,
		cfg.AutoAck,
		cfg.Exclusive,
		cfg.NoLocal,
		cfg.NoWait,
		cfg.Args,
	)

	if err != nil {
		return nil, fmt.Errorf("consume failed: %w", err)
	}

	return msgs, nil
}

func (r *RabbitMQ) MonitorConnection() {
	for {
		r.mu.RLock()
		ch := r.channel
		r.mu.RUnlock()

		if ch == nil {
			return
		}

		closeCh := ch.NotifyClose(make(chan *amqp.Error))

		select {
		case err := <-closeCh:
			if err != nil {
				log.Printf("RabbitMQ connection lost: %v", err)

				if err := r.Connect(); err != nil {
					log.Printf("Reconnect failed: %v", err)
				}
			}

		case <-r.ctx.Done():
			return
		}
	}
}

func (r *RabbitMQ) Close() error {
	r.cancel()

	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error

	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("channel close failed: %w", err))
		}
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("connection close failed: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("rabbitmq close errors: %w", errors.Join(errs...))
	}

	log.Println("RabbitMQ closed cleanly")
	return nil
}