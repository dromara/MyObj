package webdav

import (
	"fmt"
	"myobj/src/config"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/logger"
	"net/http"
	"time"

	"golang.org/x/net/webdav"
)

// Server WebDAV 服务器
type Server struct {
	auth    *Authenticator
	factory *impl.RepositoryFactory
}

// NewServer 创建 WebDAV 服务器实例
func NewServer(factory *impl.RepositoryFactory) *Server {
	auth := NewAuthenticator(
		factory.ApiKey(),
		factory.User(),
		factory.Power(),
	)

	return &Server{
		auth:    auth,
		factory: factory,
	}
}

// ServeHTTP 处理 WebDAV 请求
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. 认证
	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="MyObj WebDAV"`)
		http.Error(w, "需要认证", http.StatusUnauthorized)
		logger.LOG.Warn("WebDAV 请求缺少认证信息", "ip", r.RemoteAddr, "path", r.URL.Path)
		return
	}

	// 2. 验证用户
	user, err := s.auth.Authenticate(username, password)
	if err != nil {
		logger.LOG.Warn("WebDAV 认证失败",
			"username", username,
			"ip", r.RemoteAddr,
			"error", err,
		)
		w.Header().Set("WWW-Authenticate", `Basic realm="MyObj WebDAV"`)
		http.Error(w, "认证失败: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// 3. 检查 WebDAV 访问权限
	hasPermission, err := s.auth.CheckPermission(user.ID, user.GroupID, "webdav:access")
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
	handler := &webdav.Handler{
		FileSystem: fs,
		LockSystem: newPermissiveLockSystem(), // 使用自定义锁系统
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
				logger.LOG.Info("WebDAV 操作",
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

	// 注册处理器
	prefix := config.CONFIG.WebDAV.Prefix
	if prefix == "" {
		prefix = "/dav"
	}

	http.Handle(prefix+"/", http.StripPrefix(prefix, s))

	logger.LOG.Info("WebDAV 服务器正在启动...",
		"url", fmt.Sprintf("http://%s%s/", addr, prefix),
	)

	// 启动服务器
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.LOG.Error("WebDAV 服务器启动失败", "error", err)
		return err
	}

	return nil
}

// permissiveLockSystem 允许锁定不存在的文件的自定义锁系统
type permissiveLockSystem struct {
	ls webdav.LockSystem
}

func newPermissiveLockSystem() webdav.LockSystem {
	return &permissiveLockSystem{
		ls: webdav.NewMemLS(),
	}
}

func (pls *permissiveLockSystem) Confirm(now time.Time, name0, name1 string, conditions ...webdav.Condition) (release func(), err error) {
	// 总是返回成功，允许所有操作（单用户场景，不需要真正的锁）
	return func() {}, nil
}

func (pls *permissiveLockSystem) Create(now time.Time, details webdav.LockDetails) (token string, err error) {
	// 允许创建锁，即使文件不存在
	return pls.ls.Create(now, details)
}

func (pls *permissiveLockSystem) Refresh(now time.Time, token string, duration time.Duration) (webdav.LockDetails, error) {
	return pls.ls.Refresh(now, token, duration)
}

func (pls *permissiveLockSystem) Unlock(now time.Time, token string) error {
	return pls.ls.Unlock(now, token)
}
