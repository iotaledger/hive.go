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
	comparable
}

type Topic interface {
	constraints.Integer | ~string
}

type ClientEvent[C ClientID] struct {
	ClientID C
}

type TopicEvent[T Topic] struct {
	Topic T
}

type ClientTopicEvent[C ClientID, T Topic] struct {
	ClientID C
	Topic    T
}

type DropClientEvent[C ClientID] struct {
	ClientID C
	Reason   error
}

// Events contains all the events that are triggered by the SubscriptionManager.
type Events[C ClientID, T Topic] struct {
	// ClientConnected event is triggered when a new client connected.
	ClientConnected *event.Event[*ClientEvent[C]]
	// ClientDisconnected event is triggered when a client disconnected.
	ClientDisconnected *event.Event[*ClientEvent[C]]
	// TopicSubscribed event is triggered when a client subscribed to a topic.
	TopicSubscribed *event.Event[*ClientTopicEvent[C, T]]
	// TopicUnsubscribed event is triggered when a client unsubscribed from a topic.
	TopicUnsubscribed *event.Event[*ClientTopicEvent[C, T]]
	// TopicAdded event is triggered when a topic is subscribed for the first time by any client.
	TopicAdded *event.Event[*TopicEvent[T]]
	// TopicRemoved event is triggered when a topic is not subscribed anymore by any client.
	TopicRemoved *event.Event[*TopicEvent[T]]
	// DropClient event is triggered when a client should be dropped.
	DropClient *event.Event[*DropClientEvent[C]]
}

func newEvents[C ClientID, T Topic]() *Events[C, T] {
	return &Events[C, T]{
		ClientConnected:    event.New[*ClientEvent[C]](),
		ClientDisconnected: event.New[*ClientEvent[C]](),
		TopicSubscribed:    event.New[*ClientTopicEvent[C, T]](),
		TopicUnsubscribed:  event.New[*ClientTopicEvent[C, T]](),
		TopicAdded:         event.New[*TopicEvent[T]](),
		TopicRemoved:       event.New[*TopicEvent[T]](),
		DropClient:         event.New[*DropClientEvent[C]](),
	}
}

// SubscriptionManager keeps track of subscribed topics of clients.
// This allows to get notified when a client connects or disconnects
// or a topic is subscribed or unsubscribed.
type SubscriptionManager[C ClientID, T Topic] struct {
	sync.RWMutex

	// subscribers keeps track of the clients and their
	// subscribed topics (and the count of subscriptions per topic).
	subscribers *shrinkingmap.ShrinkingMap[C, *shrinkingmap.ShrinkingMap[T, int]]
	topics      *shrinkingmap.ShrinkingMap[T, int]

	maxTopicSubscriptionsPerClient int
	cleanupThresholdCount          int
	cleanupThresholdRatio          float32

	events *Events[C, T]
}

// WithMaxTopicSubscriptionsPerClient defines the max amount of subscriptions
// per client before the client is seen as malicious and gets dropped.
// 0 = deactivated (default).
func WithMaxTopicSubscriptionsPerClient[C ClientID, T Topic](maxTopicSubscriptionsPerClient int) options.Option[SubscriptionManager[C, T]] {
	return func(s *SubscriptionManager[C, T]) {
		s.maxTopicSubscriptionsPerClient = maxTopicSubscriptionsPerClient
	}
}

// WithShrinkingThresholdCount defines the count of
// deletions that triggers shrinking of the map.
func WithCleanupThresholdCount[C ClientID, T Topic](cleanupThresholdCount int) options.Option[SubscriptionManager[C, T]] {
	return func(s *SubscriptionManager[C, T]) {
		s.cleanupThresholdCount = cleanupThresholdCount
	}
}

