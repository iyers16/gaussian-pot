package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/iyers16/gaussian-pot/backend/internal/model"
	"github.com/iyers16/gaussian-pot/backend/internal/repository"
)

const maxSessions = 9

// SessionStore holds active sessions in memory.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*model.Session // token → session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]*model.Session)}
}

func (s *SessionStore) Add(sess *model.Session) {
	s.mu.Lock()
	s.sessions[sess.Token] = sess
	s.mu.Unlock()
}

func (s *SessionStore) Remove(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func (s *SessionStore) Get(token string) (*model.Session, bool) {
	s.mu.RLock()
	sess, ok := s.sessions[token]
	s.mu.RUnlock()
	return sess, ok
}

func (s *SessionStore) Count() int {
	s.mu.RLock()
	n := len(s.sessions)
	s.mu.RUnlock()
	return n
}

func (s *SessionStore) ActiveUsers() []gin.H {
	s.mu.RLock()
	defer s.mu.RUnlock()
	seen := map[string]bool{}
	var users []gin.H
	for _, sess := range s.sessions {
		if seen[sess.Username] {
			continue
		}
		seen[sess.Username] = true
		users = append(users, gin.H{"username": sess.Username, "role": sess.Role})
	}
	return users
}

// AuthHandler handles login and logout.
type AuthHandler struct {
	userRepo *repository.UserRepository
	sessions *SessionStore
	hub      *Hub
}

func NewAuthHandler(userRepo *repository.UserRepository, sessions *SessionStore, hub *Hub) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, sessions: sessions, hub: hub}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	hostPassword := os.Getenv("HOST_PASSWORD")
	if hostPassword == "" {
		hostPassword = "host123"
	}

	// Determine role.
	var role model.Role
	if req.Username == "host" {
		if req.Password != hostPassword {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid host credentials"})
			return
		}
		role = model.RoleHost
	} else {
		role = model.RolePlayer
	}

	// Enforce session cap.
	if h.sessions.Count() >= maxSessions {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many players at the moment — try again later"})
		return
	}

	// Find or create user.
	user, err := h.userRepo.GetByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if user == nil {
		user, err = h.userRepo.Create(req.Username, role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
	}

	// Generate session token.
	token, err := generateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	sess := &model.Session{
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	}
	h.sessions.Add(sess)

	// Notify all clients.
	h.hub.Broadcast("session_update", gin.H{"active_users": h.sessions.ActiveUsers()})

	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"username": user.Username,
		"role":     user.Role,
		"credits":  user.Credits,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token := tokenFromContext(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	h.sessions.Remove(token)
	h.hub.Broadcast("session_update", gin.H{"active_users": h.sessions.ActiveUsers()})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// AuthMiddleware validates the Bearer token and injects the session into the context.
func AuthMiddleware(sessions *SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := tokenFromContext(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		sess, ok := sessions.Get(token)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		c.Set("session", sess)
		c.Next()
	}
}

// HostMiddleware requires the session to have host role.
func HostMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sess := sessionFromContext(c)
		if sess == nil || sess.Role != model.RoleHost {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "host only"})
			return
		}
		c.Next()
	}
}

func sessionFromContext(c *gin.Context) *model.Session {
	v, exists := c.Get("session")
	if !exists {
		return nil
	}
	sess, _ := v.(*model.Session)
	return sess
}

func tokenFromContext(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
