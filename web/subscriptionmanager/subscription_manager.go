// SubscriptionManager keeps track of subscribed topics of clients.
// This allows to get notified when a client connects or disconnects
// or a topic is subscribed or unsubscribed.
package subscriptionmanager

import (
	"sync"

	"github.com/iotaledger/hive.go/constraints"
	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/runtime/event"
	"github.com/iotaledger/hive.go/runtime/options"
)

var (
	ErrMaxTopicSubscriptionsPerClientReached = ierrors.New("maximum amount of topic subscriptions per client reached")
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
	ClientConnected *event.Event1[*ClientEvent[C]]
	// ClientDisconnected event is triggered when a client disconnected.
	ClientDisconnected *event.Event1[*ClientEvent[C]]
	// TopicSubscribed event is triggered when a client subscribed to a topic.
	TopicSubscribed *event.Event1[*ClientTopicEvent[C, T]]
	// TopicUnsubscribed event is triggered when a client unsubscribed from a topic.
	TopicUnsubscribed *event.Event1[*ClientTopicEvent[C, T]]
	// TopicAdded event is triggered when a topic is subscribed for the first time by any client.
	TopicAdded *event.Event1[*TopicEvent[T]]
	// TopicRemoved event is triggered when a topic is not subscribed anymore by any client.
	TopicRemoved *event.Event1[*TopicEvent[T]]
	// DropClient event is triggered when a client should be dropped.
	DropClient *event.Event1[*DropClientEvent[C]]
}