// WithShrinkingThresholdRatio defines the ratio between the amount
// of deleted keys and the current map's size before shrinking is triggered.
func WithCleanupThresholdRatio[C ClientID, T Topic](cleanupThresholdRatio float32) options.Option[SubscriptionManager[C, T]] {
	return func(s *SubscriptionManager[C, T]) {
		s.cleanupThresholdRatio = cleanupThresholdRatio
	}
}

func New[C ClientID, T Topic](opts ...options.Option[SubscriptionManager[C, T]]) *SubscriptionManager[C, T] {

	manager := options.Apply(&SubscriptionManager[C, T]{
		maxTopicSubscriptionsPerClient: 0,
		cleanupThresholdCount:          10000,
		cleanupThresholdRatio:          1.0,
		events:                         newEvents[C, T](),
	}, opts)

	manager.subscribers = shrinkingmap.New[C, *shrinkingmap.ShrinkingMap[T, int]](
		shrinkingmap.WithShrinkingThresholdRatio(manager.cleanupThresholdRatio),
		shrinkingmap.WithShrinkingThresholdCount(manager.cleanupThresholdCount),
	)
	manager.topics = shrinkingmap.New[T, int](
		shrinkingmap.WithShrinkingThresholdRatio(manager.cleanupThresholdRatio),
		shrinkingmap.WithShrinkingThresholdCount(manager.cleanupThresholdCount),
	)

	return manager
}

func (s *SubscriptionManager[C, T]) Events() *Events[C, T] {
	return s.events
}

func (s *SubscriptionManager[C, T]) Connect(clientID C) {

	var removedTopics, unsubscribedTopics []T

	// inline function used to release the lock before firing the event
	func() {
		s.Lock()
		defer s.Unlock()

		// in case the client already exists, we cleanup old subscriptions.
		removedTopics, unsubscribedTopics = s.cleanupClientWithoutLocking(clientID)

		// create a new map for the client
		s.subscribers.Set(clientID, shrinkingmap.New[T, int](
			shrinkingmap.WithShrinkingThresholdRatio(s.cleanupThresholdRatio),
			shrinkingmap.WithShrinkingThresholdCount(s.cleanupThresholdCount),
		))
	}()

	for _, topic := range removedTopics {
		s.events.TopicRemoved.Trigger(&TopicEvent[T]{Topic: topic})
	}
	for _, topic := range unsubscribedTopics {
		s.events.TopicUnsubscribed.Trigger(&ClientTopicEvent[C, T]{ClientID: clientID, Topic: topic})
	}

	s.events.ClientConnected.Trigger(&ClientEvent[C]{ClientID: clientID})
}

func (s *SubscriptionManager[C, T]) Disconnect(clientID C) {

	var removedTopics, unsubscribedTopics []T

	// inline function used to release the lock before firing the event
	func() {
		s.Lock()
		defer s.Unlock()

		// cleanup the client
		removedTopics, unsubscribedTopics = s.cleanupClientWithoutLocking(clientID)
	}()

	for _, topic := range removedTopics {
		s.events.TopicRemoved.Trigger(&TopicEvent[T]{Topic: topic})
	}
	for _, topic := range unsubscribedTopics {
		s.events.TopicUnsubscribed.Trigger(&ClientTopicEvent[C, T]{ClientID: clientID, Topic: topic})
	}

	// send disconnect notification then delete the subscriber
	s.events.ClientDisconnected.Trigger(&ClientEvent[C]{ClientID: clientID})
}

