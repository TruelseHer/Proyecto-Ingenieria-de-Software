package platform

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"trotamundos/internal/domain"
)

const (
	pbkdf2Iterations = 15000
	pbkdf2SaltBytes  = 16
	pbkdf2KeyBytes   = 32
)

// NewID generates a collision-resistant identifier with a readable prefix.
func NewID(prefix string) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b))
}

// HashPassword hashes a plaintext password using PBKDF2-HMAC-SHA256 with a cryptographic random salt.
func HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", errors.New("la contraseña no puede estar vacía")
	}
	salt := make([]byte, pbkdf2SaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generar sal criptográfica: %w", err)
	}
	dk := pbkdf2SHA256([]byte(password), salt, pbkdf2Iterations, pbkdf2KeyBytes)
	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s",
		pbkdf2Iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(dk),
	), nil
}

// VerifyPassword validates a plaintext password against an encoded PBKDF2-HMAC-SHA256 hash in constant time.
func VerifyPassword(password, encodedHash string) bool {
	if password == "" || encodedHash == "" {
		return false
	}
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2_sha256" {
		return false
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations < 1000 || iterations > 1000000 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil || len(salt) == 0 {
		return false
	}
	expectedDK, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(expectedDK) == 0 {
		return false
	}
	actualDK := pbkdf2SHA256([]byte(password), salt, iterations, len(expectedDK))
	return subtle.ConstantTimeCompare(actualDK, expectedDK) == 1
}

// pbkdf2SHA256 implements RFC 2898 / PKCS #5 v2.0 PBKDF2 using HMAC-SHA256 from the standard library.
func pbkdf2SHA256(password, salt []byte, iter, keyLen int) []byte {
	prf := hmac.New(sha256.New, password)
	hashLen := prf.Size()
	numBlocks := (keyLen + hashLen - 1) / hashLen

	var buf [4]byte
	dk := make([]byte, 0, numBlocks*hashLen)
	u := make([]byte, hashLen)

	for block := 1; block <= numBlocks; block++ {
		prf.Reset()
		prf.Write(salt)
		binary.BigEndian.PutUint32(buf[:], uint32(block))
		prf.Write(buf[:])
		dk = prf.Sum(dk)
		t := dk[len(dk)-hashLen:]
		copy(u, t)

		for n := 2; n <= iter; n++ {
			prf.Reset()
			prf.Write(u)
			u = u[:0]
			u = prf.Sum(u)
			for x := range u {
				t[x] ^= u[x]
			}
		}
	}
	return dk[:keyLen]
}

// AnonymizeIdentifier produces a salted irreversible HMAC-SHA256 pseudonym for analytics records.
func AnonymizeIdentifier(rawID, salt string) string {
	mac := hmac.New(sha256.New, []byte(salt))
	mac.Write([]byte(rawID))
	sum := hex.EncodeToString(mac.Sum(nil))
	return "anon_" + sum[:16]
}

// Claims represents the signed session payload.
type Claims struct {
	UserID string `json:"sub"`
	Role   string `json:"role"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
}

// TokenClaims is an alias for Claims.
type TokenClaims = Claims

// TokenService issues, validates, and revokes signed session tokens backed by CacheClient and memory.
type TokenService struct {
	secret  []byte
	cache   *CacheClient
	mu      sync.RWMutex
	revoked map[string]time.Time
}

func NewTokenService(secret string, cache *CacheClient) *TokenService {
	if strings.TrimSpace(secret) == "" {
		secret = "trotamundos-v2-default-secret-key-2026"
	}
	return &TokenService{
		secret:  []byte(secret),
		cache:   cache,
		revoked: make(map[string]time.Time),
	}
}

func (s *TokenService) Issue(userID, role string, ttl time.Duration) (string, error) {
	if strings.TrimSpace(userID) == "" {
		return "", domain.ErrUnauthorized
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	now := time.Now().UTC()
	claims := Claims{
		UserID: userID,
		Role:   role,
		Iat:    now.Unix(),
		Exp:    now.Add(ttl).Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	sig := s.sign(encodedPayload)
	return encodedPayload + "." + sig, nil
}

func (s *TokenService) Validate(token string) (Claims, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return Claims{}, domain.ErrUnauthorized
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return Claims{}, domain.ErrUnauthorized
	}

	if s.isRevoked(token) {
		return Claims{}, domain.ErrUnauthorized
	}

	expectedSig := s.sign(parts[0])
	if subtle.ConstantTimeCompare([]byte(parts[1]), []byte(expectedSig)) != 1 {
		return Claims{}, domain.ErrUnauthorized
	}

	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Claims{}, domain.ErrUnauthorized
	}
	var claims Claims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return Claims{}, domain.ErrUnauthorized
	}
	if time.Now().UTC().Unix() > claims.Exp || claims.UserID == "" {
		return Claims{}, domain.ErrUnauthorized
	}
	return claims, nil
}

// Verify is an alias for Validate.
func (s *TokenService) Verify(token string) (Claims, error) {
	return s.Validate(token)
}

func (s *TokenService) Revoke(token string, ttl time.Duration) {
	token = strings.TrimSpace(token)
	if token == "" {
		return
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	s.mu.Lock()
	s.revoked[token] = time.Now().UTC().Add(ttl)
	s.mu.Unlock()

	if s.cache != nil {
		s.cache.Set("revoked:"+token, "1", ttl)
	}
}

func (s *TokenService) AnonymizeIdentifier(rawID, salt string) string {
	return AnonymizeIdentifier(rawID, salt)
}

func (s *TokenService) isRevoked(token string) bool {
	s.mu.RLock()
	exp, exists := s.revoked[token]
	s.mu.RUnlock()
	if exists {
		if time.Now().UTC().Before(exp) {
			return true
		}
		s.mu.Lock()
		delete(s.revoked, token)
		s.mu.Unlock()
	}
	if s.cache != nil {
		if _, found := s.cache.Get("revoked:" + token); found {
			return true
		}
	}
	return false
}

func (s *TokenService) sign(data string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

type rateBucket struct {
	count     int
	windowEnd time.Time
}

// RateLimiter enforces per-key request quotas within a one-minute sliding window.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]rateBucket
	limit   int
}

func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 240
	}
	return &RateLimiter{
		buckets: make(map[string]rateBucket),
		limit:   requestsPerMinute,
	}
}

func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	b, exists := r.buckets[key]
	if !exists || now.After(b.windowEnd) {
		r.buckets[key] = rateBucket{
			count:     1,
			windowEnd: now.Add(time.Minute),
		}
		return true
	}
	if b.count >= r.limit {
		return false
	}
	b.count++
	r.buckets[key] = b
	return true
}