func newEvents[C ClientID, T Topic]() *Events[C, T] {
	return &Events[C, T]{
		ClientConnected:    event.New1[*ClientEvent[C]](),
		ClientDisconnected: event.New1[*ClientEvent[C]](),
		TopicSubscribed:    event.New1[*ClientTopicEvent[C, T]](),
		TopicUnsubscribed:  event.New1[*ClientTopicEvent[C, T]](),
		TopicAdded:         event.New1[*TopicEvent[T]](),
		TopicRemoved:       event.New1[*TopicEvent[T]](),
		DropClient:         event.New1[*DropClientEvent[C]](),
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

// WithCleanupThresholdCount defines the count of
// deletions that triggers shrinking of the map.
func WithCleanupThresholdCount[C ClientID, T Topic](cleanupThresholdCount int) options.Option[SubscriptionManager[C, T]] {
	return func(s *SubscriptionManager[C, T]) {
		s.cleanupThresholdCount = cleanupThresholdCount
	}
}

// WithCleanupThresholdRatio defines the ratio between the amount
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

// Connect connects the client.
func (s *SubscriptionManager[C, T]) Connect(clientID C) {
	wasConnected := false
	var removedTopics, unsubscribedTopics []T

	// inline function used to release the lock before firing the event
	func() {
		s.Lock()
		defer s.Unlock()

		// in case the client already exists, we cleanup old subscriptions.
		wasConnected, removedTopics, unsubscribedTopics = s.cleanupClientWithoutLocking(clientID)

		// create a new map for the client
		s.subscribers.Set(clientID, shrinkingmap.New[T, int](
			shrinkingmap.WithShrinkingThresholdRatio(s.cleanupThresholdRatio),
			shrinkingmap.WithShrinkingThresholdCount(s.cleanupThresholdCount),
		))
	}()

	if wasConnected {
		for _, topic := range removedTopics {
			s.events.TopicRemoved.Trigger(&TopicEvent[T]{Topic: topic})
		}
		for _, topic := range unsubscribedTopics {
			s.events.TopicUnsubscribed.Trigger(&ClientTopicEvent[C, T]{ClientID: clientID, Topic: topic})
		}
		s.events.ClientDisconnected.Trigger(&ClientEvent[C]{ClientID: clientID})
	}

	s.events.ClientConnected.Trigger(&ClientEvent[C]{ClientID: clientID})
}

// Disconnect disconnects the client.
// Returns true if the client was connected.
func (s *SubscriptionManager[C, T]) Disconnect(clientID C) bool {
	// cleanup the client
	if wasConnected, removedTopics, unsubscribedTopics := s.cleanupClient(clientID); wasConnected {
		for _, topic := range removedTopics {
			s.events.TopicRemoved.Trigger(&TopicEvent[T]{Topic: topic})
		}
		for _, topic := range unsubscribedTopics {
			s.events.TopicUnsubscribed.Trigger(&ClientTopicEvent[C, T]{ClientID: clientID, Topic: topic})
		}
		s.events.ClientDisconnected.Trigger(&ClientEvent[C]{ClientID: clientID})

		return true
	}

	return false
}

// Subscribe subscribes the client to the topic.
// Returns true if the client successfully subscribed to the topic.
func (s *SubscriptionManager[C, T]) Subscribe(clientID C, topic T) bool {
	clientDropped := false
	var removedTopics, unsubscribedTopics []T
	topicAdded := false

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
				_, removedTopics, unsubscribedTopics = s.cleanupClientWithoutLocking(clientID)
				clientDropped = true

				// the client gets dropped
				// do not modify the global map
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

	if clientDropped {
		for _, topic := range removedTopics {
			s.events.TopicRemoved.Trigger(&TopicEvent[T]{Topic: topic})
		}
		for _, topic := range unsubscribedTopics {
			s.events.TopicUnsubscribed.Trigger(&ClientTopicEvent[C, T]{ClientID: clientID, Topic: topic})
		}
		// notify the caller to drop the client
		s.events.DropClient.Trigger(&DropClientEvent[C]{ClientID: clientID, Reason: ErrMaxTopicSubscriptionsPerClientReached})

		s.events.ClientDisconnected.Trigger(&ClientEvent[C]{ClientID: clientID})

		// do not fire the subscribed events
		return false
	}

	if topicAdded {
		s.events.TopicAdded.Trigger(&TopicEvent[T]{Topic: topic})
	}

	s.events.TopicSubscribed.Trigger(&ClientTopicEvent[C, T]{ClientID: clientID, Topic: topic})

	return true
}

// Unsubscribe unsubscribes the client from the topic.
// Returns true if the client successfully unsubscribed from the topic.
func (s *SubscriptionManager[C, T]) Unsubscribe(clientID C, topic T) bool {
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

		// check if the client was subscribed to the topic
		count, has := subscribedTopics.Get(topic)
		if !has {
			return
		}

		if count <= 1 {
			// delete the topic
			subscribedTopics.Delete(topic)
		} else {
			subscribedTopics.Set(topic, count-1)
		}

		topicUnsubscribed = true

		// check global topics map
		count, has = s.topics.Get(topic)
		if !has {
			return
		}

		if count <= 1 {
			// delete the topic
			s.topics.Delete(topic)
			topicRemoved = true
		} else {
			s.topics.Set(topic, count-1)
		}
	}()

	if !topicUnsubscribed {
		// do not fire the unsubscribed events
		return false
	}

	if topicRemoved {
		s.events.TopicRemoved.Trigger(&TopicEvent[T]{Topic: topic})
	}

	s.events.TopicUnsubscribed.Trigger(&ClientTopicEvent[C, T]{ClientID: clientID, Topic: topic})

	return true
}

func (s *SubscriptionManager[C, T]) TopicHasSubscribers(topic T) bool {
	s.RLock()
	defer s.RUnlock()

	return s.topics.Has(topic)
}

func (s *SubscriptionManager[C, T]) ClientSubscribedToTopic(clientID C, topic T) bool {
	s.RLock()
	defer s.RUnlock()

	// check if the client is connected
	subscribedTopics, exists := s.subscribers.Get(clientID)
	if !exists {
		return false
	}

	return subscribedTopics.Has(topic)
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
	//nolint:revive // better be explicit here
	s.subscribers.ForEach(func(clientID C, topics *shrinkingmap.ShrinkingMap[T, int]) bool {
		count += topics.Size()

		return true
	})

	return count
}

// cleanupClientWithoutLocking removes all subscriptions and the client itself.
func (s *SubscriptionManager[C, T]) cleanupClientWithoutLocking(clientID C) (bool, []T, []T) {
	removedTopics := make([]T, 0)
	unsubscribedTopics := make([]T, 0)

	// check if the client exists
	subscribedTopics, has := s.subscribers.Get(clientID)
	if !has {
		return false, removedTopics, unsubscribedTopics
	}

	// loop over all topics and delete them
	subscribedTopics.ForEach(func(topic T, count int) bool {
		// global topics map
		topicsCount, has := s.topics.Get(topic)
		if has {
			if topicsCount-count <= 0 {
				// delete the topic
				s.topics.Delete(topic)
				removedTopics = append(removedTopics, topic)
			} else {
				s.topics.Set(topic, topicsCount-count)
			}
		}

		// call the topic unsubscribe as many times as it was subscribed
		for range count {
			unsubscribedTopics = append(unsubscribedTopics, topic)
		}

		// delete the topic
		subscribedTopics.Delete(topic)

		return true
	})

	// delete the client
	s.subscribers.Delete(clientID)

	return true, removedTopics, unsubscribedTopics
}

// cleanupClient removes all subscriptions and the client itself.
func (s *SubscriptionManager[C, T]) cleanupClient(clientID C) (bool, []T, []T) {
	s.Lock()
	defer s.Unlock()

	return s.cleanupClientWithoutLocking(clientID)
}
