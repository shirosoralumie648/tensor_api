package relay

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSEManagerRegisterClient(t *testing.T) {
	manager := NewSSEManager()
	manager.Start()
	defer manager.Stop()

	t.Run("register_single_client", func(t *testing.T) {
		client, err := manager.RegisterClient("client-1", "user-1", "127.0.0.1")
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "client-1", client.ID)
		assert.Equal(t, "user-1", client.UserID)
		assert.Equal(t, 1, manager.GetActiveClientCount())
	})

	t.Run("register_multiple_clients", func(t *testing.T) {
		manager := NewSSEManager()
		manager.Start()
		defer manager.Stop()

		for i := 0; i < 100; i++ {
			clientID := fmt.Sprintf("client-%d", i)
			userID := fmt.Sprintf("user-%d", i%10)
			client, err := manager.RegisterClient(clientID, userID, "127.0.0.1")
			require.NoError(t, err)
			assert.NotNil(t, client)
		}

		assert.Equal(t, 100, manager.GetActiveClientCount())
	})

	t.Run("max_clients_reached", func(t *testing.T) {
		manager := NewSSEManager()
		manager.SetMaxClients(2)
		manager.Start()
		defer manager.Stop()

		// 注册两个客户端
		_, err1 := manager.RegisterClient("client-1", "user-1", "127.0.0.1")
		require.NoError(t, err1)

		_, err2 := manager.RegisterClient("client-2", "user-1", "127.0.0.1")
		require.NoError(t, err2)

		// 第三个应该失败
		_, err3 := manager.RegisterClient("client-3", "user-1", "127.0.0.1")
		require.Error(t, err3)
	})
}

func TestSSEManagerUnregisterClient(t *testing.T) {
	manager := NewSSEManager()
	manager.Start()
	defer manager.Stop()

	client, _ := manager.RegisterClient("client-1", "user-1", "127.0.0.1")
	assert.Equal(t, 1, manager.GetActiveClientCount())

	manager.UnregisterClient("client-1")
	assert.Equal(t, 0, manager.GetActiveClientCount())
	assert.True(t, client.IsClosed())
}

func TestSSEManagerBroadcast(t *testing.T) {
	manager := NewSSEManager()
	manager.Start()
	defer manager.Stop()

	// 注册 3 个客户端
	clients := make([]*SSEClient, 3)
	for i := 0; i < 3; i++ {
		clientID := fmt.Sprintf("client-%d", i)
		client, _ := manager.RegisterClient(clientID, "user-1", "127.0.0.1")
		clients[i] = client
	}

	// 广播消息
	msg := &SSEMessage{
		ID:    "msg-1",
		Event: "test",
		Data:  "hello",
	}
	manager.BroadcastMessage(msg)

	// 检查所有客户端都收到了消息
	for i, client := range clients {
		select {
		case received := <-client.MessageChan:
			assert.Equal(t, msg.ID, received.ID)
			assert.Equal(t, msg.Event, received.Event)
			assert.Equal(t, msg.Data, received.Data)
		case <-time.After(1 * time.Second):
			t.Fatalf("client %d did not receive message", i)
		}
	}
}

func TestSSEManagerSendToClient(t *testing.T) {
	manager := NewSSEManager()
	manager.Start()
	defer manager.Stop()

	// 注册客户端
	client1, _ := manager.RegisterClient("client-1", "user-1", "127.0.0.1")
	client2, _ := manager.RegisterClient("client-2", "user-1", "127.0.0.1")

	// 发送消息给特定客户端
	msg := &SSEMessage{
		ID:    "msg-1",
		Event: "test",
		Data:  "hello",
	}

	err := manager.SendMessageToClient("client-1", msg)
	require.NoError(t, err)

	// client-1 应该收到消息
	select {
	case received := <-client1.MessageChan:
		assert.Equal(t, msg.Data, received.Data)
	case <-time.After(1 * time.Second):
		t.Fatal("client-1 did not receive message")
	}

	// client-2 不应该收到消息（通道应该是空的）
	select {
	case <-client2.MessageChan:
		t.Fatal("client-2 should not receive message")
	default:
		// 预期的结果
	}

	// 发送消息给不存在的客户端
	err = manager.SendMessageToClient("client-non-existent", msg)
	require.Error(t, err)
}

