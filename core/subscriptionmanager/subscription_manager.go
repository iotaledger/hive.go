// SubscriptionManager keeps track of subscribed topics of clients.
// This allows to get notified when a client connects or disconnects
// or a topic is subscribed or unsubscribed.
package subscriptionmanager

import (
	"errors"
	"sync"

	"github.com/iotaledger/hive.go/core/generics/constraints"
	"github.com/iotaledger/hive.go/core/generics/event"
	"github.com/iotaledger/hive.go/core/generics/options"
	"github.com/iotaledger/hive.go/core/generics/shrinkingmap"
)

var (
	ErrMaxTopicSubscriptionsPerClientReached = errors.New("maximum amount of topic subscriptions per client reached")
)

type ClientID interface {
	constraints.Integer | ~string
}

type Topic interface {
	constraints.Integer | ~string
}

type ClientEvent[K ClientID] struct {
	ClientID K
}

type TopicEvent[V Topic] struct {
	Topic V
}

type ClientTopicEvent[K ClientID, V Topic] struct {
	ClientID K
	Topic    V
}

type DropClientEvent[K ClientID] struct {
	ClientID K
	Reason   error
}

// Events contains all the events that are triggered by the SubscriptionManager.
type Events[K ClientID, V Topic] struct {
	// ClientConnected event is triggered when a new client connected.
	ClientConnected *event.Event[*ClientEvent[K]]
	// ClientDisconnected event is triggered when a client disconnected.
	ClientDisconnected *event.Event[*ClientEvent[K]]
	// TopicSubscribed event is triggered when a client subscribed to a topic.
	TopicSubscribed *event.Event[*ClientTopicEvent[K, V]]
	// TopicUnsubscribed event is triggered when a client unsubscribed from a topic.
	TopicUnsubscribed *event.Event[*ClientTopicEvent[K, V]]
	// TopicAdded event is triggered when a topic is subscribed for the first time by any client.
	TopicAdded *event.Event[*TopicEvent[V]]
	// TopicRemoved event is triggered when a topic is not subscribed anymore by any client.
	TopicRemoved *event.Event[*TopicEvent[V]]
	// DropClient event is triggered when a client should be dropped.
	DropClient *event.Event[*DropClientEvent[K]]
}

func newEvents[K ClientID, V Topic]() *Events[K, V] {
	return &Events[K, V]{
		ClientConnected:    event.New[*ClientEvent[K]](),
		ClientDisconnected: event.New[*ClientEvent[K]](),
		TopicSubscribed:    event.New[*ClientTopicEvent[K, V]](),
		TopicUnsubscribed:  event.New[*ClientTopicEvent[K, V]](),
		TopicAdded:         event.New[*TopicEvent[V]](),
		TopicRemoved:       event.New[*TopicEvent[V]](),
		DropClient:         event.New[*DropClientEvent[K]](),
	}
}

// SubscriptionManager keeps track of subscribed topics of clients.
// This allows to get notified when a client connects or disconnects
// or a topic is subscribed or unsubscribed.
type SubscriptionManager[K ClientID, V Topic] struct {
	sync.RWMutex

	// subscribers keeps track of the clients and their
	// subscribed topics (and the count of subscriptions per topic).
	subscribers *shrinkingmap.ShrinkingMap[K, *shrinkingmap.ShrinkingMap[V, int]]
	topics      *shrinkingmap.ShrinkingMap[V, int]

	maxTopicSubscriptionsPerClient int
	cleanupThresholdCount          int
	cleanupThresholdRatio          float32

	events *Events[K, V]
}

// WithMaxTopicSubscriptionsPerClient defines the max amount of subscriptions
// per client before the client is seen as malicious and gets dropped.
// 0 = deactivated (default).
func WithMaxTopicSubscriptionsPerClient[K ClientID, V Topic](maxTopicSubscriptionsPerClient int) options.Option[SubscriptionManager[K, V]] {
	return func(s *SubscriptionManager[K, V]) {
		s.maxTopicSubscriptionsPerClient = maxTopicSubscriptionsPerClient
	}
}

// WithShrinkingThresholdCount defines the count of
// deletions that triggers shrinking of the map.
func WithCleanupThresholdCount[K ClientID, V Topic](cleanupThresholdCount int) options.Option[SubscriptionManager[K, V]] {
	return func(s *SubscriptionManager[K, V]) {
		s.cleanupThresholdCount = cleanupThresholdCount
	}
}

// WithShrinkingThresholdRatio defines the ratio between the amount
// of deleted keys and the current map's size before shrinking is triggered.
func WithCleanupThresholdRatio[K ClientID, V Topic](cleanupThresholdRatio float32) options.Option[SubscriptionManager[K, V]] {
	return func(s *SubscriptionManager[K, V]) {
		s.cleanupThresholdRatio = cleanupThresholdRatio
	}
}

func New[K ClientID, V Topic](opts ...options.Option[SubscriptionManager[K, V]]) *SubscriptionManager[K, V] {

	manager := options.Apply(&SubscriptionManager[K, V]{
		maxTopicSubscriptionsPerClient: 1000,
		cleanupThresholdCount:          10000,
		cleanupThresholdRatio:          1.0,
		events:                         newEvents[K, V](),
	}, opts)

	manager.subscribers = shrinkingmap.New[K, *shrinkingmap.ShrinkingMap[V, int]](
		shrinkingmap.WithShrinkingThresholdRatio(manager.cleanupThresholdRatio),
		shrinkingmap.WithShrinkingThresholdCount(manager.cleanupThresholdCount),
	)
	manager.topics = shrinkingmap.New[V, int](
		shrinkingmap.WithShrinkingThresholdRatio(manager.cleanupThresholdRatio),
		shrinkingmap.WithShrinkingThresholdCount(manager.cleanupThresholdCount),
	)

	return manager
}

