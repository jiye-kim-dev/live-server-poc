package handler

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jiye-kim-dev/live-server-poc/mediamtx"
	"github.com/jiye-kim-dev/live-server-poc/session"
)

type RegisterRequest struct {
	AppName   string            `form:"app_name" binding:"required"`
	StreamKey string            `form:"stream_key" binding:"required"`
	Metadata  map[string]string `form:"metadata,omitempty"`
}

type StreamHandler struct {
	store     *session.Store
	mtxClient *mediamtx.MtxClient
}

func NewStreamHandler(store *session.Store, mtxClient *mediamtx.MtxClient) *StreamHandler {
	return &StreamHandler{
		store:     store,
		mtxClient: mtxClient,
	}
}

// Register : 스트림 등록
func (v *StreamHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	pathName := fmt.Sprintf("%s/%s", req.AppName, req.StreamKey)

	// 이미 등록된 스트림인지 먼저 확인
	exists, err := v.store.IsStreamRegistered(ctx, req.AppName, req.StreamKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "세션 조회 실패"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "이미 등록된 스트림입니다"})
		return
	}

	if err := v.mtxClient.AddPath(pathName, &mediamtx.PathConfig{}); err != nil {
		log.Printf("[Register] MediaMTX path addpath fail: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "path 등록 실패"})
		return
	}

	info := &session.StreamInfo{
		AppName:   req.AppName,
		StreamKey: req.StreamKey,
		CreatedAt: time.Now(),
		Metadata:  req.Metadata,
	}

	if err := v.store.RegisterStream(ctx, info); err != nil {
		_ = v.mtxClient.RemovePath(pathName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "세션 등록 실패"})
		return
	}

	log.Printf("[Register] MediaMTX path register success: %s", pathName)
	c.JSON(http.StatusOK, gin.H{
		"pathName":  pathName,
		"rtmpURL":   "rtmp://localhost:1935/" + pathName,
		"srtURL":    "srt://localhost:8890?streamid=publish:" + pathName,
		"hlsURL":    "http://localhost:8888/" + pathName + "/index.m3u8",
		"createdAt": info.CreatedAt,
	})
}

// Unregister - 스트림 해제
// DELETE /v1/streams/:streamKey?appName=xxx
func (v *StreamHandler) Unregister(c *gin.Context) {
	streamKey := c.Param("streamKey")
	appName := c.Query("appName")

	if appName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "appName required"})
		return
	}

	ctx := c.Request.Context()
	pathName := fmt.Sprintf("%s/%s", appName, streamKey)

	if err := v.mtxClient.RemovePath(pathName); err != nil {
		log.Printf("[Unregister] MediaMTX remove path fail: %v", err)
		// 로깅만 함
	}

	if err := v.store.UnregisterStream(ctx, appName, streamKey); err != nil {
		log.Printf("[Unregister] MediaMTX unregister fail: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "세션 해제 실패",
		})
		return
	}

	log.Printf("[Unregister] MediaMTX path unregister success: %s", pathName)
	c.Status(http.StatusOK)
}

func (v *StreamHandler) List(c *gin.Context) {
	data, err := v.mtxClient.ListPaths()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}
