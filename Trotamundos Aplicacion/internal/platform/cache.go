package platform

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type cacheEntry struct {
	value     string
	expiresAt time.Time
}

// CacheClient provides Redis 7 RESP over TCP caching with automatic thread-safe in-memory fallback.
type CacheClient struct {
	mu        sync.RWMutex
	entries   map[string]cacheEntry
	redisAddr string
}

// Cache is an alias for CacheClient.
type Cache = CacheClient

func NewCacheClient(redisAddr string) *CacheClient {
	return &CacheClient{
		entries:   make(map[string]cacheEntry),
		redisAddr: strings.TrimSpace(redisAddr),
	}
}

func (c *CacheClient) Get(key string) (string, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()

	if ok {
		if time.Now().Before(entry.expiresAt) {
			return entry.value, true
		}
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
	}

	if c.redisAddr != "" {
		if raw, found := c.redisGet(key); found {
			c.mu.Lock()
			c.entries[key] = cacheEntry{
				value:     raw,
				expiresAt: time.Now().Add(2 * time.Minute),
			}
			c.mu.Unlock()
			return raw, true
		}
	}
	return "", false
}

func (c *CacheClient) Set(key, val string, ttl time.Duration) {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	c.mu.Lock()
	c.entries[key] = cacheEntry{
		value:     val,
		expiresAt: time.Now().Add(ttl),
	}
	c.mu.Unlock()

	if c.redisAddr != "" {
		c.redisSet(key, val, ttl)
	}
}

func (c *CacheClient) GetJSON(key string, dst any) bool {
	raw, ok := c.Get(key)
	if !ok {
		return false
	}
	return json.Unmarshal([]byte(raw), dst) == nil
}

func (c *CacheClient) SetJSON(key string, val any, ttl time.Duration) {
	raw, err := json.Marshal(val)
	if err != nil {
		return
	}
	c.Set(key, string(raw), ttl)
}

func (c *CacheClient) Delete(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
	if c.redisAddr != "" {
		c.redisDel(key)
	}
}

func (c *CacheClient) DeletePrefix(prefix string) {
	c.mu.Lock()
	var keys []string
	for k := range c.entries {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
			delete(c.entries, k)
		}
	}
	c.mu.Unlock()
	if c.redisAddr != "" {
		for _, k := range keys {
			c.redisDel(k)
		}
	}
}

// InvalidatePrefix is an alias for DeletePrefix.
func (c *CacheClient) InvalidatePrefix(prefix string) {
	c.DeletePrefix(prefix)
}

func (c *CacheClient) Mode() string {
	if c.redisAddr != "" && c.Ping() == nil {
		return "redis"
	}
	return "memory"
}

func (c *CacheClient) Ping() error {
	if c.redisAddr == "" {
		return nil
	}
	conn, err := net.DialTimeout("tcp", c.redisAddr, 600*time.Millisecond)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(600 * time.Millisecond))
	if _, err := fmt.Fprint(conn, "*1\r\n$4\r\nPING\r\n"); err != nil {
		return err
	}
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if !strings.HasPrefix(line, "+PONG") {
		return fmt.Errorf("unexpected redis ping response: %s", strings.TrimSpace(line))
	}
	return nil
}

func (c *CacheClient) redisSet(key, val string, ttl time.Duration) {
	conn, err := net.DialTimeout("tcp", c.redisAddr, 250*time.Millisecond)
	if err != nil {
		return
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(250 * time.Millisecond))
	secs := int(ttl.Seconds())
	if secs <= 0 {
		secs = 60
	}
	secsStr := strconv.Itoa(secs)
	cmd := fmt.Sprintf("*4\r\n$5\r\nSETEX\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
		len(key), key, len(secsStr), secsStr, len(val), val)
	_, _ = conn.Write([]byte(cmd))
}

func (c *CacheClient) redisGet(key string) (string, bool) {
	conn, err := net.DialTimeout("tcp", c.redisAddr, 250*time.Millisecond)
	if err != nil {
		return "", false
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(250 * time.Millisecond))
	cmd := fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(key), key)
	if _, err := conn.Write([]byte(cmd)); err != nil {
		return "", false
	}
	reader := bufio.NewReader(conn)
	header, err := reader.ReadString('\n')
	if err != nil || !strings.HasPrefix(header, "$") {
		return "", false
	}
	size, err := strconv.Atoi(strings.TrimSpace(header[1:]))
	if err != nil || size <= 0 {
		return "", false
	}
	buf := make([]byte, size+2)
	read := 0
	for read < len(buf) {
		n, err := reader.Read(buf[read:])
		if err != nil {
			return "", false
		}
		read += n
	}
	return string(buf[:size]), true
}

func (c *CacheClient) redisDel(key string) {
	conn, err := net.DialTimeout("tcp", c.redisAddr, 250*time.Millisecond)
	if err != nil {
		return
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(250 * time.Millisecond))
	cmd := fmt.Sprintf("*2\r\n$3\r\nDEL\r\n$%d\r\n%s\r\n", len(key), key)
	_, _ = conn.Write([]byte(cmd))
}
