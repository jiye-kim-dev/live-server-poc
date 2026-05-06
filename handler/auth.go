package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jiye-kim-dev/live-server-poc/mediamtx"
	"github.com/jiye-kim-dev/live-server-poc/session"
)

type AuthRequest struct {
	User     string `json:"user"`
	Password string `json:"password"` // 계정정보가 필요하긴 하려나 ㅇㅂㅇ...
	Token    string `json:"token"`
	IP       string `json:"ip"`
	Action   string `json:"action"`
	Path     string `json:"path"`
	Protocol string `json:"protocol"`
	Id       string `json:"id"`
	Query    string `json:"query"`
}

type AuthHandler struct {
	store     *session.Store
	mtxClient *mediamtx.MtxClient
}

func NewAuthHandler(store *session.Store, mtxClient *mediamtx.MtxClient) *AuthHandler {
	return &AuthHandler{store: store, mtxClient: mtxClient}
}

func (v *AuthHandler) Authenticate(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	ctx := c.Request.Context()
	appName, streamKey := parsePath(req.Path)
	if appName == "" || streamKey == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	switch req.Action {
	case "publish":
		ok, err := v.store.IsStreamRegistered(ctx, appName, streamKey)
		if err != nil || !ok {
			c.Status(http.StatusUnauthorized)
			return
		}

	case "read":
		userId := extractParam(req.Query, "userId")
		if userId == "" {
			c.Status(http.StatusUnauthorized)
			return
		}

		// drm 이 있으면 검증할거임 ㅇㅂㅇ 근데 지금은 사실상 패스임
		if !validateDRMToken(req.Token, appName, streamKey) {
			c.Status(http.StatusUnauthorized)
			return
		}

		// 야매로 만들어본 중복재생차단 ㅇㅂㅇ
		if err := v.store.AddViewer(ctx, appName, streamKey, userId); err != nil {
			c.Status(http.StatusUnauthorized)
			return
		}
	default:
		// 임시 : 일단 허용함
	}

	c.Status(http.StatusOK)
}

func parsePath(path string) (appName, streamKey string) {
	parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func extractParam(query, key string) string {
	for _, part := range strings.Split(query, "&") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 && kv[0] == key {
			return kv[1]
		}
	}
	return ""
}

// validateDRMToken : DRM 토큰 검증
func validateDRMToken(token, appName, streamKey string) bool {
	// PoC 니까 그냥 아무 토큰값 들어와도 맞는걸로 간주하겠음
	return token != ""
}