func TestSSEClientClose(t *testing.T) {
	client := &SSEClient{
		ID:          "client-1",
		MessageChan: make(chan *SSEMessage, 10),
		CloseChan:   make(chan struct{}),
	}

	assert.False(t, client.IsClosed())

	client.Close()
	assert.True(t, client.IsClosed())

	// 再次关闭应该不会panic
	client.Close()
	assert.True(t, client.IsClosed())
}

func TestSSEClientUpdateActivity(t *testing.T) {
	client := &SSEClient{
		ID:             "client-1",
		MessageChan:    make(chan *SSEMessage, 10),
		CloseChan:      make(chan struct{}),
		CreatedAt:      time.Now().Add(-1 * time.Hour),
		LastActivityAt: time.Now().Add(-1 * time.Hour),
	}

	oldTime := client.LastActivityAt
	time.Sleep(100 * time.Millisecond)

	client.UpdateActivity()

	assert.True(t, client.LastActivityAt.After(oldTime))
}

func TestSSEManagerHeartbeat(t *testing.T) {
	manager := NewSSEManager()
	manager.SetHeartbeatInterval(100 * time.Millisecond)
	manager.Start()
	defer manager.Stop()

	client, _ := manager.RegisterClient("client-1", "user-1", "127.0.0.1")

	// 等待心跳消息
	var heartbeatCount int
	timeout := time.After(500 * time.Millisecond)

	for {
		select {
		case msg := <-client.MessageChan:
			if msg.Comment == "heartbeat" {
				heartbeatCount++
			}
		case <-timeout:
			goto done
		}
	}

done:
	// 应该收到至少 3-4 个心跳
	assert.Greater(t, heartbeatCount, 2)
}

func TestSSEManagerStatistics(t *testing.T) {
	manager := NewSSEManager()
	manager.Start()
	defer manager.Stop()

	// 注册客户端
	for i := 0; i < 5; i++ {
		clientID := fmt.Sprintf("client-%d", i)
		manager.RegisterClient(clientID, "user-1", "127.0.0.1")
	}

	// 广播消息
	msg := &SSEMessage{
		ID:   "msg-1",
		Data: "test data",
	}
	manager.BroadcastMessage(msg)

	stats := manager.GetStatistics()
	assert.Equal(t, int32(5), stats["active_connections"])
	assert.Equal(t, int64(5), stats["total_messages"])
}

func TestSSEManagerConcurrency(t *testing.T) {
	manager := NewSSEManager()
	manager.Start()
	defer manager.Stop()

	numGoroutines := 100
	var wg sync.WaitGroup
	var errors int32

	// 并发注册客户端
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			clientID := fmt.Sprintf("client-%d", i)
			_, err := manager.RegisterClient(clientID, "user-1", "127.0.0.1")
			if err != nil {
				atomic.AddInt32(&errors, 1)
			}
		}(i)
	}

	wg.Wait()

	assert.Equal(t, int32(0), errors)
	assert.Equal(t, numGoroutines, manager.GetActiveClientCount())

	// 并发广播消息
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			msg := &SSEMessage{
				Data: fmt.Sprintf("message-%d", i),
			}
			manager.BroadcastMessage(msg)
		}(i)
	}

	wg.Wait()

	stats := manager.GetStatistics()
	assert.Equal(t, int64(10), stats["total_messages"])
}

func BenchmarkSSEManagerBroadcast(b *testing.B) {
	manager := NewSSEManager()
	manager.Start()
	defer manager.Stop()

	// 注册 1000 个客户端
	for i := 0; i < 1000; i++ {
		clientID := fmt.Sprintf("client-%d", i)
		manager.RegisterClient(clientID, "user-1", "127.0.0.1")
	}

	msg := &SSEMessage{
		Data: "benchmark test",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.BroadcastMessage(msg)
	}
}

func BenchmarkSSEManagerRegister(b *testing.B) {
	manager := NewSSEManager()
	manager.Start()
	defer manager.Stop()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		clientID := fmt.Sprintf("bench-client-%d", i)
		manager.RegisterClient(clientID, "user-1", "127.0.0.1")
	}
}

