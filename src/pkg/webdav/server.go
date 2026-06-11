package webdav

import (
	"fmt"
	"myobj/src/config"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/logger"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/webdav"
)

// rateLimitEntry 速率限制条目
type rateLimitEntry struct {
	count   int
	resetAt time.Time
}

// ipRateLimiter 基于 IP 的速率限制器
type ipRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*rateLimitEntry
	limit    int
	window   time.Duration
}

func newIPRateLimiter(limit int, window time.Duration) *ipRateLimiter {
	rl := &ipRateLimiter{
		visitors: make(map[string]*rateLimitEntry),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (l *ipRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	entry, exists := l.visitors[ip]

	if !exists || now.After(entry.resetAt) {
		l.visitors[ip] = &rateLimitEntry{
			count:   1,
			resetAt: now.Add(l.window),
		}
		return true
	}

	if entry.count >= l.limit {
		return false
	}

	entry.count++
	return true
}

func (l *ipRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for ip, entry := range l.visitors {
			if now.After(entry.resetAt) {
				delete(l.visitors, ip)
			}
		}
		l.mu.Unlock()
	}
}

// extractIP 从 RemoteAddr 中提取 IP 地址
func extractIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

// Server WebDAV 服务器
type Server struct {
	auth        *Authenticator
	factory     *impl.RepositoryFactory
	rateLimiter *ipRateLimiter
	redis       *redis.Client
	lockSystem  *RedisLockSystem
}

// NewServer 创建 WebDAV 服务器实例
func NewServer(factory *impl.RepositoryFactory) *Server {
	auth := NewAuthenticator(
		factory.ApiKey(),
		factory.User(),
		factory.Power(),
		factory.SysConfig(),
	)

	s := &Server{
		auth:        auth,
		factory:     factory,
		rateLimiter: newIPRateLimiter(60, 1*time.Minute), // 每分钟最多 60 次认证请求
	}

	// 初始化 Redis 分布式锁
	redisClient, err := NewRedisClientFromConfig(&config.CONFIG.Cache)
	if err != nil {
		logger.LOG.Warn("Redis 连接失败，WebDAV 将使用内存锁（仅单实例可用）", "error", err)
	} else {
		s.redis = redisClient
		s.lockSystem = NewRedisLockSystem(redisClient)
		logger.LOG.Info("WebDAV 使用 Redis 分布式锁")
	}

	return s
}

// ServeHTTP 处理 WebDAV 请求
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 0. 限速检查
	clientIP := extractIP(r.RemoteAddr)
	if !s.rateLimiter.allow(clientIP) {
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

	// 1. 认证
	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="MyObj WebDAV"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		logger.LOG.Warn("WebDAV 请求缺少认证信息", "ip", r.RemoteAddr, "path", r.URL.Path)
		return
	}

	// 2. 验证用户
	user, err := s.auth.Authenticate(r.Context(), username, password)
	if err != nil {
		logger.LOG.Warn("WebDAV 认证失败",
			"username", username,
			"ip", r.RemoteAddr,
			"error", err,
		)
		w.Header().Set("WWW-Authenticate", `Basic realm="MyObj WebDAV"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 3. 检查 WebDAV 访问权限
	hasPermission, err := s.auth.CheckPermission(r.Context(), user.ID, user.GroupID, "webdav:access")
	if err != nil || !hasPermission {
		logger.LOG.Warn("WebDAV 权限不足",
			"user_id", user.ID,
			"username", username,
			"group_id", user.GroupID,
		)
		http.Error(w, "无权限访问 WebDAV", http.StatusForbidden)
		return
	}

	// 4. 创建用户专属的文件系统
	fs := NewMyObjFileSystem(user, s.factory)

	// 5. 创建 WebDAV Handler
	lockSys := s.getLockSystem()
	handler := &webdav.Handler{
		FileSystem: fs,
		LockSystem: lockSys,
		Logger: func(r *http.Request, err error) {
			if err != nil {
				logger.LOG.Error("WebDAV 操作错误",
					"user_id", user.ID,
					"username", username,
					"method", r.Method,
					"path", r.URL.Path,
					"error", err,
				)
			} else {
				logger.LOG.Debug("WebDAV 操作",
					"user_id", user.ID,
					"username", username,
					"method", r.Method,
					"path", r.URL.Path,
				)
			}
		},
	}

	// 6. 处理请求
	handler.ServeHTTP(w, r)
}

// Stop 关闭 WebDAV 服务器的 Redis 连接
func (s *Server) Stop() {
	if s.lockSystem != nil {
		if err := s.lockSystem.Close(); err != nil {
			logger.LOG.Error("关闭 WebDAV Redis 连接失败", "error", err)
		}
	}
	if s.redis != nil {
		if err := s.redis.Close(); err != nil {
			logger.LOG.Error("关闭 Redis 客户端失败", "error", err)
		}
	}
}

// Start 启动 WebDAV 服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d",
		config.CONFIG.WebDAV.Host,
		config.CONFIG.WebDAV.Port,
	)

	logger.LOG.Info("========== WebDAV 服务器启动 ==========")
	logger.LOG.Info("WebDAV 服务器配置",
		"address", addr,
		"prefix", config.CONFIG.WebDAV.Prefix,
	)

	// 记录锁系统类型
	if s.lockSystem != nil {
		logger.LOG.Info("WebDAV 锁系统", "type", "Redis 分布式锁")
	} else {
		logger.LOG.Info("WebDAV 锁系统", "type", "内存锁（单实例模式）")
	}

	// 注册处理器
	prefix := config.CONFIG.WebDAV.Prefix
	if prefix == "" {
		prefix = "/dav"
	}

	http.Handle(prefix+"/", http.StripPrefix(prefix, s))

	// 启动服务器（支持 TLS）
	if config.CONFIG.Server.SSL {
		logger.LOG.Info("WebDAV 服务器正在启动（TLS）...",
			"url", fmt.Sprintf("https://%s%s/", addr, prefix),
		)
		if err := http.ListenAndServeTLS(addr, config.CONFIG.Server.SSLCert, config.CONFIG.Server.SSLKey, nil); err != nil {
			logger.LOG.Error("WebDAV 服务器启动失败", "error", err)
			return err
		}
	} else {
		logger.LOG.Warn("WebDAV 服务使用 HTTP（建议在反向代理后部署或启用 SSL）")
		logger.LOG.Info("WebDAV 服务器正在启动...",
			"url", fmt.Sprintf("http://%s%s/", addr, prefix),
		)
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.LOG.Error("WebDAV 服务器启动失败", "error", err)
			return err
		}
	}

	return nil
}

// getLockSystem 获取锁系统，优先使用 Redis 分布式锁，回退到内存锁
func (s *Server) getLockSystem() webdav.LockSystem {
	if s.lockSystem != nil {
		return s.lockSystem
	}
	// 回退到内存锁（单实例模式）
	return newPermissiveLockSystem()
}

// permissiveLockSystem 允许锁定不存在的文件的自定义锁系统（内存锁，仅单实例使用）
type permissiveLockSystem struct {
	ls webdav.LockSystem
}

func newPermissiveLockSystem() webdav.LockSystem {
	return &permissiveLockSystem{
		ls: webdav.NewMemLS(),
	}
}

func (pls *permissiveLockSystem) Confirm(now time.Time, name0, name1 string, conditions ...webdav.Condition) (release func(), err error) {
	return pls.ls.Confirm(now, name0, name1, conditions...)
}

func (pls *permissiveLockSystem) Create(now time.Time, details webdav.LockDetails) (token string, err error) {
	return pls.ls.Create(now, details)
}

func (pls *permissiveLockSystem) Refresh(now time.Time, token string, duration time.Duration) (webdav.LockDetails, error) {
	return pls.ls.Refresh(now, token, duration)
}

func (pls *permissiveLockSystem) Unlock(now time.Time, token string) error {
	return pls.ls.Unlock(now, token)
}