func (s *SubscriptionManager[K, V]) Events() *Events[K, V] {
	return s.events
}

func (s *SubscriptionManager[K, V]) Connect(clientID K) {
	s.Lock()
	defer s.Unlock()

	// in case the client already exists, we cleanup old subscriptions.
	s.cleanupClientWithoutLocking(clientID)

	// create a new map for the client
	s.subscribers.Set(clientID, shrinkingmap.New[V, int](
		shrinkingmap.WithShrinkingThresholdRatio(s.cleanupThresholdRatio),
		shrinkingmap.WithShrinkingThresholdCount(s.cleanupThresholdCount),
	))

	s.events.ClientConnected.Trigger(&ClientEvent[K]{ClientID: clientID})
}

func (s *SubscriptionManager[K, V]) Disconnect(clientID K) {
	s.Lock()
	defer s.Unlock()

	// cleanup the client
	s.cleanupClientWithoutLocking(clientID)

	// send disconnect notification then delete the subscriber
	s.events.ClientDisconnected.Trigger(&ClientEvent[K]{ClientID: clientID})
}

func (s *SubscriptionManager[K, V]) Subscribe(clientID K, topic V) {
	s.Lock()
	defer s.Unlock()

	// check if the client is connected
	subscribedTopics, has := s.subscribers.Get(clientID)
	if !has {
		return
	}

	count, has := subscribedTopics.Get(topic)
	if has {
		subscribedTopics.Set(topic, count+1)
	} else {
		// add a new topic
		subscribedTopics.Set(topic, 1)

		// check if the client has reached the max number of subscriptions
		if s.maxTopicSubscriptionsPerClient != 0 && subscribedTopics.Size() >= s.maxTopicSubscriptionsPerClient {
			// cleanup the client
			s.cleanupClientWithoutLocking(clientID)
			// drop the client
			s.events.DropClient.Trigger(&DropClientEvent[K]{ClientID: clientID, Reason: ErrMaxTopicSubscriptionsPerClientReached})

			// do not fire the subscribed events
			return
		}
	}

	// global topics map
	count, has = s.topics.Get(topic)
	if has {
		s.topics.Set(topic, count+1)
	} else {
		// add a new topic
		s.topics.Set(topic, 1)
		s.events.TopicAdded.Trigger(&TopicEvent[V]{Topic: topic})
	}

	s.events.TopicSubscribed.Trigger(&ClientTopicEvent[K, V]{ClientID: clientID, Topic: topic})
}

func (s *SubscriptionManager[K, V]) Unsubscribe(clientID K, topic V) {
	s.Lock()
	defer s.Unlock()

	// check if the client is connected
	subscribedTopics, has := s.subscribers.Get(clientID)
	if !has {
		return
	}

	count, has := subscribedTopics.Get(topic)
	if has {
		if count <= 1 {
			// delete the topic
			subscribedTopics.Delete(topic)
		} else {
			subscribedTopics.Set(topic, count-1)
		}
	}

	// global topics map
	count, has = s.topics.Get(topic)
	if has {
		if count <= 1 {
			// delete the topic
			s.topics.Delete(topic)
			s.events.TopicRemoved.Trigger(&TopicEvent[V]{Topic: topic})
		} else {
			s.topics.Set(topic, count-1)
		}
	}

	s.events.TopicUnsubscribed.Trigger(&ClientTopicEvent[K, V]{ClientID: clientID, Topic: topic})
}

func (s *SubscriptionManager[K, V]) HasSubscribers(topic V) bool {
	s.RLock()
	defer s.RUnlock()

	_, hasSubscribers := s.topics.Get(topic)

	return hasSubscribers
}

// SubscribersSize returns the size of the underlying map of the SubscriptionManager.
func (s *SubscriptionManager[K, V]) SubscribersSize() int {
	s.RLock()
	defer s.RUnlock()

	return s.subscribers.Size()
}

// TopicsSize returns the size of the underlying map of the SubscriptionManager.
func (s *SubscriptionManager[K, V]) TopicsSize() int {
	s.RLock()
	defer s.RUnlock()

	return s.topics.Size()
}

// TopicsSize returns the size of all underlying maps of the SubscriptionManager.
func (s *SubscriptionManager[K, V]) TopicsSizeAll() int {
	s.RLock()
	defer s.RUnlock()

	count := 0

	// loop over all clients
	s.subscribers.ForEach(func(clientID K, topics *shrinkingmap.ShrinkingMap[V, int]) bool {
		count += topics.Size()

		return true
	})

	return count
}

// cleanupClientWithoutLocking removes all subscriptions and the client itself.
func (s *SubscriptionManager[K, V]) cleanupClientWithoutLocking(clientID K) {

	// check if the client exists
	subscribedTopics, has := s.subscribers.Get(clientID)
	if !has {
		return
	}

	// loop over all topics and delete them
	subscribedTopics.ForEach(func(topic V, count int) bool {

		// global topics map
		topicsCount, has := s.topics.Get(topic)
		if has {
			if topicsCount-count <= 1 {
				// delete the topic
				s.topics.Delete(topic)
				s.events.TopicRemoved.Trigger(&TopicEvent[V]{Topic: topic})
			} else {
				s.topics.Set(topic, topicsCount-count)
			}
		}

		// call the topic unsubscribe as many times as it was subscribed
		for i := 0; i < count; i++ {
			s.events.TopicUnsubscribed.Trigger(&ClientTopicEvent[K, V]{ClientID: clientID, Topic: topic})
		}

		// delete the topic
		subscribedTopics.Delete(topic)

		return true
	})

	// delete the client
	s.subscribers.Delete(clientID)
}
