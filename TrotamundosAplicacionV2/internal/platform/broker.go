package platform

import (
	"hash/fnv"
	"sync"
	"time"
	"trotamundos/internal/domain"
)

const (
	TopicTripLifecycle       = "trip.lifecycle"
	TopicReservationDetected = "reservation.detected"
	TopicCatalogBooked       = "catalog.booked"
	TopicEmailReceived       = "email.received"

	brokerPartitions = 4
	brokerQueueSize  = 256
)

// Event represents an asynchronous domain message routed across partitioned workers.
type Event struct {
	ID        string    `json:"id"`
	Topic     string    `json:"topic"`
	Key       string    `json:"key"`
	Payload   any       `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

// EventHandler processes a single Event on a partition worker.
type EventHandler func(event Event)

// TripLifecycleMessage carries anonymized or domain lifecycle transitions.
type TripLifecycleMessage struct {
	EventID         string
	TripID          string
	UserID          string
	EventType       string
	DestinationCity string
	CountryCode     string
	TransportTypes  []domain.ReservationType
	Disrupted       bool
	ProviderName    string
	Timestamp       time.Time
}

// ReservationDetectedMessage carries newly detected or imported reservation segments.
type ReservationDetectedMessage struct {
	EventID          string
	UserID           string
	Source           domain.ReservationSource
	ProviderName     string
	Type             domain.ReservationType
	Status           domain.ReservationStatus
	ConfirmationCode string
	Origin           string
	Destination      string
	StartDate        time.Time
	EndDate          time.Time
	Price            float64
	Currency         string
	Notes            string
	Details          domain.ReservationDetails
	Timestamp        time.Time
}

type queuedEvent struct {
	event Event
	done  chan struct{}
}

// EventBroker dispatches events across key-hashed partitions in memory.
type EventBroker struct {
	mu          sync.RWMutex
	subscribers map[string][]EventHandler
	partitions  [brokerPartitions]chan queuedEvent
	stopCh      chan struct{}
	closeOnce   sync.Once
}

func NewEventBroker() *EventBroker {
	b := &EventBroker{
		subscribers: make(map[string][]EventHandler),
		stopCh:      make(chan struct{}),
	}
	for i := 0; i < brokerPartitions; i++ {
		b.partitions[i] = make(chan queuedEvent, brokerQueueSize)
		go b.runPartition(b.partitions[i])
	}
	return b
}

func (b *EventBroker) Subscribe(topic string, handler EventHandler) {
	if topic == "" || handler == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[topic] = append(b.subscribers[topic], handler)
}

func (b *EventBroker) Publish(topic, key string, payload any) {
	ev := Event{
		ID:        NewID("evt"),
		Topic:     topic,
		Key:       key,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
	idx := partitionIndex(topic, key)
	item := queuedEvent{
		event: ev,
		done:  make(chan struct{}),
	}

	select {
	case b.partitions[idx] <- item:
		// Wait briefly so fast in-memory subscribers complete deterministically while slow I/O stays non-blocking.
		select {
		case <-item.done:
		case <-time.After(40 * time.Millisecond):
		}
	default:
		b.dispatch(ev)
	}
}

func (b *EventBroker) Close() {
	b.closeOnce.Do(func() {
		close(b.stopCh)
	})
}

func (b *EventBroker) runPartition(ch <-chan queuedEvent) {
	for {
		select {
		case <-b.stopCh:
			return
		case item := <-ch:
			b.dispatch(item.event)
			close(item.done)
		}
	}
}

func (b *EventBroker) dispatch(ev Event) {
	b.mu.RLock()
	handlers := append([]EventHandler(nil), b.subscribers[ev.Topic]...)
	b.mu.RUnlock()

	for _, h := range handlers {
		func(fn EventHandler) {
			defer func() { _ = recover() }()
			fn(ev)
		}(h)
	}
}

func partitionIndex(topic, key string) int {
	h := fnv.New32a()
	if key != "" {
		_, _ = h.Write([]byte(key))
	} else {
		_, _ = h.Write([]byte(topic))
	}
	return int(h.Sum32() % brokerPartitions)
}
