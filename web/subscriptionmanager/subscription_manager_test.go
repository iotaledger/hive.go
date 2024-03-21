//nolint:golint,revive,stylecheck // we don't care about these linters in test cases
package subscriptionmanager_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/web/subscriptionmanager"
)

const (
	clientID_1 = "client1"
	clientID_2 = "client2"

	topic_1 = "topic1"
	topic_2 = "topic2"
	topic_3 = "topic3"
	topic_4 = "topic4"
	topic_5 = "topic5"
	topic_6 = "topic6"
)

func TestSubscriptionManager_ConnectWithNoTopics(t *testing.T) {
	manager := subscriptionmanager.New(
		subscriptionmanager.WithMaxTopicSubscriptionsPerClient[string, string](1000),
		subscriptionmanager.WithCleanupThresholdRatio[string, string](0.0),
		subscriptionmanager.WithCleanupThresholdCount[string, string](0))

	require.Equal(t, 0, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())

	manager.Connect(clientID_1)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())

	manager.Disconnect(clientID_1)
	require.Equal(t, 0, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())
}

func TestSubscriptionManager_ConnectWithSameID(t *testing.T) {
	manager := subscriptionmanager.New(
		subscriptionmanager.WithMaxTopicSubscriptionsPerClient[string, string](1000),
		subscriptionmanager.WithCleanupThresholdRatio[string, string](0.0),
		subscriptionmanager.WithCleanupThresholdCount[string, string](0))

	manager.Connect(clientID_1)
	manager.Subscribe(clientID_1, topic_1)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 1, manager.TopicsSize())

	manager.Connect(clientID_1)
	manager.Subscribe(clientID_1, topic_2)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 1, manager.TopicsSize())

	manager.Disconnect(clientID_1)
	require.Equal(t, 0, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())
}

func TestSubscriptionManager_SubscribeWithoutConnect(t *testing.T) {
	manager := subscriptionmanager.New(
		subscriptionmanager.WithMaxTopicSubscriptionsPerClient[string, string](1000),
		subscriptionmanager.WithCleanupThresholdRatio[string, string](0.0),
		subscriptionmanager.WithCleanupThresholdCount[string, string](0))

	manager.Subscribe(clientID_1, topic_1)
	require.Equal(t, 0, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())
}

func TestSubscriptionManager_SubscribeWithSameTopic(t *testing.T) {
	manager := subscriptionmanager.New(
		subscriptionmanager.WithMaxTopicSubscriptionsPerClient[string, string](1000),
		subscriptionmanager.WithCleanupThresholdRatio[string, string](0.0),
		subscriptionmanager.WithCleanupThresholdCount[string, string](0))

	manager.Connect(clientID_1)
	manager.Subscribe(clientID_1, topic_1)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 1, manager.TopicsSize())

	manager.Subscribe(clientID_1, topic_1)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 1, manager.TopicsSize())
}

func TestSubscriptionManager_UnsubscribeWithoutConnect(t *testing.T) {
	manager := subscriptionmanager.New(
		subscriptionmanager.WithMaxTopicSubscriptionsPerClient[string, string](1000),
		subscriptionmanager.WithCleanupThresholdRatio[string, string](0.0),
		subscriptionmanager.WithCleanupThresholdCount[string, string](0))

	manager.Unsubscribe(clientID_1, topic_1)
	require.Equal(t, 0, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())
}

func TestSubscriptionManager_UnsubscribeWithSameTopic(t *testing.T) {
	manager := subscriptionmanager.New(
		subscriptionmanager.WithMaxTopicSubscriptionsPerClient[string, string](1000),
		subscriptionmanager.WithCleanupThresholdRatio[string, string](0.0),
		subscriptionmanager.WithCleanupThresholdCount[string, string](0))

	manager.Connect(clientID_1)
	manager.Subscribe(clientID_1, topic_1)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 1, manager.TopicsSize())

	manager.Unsubscribe(clientID_1, topic_1)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())

	manager.Unsubscribe(clientID_1, topic_1)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())
}

func TestSubscriptionManager_Subscribers(t *testing.T) {
	manager := subscriptionmanager.New(
		subscriptionmanager.WithMaxTopicSubscriptionsPerClient[string, string](1000),
		subscriptionmanager.WithCleanupThresholdRatio[string, string](0.0),
		subscriptionmanager.WithCleanupThresholdCount[string, string](0))

	manager.Connect(clientID_1)
	manager.Connect(clientID_1)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())

	manager.Connect(clientID_2)
	manager.Subscribe(clientID_2, topic_1)
	require.Equal(t, 2, manager.SubscribersSize())
	require.Equal(t, 1, manager.TopicsSize())

	manager.Subscribe(clientID_2, topic_2)
	require.Equal(t, 2, manager.SubscribersSize())
	require.Equal(t, 2, manager.TopicsSize())

	manager.Disconnect(clientID_2)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())
}

