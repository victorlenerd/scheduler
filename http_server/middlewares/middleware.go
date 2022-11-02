package middlewares

import (
	"context"
	"github.com/hashicorp/raft"
	"github.com/segmentio/ksuid"
	"log"
	"net/http"
	"scheduler0/http_server/middlewares/auth"
	"scheduler0/service"
	"scheduler0/utils"
	"strings"
	"sync"
)

const (
	RequestID = iota + 1
)

// middlewareHandler middleware type
type middlewareHandler struct {
	logger *log.Logger
	doOnce sync.Once
	ctx    context.Context
}

type MiddlewareHandler interface {
	ContextMiddleware(next http.Handler) http.Handler
	AuthMiddleware(credentialService service.Credential) func(next http.Handler) http.Handler
}

func NewMiddlewareHandler(logger *log.Logger) *middlewareHandler {
	return &middlewareHandler{
		logger: logger,
	}
}

// ContextMiddleware context middleware
func (m *middlewareHandler) ContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := ksuid.New().String()
		ctx := r.Context()
		ctx = context.WithValue(ctx, RequestID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthMiddleware authentication middleware
func (m *middlewareHandler) AuthMiddleware(credentialService service.Credential) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			paths := strings.Split(r.URL.Path, "/")
			if len(paths) < 1 {
				utils.SendJSON(w, "endpoint is not supported", false, http.StatusNotImplemented, nil)
				return
			}

			if paths[1] == "api-docs" || paths[1] == "healthcheck" {
				next.ServeHTTP(w, r)
				return
			}

			if auth.IsServerClient(r) {
				if validity, _ := auth.IsAuthorizedServerClient(r, credentialService); validity {
					next.ServeHTTP(w, r)
					return
				}
			}

			if auth.IsIOSClient(r) {
				if validity, _ := auth.IsAuthorizedIOSClient(r, credentialService); validity {
					next.ServeHTTP(w, r)
					return
				}
			}

			if auth.IsAndroidClient(r) {
				if validity, _ := auth.IsAuthorizedAndroidClient(r, credentialService); validity {
					next.ServeHTTP(w, r)
					return
				}
			}

			if auth.IsWebClient(r) {
				if validity, _ := auth.IsAuthorizedWebClient(r, credentialService); validity {
					next.ServeHTTP(w, r)
					return
				}
			}

			if auth.IsPeerClient(r) {
				if validity := auth.IsAuthorizedPeerClient(r, m.logger); validity {
					next.ServeHTTP(w, r)
					return
				}
			}

			utils.SendJSON(w, "unauthorized requests", false, http.StatusUnauthorized, nil)
			return
		})
	}
}

func (m *middlewareHandler) EnsureRaftLeaderMiddleware(raft *raft.Raft) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			af := raft.VerifyLeader()
			if af.Error() != nil && (r.Method == http.MethodPost || r.Method == http.MethodDelete || r.Method == http.MethodPut) {
				configs := utils.GetScheduler0Configurations(m.logger)
				_, serverAddr := raft.LeaderWithID()

				redirectUrl := ""

				for _, leaderPeer := range configs.Replicas {
					if leaderPeer.RaftAddress == string(serverAddr) {
						redirectUrl = leaderPeer.Address
						break
					}
				}

				if redirectUrl == "" {
					m.logger.Fatalln("failed to get redirect url from replicas")
				}

				http.Redirect(w, r, redirectUrl, 301)
				return
			}

			next.ServeHTTP(w, r)
			return
		})
	}
}