func (s *SubscriptionManager[C, T]) Subscribe(clientID C, topic T) {

	clientDropped := false
	topicAdded := false

	var removedTopics, unsubscribedTopics []T

	// inline function used to release the lock before firing the event
	func() {
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
				removedTopics, unsubscribedTopics = s.cleanupClientWithoutLocking(clientID)
				clientDropped = true

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
			topicAdded = true
		}
	}()

	for _, topic := range removedTopics {
		s.events.TopicRemoved.Trigger(&TopicEvent[T]{Topic: topic})
	}
	for _, topic := range unsubscribedTopics {
		s.events.TopicUnsubscribed.Trigger(&ClientTopicEvent[C, T]{ClientID: clientID, Topic: topic})
	}

	if clientDropped {
		// drop the client
		s.events.DropClient.Trigger(&DropClientEvent[C]{ClientID: clientID, Reason: ErrMaxTopicSubscriptionsPerClientReached})

		// do not fire the subscribed events
		return
	}

	if topicAdded {
		s.events.TopicAdded.Trigger(&TopicEvent[T]{Topic: topic})
	}

	s.events.TopicSubscribed.Trigger(&ClientTopicEvent[C, T]{ClientID: clientID, Topic: topic})
}

func (s *SubscriptionManager[C, T]) Unsubscribe(clientID C, topic T) {

	topicRemoved := false
	topicUnsubscribed := false

	// inline function used to release the lock before firing the event
	func() {
		s.Lock()
		defer s.Unlock()

		// check if the client is connected
		subscribedTopics, has := s.subscribers.Get(clientID)
		if !has {
			return
		}

		topicUnsubscribed = true

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
				topicRemoved = true
			} else {
				s.topics.Set(topic, count-1)
			}
		}
	}()

	if !topicUnsubscribed {
		// do not fire the unsubscribed events
		return
	}

	if topicRemoved {
		s.events.TopicRemoved.Trigger(&TopicEvent[T]{Topic: topic})
	}

	s.events.TopicUnsubscribed.Trigger(&ClientTopicEvent[C, T]{ClientID: clientID, Topic: topic})
}

func (s *SubscriptionManager[C, T]) HasSubscribers(topic T) bool {
	s.RLock()
	defer s.RUnlock()

	_, hasSubscribers := s.topics.Get(topic)

	return hasSubscribers
}

// SubscribersSize returns the size of the underlying map of the SubscriptionManager.
func (s *SubscriptionManager[C, T]) SubscribersSize() int {
	s.RLock()
	defer s.RUnlock()

	return s.subscribers.Size()
}

// TopicsSize returns the size of the underlying map of the SubscriptionManager.
func (s *SubscriptionManager[C, T]) TopicsSize() int {
	s.RLock()
	defer s.RUnlock()

	return s.topics.Size()
}

// TopicsSize returns the size of all underlying maps of the SubscriptionManager.
func (s *SubscriptionManager[C, T]) TopicsSizeAll() int {
	s.RLock()
	defer s.RUnlock()

	count := 0

	// loop over all clients
	s.subscribers.ForEach(func(clientID C, topics *shrinkingmap.ShrinkingMap[T, int]) bool {
		count += topics.Size()

		return true
	})

	return count
}

// cleanupClientWithoutLocking removes all subscriptions and the client itself.
func (s *SubscriptionManager[C, T]) cleanupClientWithoutLocking(clientID C) ([]T, []T) {

	removedTopics := make([]T, 0)
	unsubscribedTopics := make([]T, 0)

	// check if the client exists
	subscribedTopics, has := s.subscribers.Get(clientID)
	if !has {
		return removedTopics, unsubscribedTopics
	}

	// loop over all topics and delete them
	subscribedTopics.ForEach(func(topic T, count int) bool {

		// global topics map
		topicsCount, has := s.topics.Get(topic)
		if has {
			if topicsCount-count <= 1 {
				// delete the topic
				s.topics.Delete(topic)
				removedTopics = append(removedTopics, topic)
			} else {
				s.topics.Set(topic, topicsCount-count)
			}
		}

		// call the topic unsubscribe as many times as it was subscribed
		for i := 0; i < count; i++ {
			unsubscribedTopics = append(unsubscribedTopics, topic)
		}

		// delete the topic
		subscribedTopics.Delete(topic)

		return true
	})

	// delete the client
	s.subscribers.Delete(clientID)

	return removedTopics, unsubscribedTopics
}