func TestSubscriptionManager_ClientCleanup(t *testing.T) {

	manager := subscriptionmanager.New(
		subscriptionmanager.WithMaxTopicSubscriptionsPerClient[string, string](1000),
		subscriptionmanager.WithCleanupThresholdRatio[string, string](0.0),
		subscriptionmanager.WithCleanupThresholdCount[string, string](0))

	subscribe_client_1 := 0
	unsubscribe_client_1 := 0

	manager.Events().TopicSubscribed.Hook(func(event *subscriptionmanager.ClientTopicEvent[string, string]) {
		if event.ClientID == clientID_1 {
			subscribe_client_1++
		}
	})

	manager.Events().TopicUnsubscribed.Hook(func(event *subscriptionmanager.ClientTopicEvent[string, string]) {
		if event.ClientID == clientID_1 {
			unsubscribe_client_1++
		}
	})

	manager.Connect(clientID_1)
	manager.Subscribe(clientID_1, topic_1)
	manager.Subscribe(clientID_1, topic_2)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 2, manager.TopicsSize())
	require.Equal(t, 2, subscribe_client_1)
	require.Equal(t, 0, unsubscribe_client_1)

	manager.Connect(clientID_1)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())
	require.Equal(t, 2, subscribe_client_1)
	require.Equal(t, 2, unsubscribe_client_1)

	manager.Subscribe(clientID_1, topic_1)
	manager.Subscribe(clientID_1, topic_2)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 2, manager.TopicsSize())
	require.Equal(t, 4, subscribe_client_1)
	require.Equal(t, 2, unsubscribe_client_1)

	manager.Disconnect(clientID_1)
	require.Equal(t, 0, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())
	require.Equal(t, 4, subscribe_client_1)
	require.Equal(t, 4, unsubscribe_client_1)
}

func TestSubscriptionManager_MaxTopicSubscriptionsPerClient(t *testing.T) {

	manager := subscriptionmanager.New(
		subscriptionmanager.WithMaxTopicSubscriptionsPerClient[string, string](5),
		subscriptionmanager.WithCleanupThresholdRatio[string, string](0.0),
		subscriptionmanager.WithCleanupThresholdCount[string, string](0))

	clientDropped := false
	manager.Events().DropClient.Hook(func(event *subscriptionmanager.DropClientEvent[string]) {
		clientDropped = true
	})

	require.Equal(t, 0, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())
	require.Equal(t, false, clientDropped)

	manager.Connect(clientID_1)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())
	require.Equal(t, false, clientDropped)

	manager.Subscribe(clientID_1, topic_1)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 1, manager.TopicsSize())
	require.Equal(t, false, clientDropped)

	manager.Subscribe(clientID_1, topic_2)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 2, manager.TopicsSize())
	require.Equal(t, false, clientDropped)

	manager.Subscribe(clientID_1, topic_3)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 3, manager.TopicsSize())
	require.Equal(t, false, clientDropped)

	manager.Subscribe(clientID_1, topic_4)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 4, manager.TopicsSize())
	require.Equal(t, false, clientDropped)

	manager.Subscribe(clientID_1, topic_5)
	require.Equal(t, true, clientDropped)
	require.Equal(t, 0, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())

	manager.Subscribe(clientID_1, topic_6)
	require.Equal(t, 0, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())
	require.Equal(t, true, clientDropped)
}

func TestSubscriptionManager_TopicEvents(t *testing.T) {
	manager := subscriptionmanager.New(
		subscriptionmanager.WithMaxTopicSubscriptionsPerClient[string, string](1000),
		subscriptionmanager.WithCleanupThresholdRatio[string, string](0.0),
		subscriptionmanager.WithCleanupThresholdCount[string, string](0))

	topicAdded := false
	manager.Events().TopicAdded.Hook(func(event *subscriptionmanager.TopicEvent[string]) {
		if event.Topic == topic_1 {
			topicAdded = true
		}
	})

	topicRemoved := false
	manager.Events().TopicRemoved.Hook(func(event *subscriptionmanager.TopicEvent[string]) {
		if event.Topic == topic_1 {
			topicRemoved = true
		}
	})

	require.Equal(t, 0, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())

	manager.Connect(clientID_1)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())

	manager.Connect(clientID_2)
	require.Equal(t, 2, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())

	manager.Subscribe(clientID_1, topic_1)
	require.Equal(t, 2, manager.SubscribersSize())
	require.Equal(t, 1, manager.TopicsSize())
	require.Equal(t, true, topicAdded)

	manager.Subscribe(clientID_2, topic_1)
	require.Equal(t, 2, manager.SubscribersSize())
	require.Equal(t, 1, manager.TopicsSize())

	manager.Disconnect(clientID_1)
	manager.Disconnect(clientID_2)

	require.Equal(t, 0, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())
	require.Equal(t, true, topicRemoved)

	topicAdded = false

	manager.Connect(clientID_1)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 0, manager.TopicsSize())

	manager.Subscribe(clientID_1, topic_1)
	require.Equal(t, 1, manager.SubscribersSize())
	require.Equal(t, 1, manager.TopicsSize())
	require.Equal(t, true, topicAdded)
}
