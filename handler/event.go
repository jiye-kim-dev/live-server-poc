package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiye-kim-dev/live-server-poc/mediamtx"
	"github.com/jiye-kim-dev/live-server-poc/session"
)

type EventRequest struct {
	Path     string `json:"path"`
	Protocol string `json:"protocol,omitempty"`
}

type EventHandler struct {
	store     *session.Store
	mtxClient *mediamtx.MtxClient
}

func NewEventHandler(store *session.Store, mtxClient *mediamtx.MtxClient) *EventHandler {
	return &EventHandler{store: store, mtxClient: mtxClient}
}

// OnPublish : 스트림 전송을 시작했을때
// POST /event/publish
func (v *EventHandler) OnPublish(c *gin.Context) {
	var req EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	appName, streamKey := parsePath(req.Path)
	log.Printf("[OnPublish] Stream start : app=%s, key=%s, protocol=%s", appName, streamKey, req.Protocol)

	// 방송 시작할때 이벤트 주고 싶으면 여기다 구현하면 좋을듯 ㅇㅂㅇ

	c.Status(http.StatusOK)
}

func (v *EventHandler) OnUnpublish(c *gin.Context) {
	var req EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	ctx := c.Request.Context()
	appName, streamKey := parsePath(req.Path)

	log.Printf("[OnUnpublish] Stream End : app=%s, key=%s", appName, streamKey)
	pathName := req.Path

	if err := v.mtxClient.RemovePath(pathName); err != nil {
		log.Printf("[OnUnpublish] Remove Path Error : %v", err)
	}

	if err := v.store.UnregisterStream(ctx, appName, streamKey); err != nil {
		log.Printf("[OnUnpublish] Unregister Stream Error : %v", err)
	}

	// 마찬가지로 방송종료 이벤트 구현하면 좋을듯 ㅇㅂㅇ

	c.Status(http.StatusOK)
}
