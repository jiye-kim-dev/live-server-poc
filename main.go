package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jiye-kim-dev/live-server-poc/handler"
	"github.com/jiye-kim-dev/live-server-poc/mediamtx"
	"github.com/jiye-kim-dev/live-server-poc/session"
)

func main() {
	redisAddr := getEnv("REDIS_ADDR", "redis:6379")
	store := session.NewStore(redisAddr)

	mtxAddr := getEnv("MTX_API_ADDR", "http://mediamtx:9997")
	mtxClient := mediamtx.NewClient(mtxAddr)

	authHandler := handler.NewAuthHandler(store, mtxClient)
	eventHandler := handler.NewEventHandler(store, mtxClient)
	streamHandler := handler.NewStreamHandler(store, mtxClient)

	r := gin.Default()

	// player.html
	r.StaticFile("/player", "./player.html")

	// 인증전용
	r.POST("/auth", authHandler.Authenticate)

	r.POST("/event/publish", eventHandler.OnPublish)
	r.POST("/event/unpublish", eventHandler.OnUnpublish)

	// 실제 서비스의 외부 API가 여기를 호출해서 방송 준비/종료 처리
	v1 := r.Group("/v1/streams")
	{
		v1.POST("", streamHandler.Register)                // 스트림 등록 + MediaMTX path 추가
		v1.DELETE("/:streamKey", streamHandler.Unregister) // 스트림 해제 + MediaMTX path 삭제
		v1.GET("", streamHandler.List)                     // 활성 스트림 목록
	}

	addr := getEnv("SERVER_ADDR", ":8085")
	log.Printf("Go 서버 시작: %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
