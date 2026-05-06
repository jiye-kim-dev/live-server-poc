package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// 스트림 정보 키 - stream:{appName}/{streamKey}
	streamKeyPrefix = "stream:"
	// 시청자 세션 키 - viewer:{appName}/{streamKey}:{userID}
	viewerKeyPrefix = "viewer:"
	// 시청자 세션 TTL (하트비트 없으면 자동 만료)
	viewerTTL = 30 * time.Second
)

type StreamInfo struct {
	AppName   string            `json:"app_name"`
	StreamKey string            `json:"stream_key"`
	CreatedAt time.Time         `json:"created_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type Store struct {
	rdb *redis.Client
}

func NewStore(addr string) *Store {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &Store{
		rdb: rdb,
	}
}

func (s *Store) RegisterStream(ctx context.Context, info *StreamInfo) error {
	key := streamKey(info.AppName, info.StreamKey)
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, key, data, 0).Err() // TTL 없음 (명시적 삭제)
}

func (s *Store) UnregisterStream(ctx context.Context, appName, key string) error {
	result, err := s.rdb.Del(ctx, streamKey(appName, key)).Result()
	if err != nil {
		return err
	}
	// 이미 없어도 에러 아님 (멱등성 보장하도록 처리)
	_ = result
	return nil
}

// IsStreamRegistered : 스트림 확인
func (s *Store) IsStreamRegistered(ctx context.Context, appName, key string) (bool, error) {
	n, err := s.rdb.Exists(ctx, streamKey(appName, key)).Result()
	return n > 0, err
}

// GetStream : 스트림 정보 조회
func (s *Store) GetStream(ctx context.Context, appName, key string) (*StreamInfo, error) {
	data, err := s.rdb.Get(ctx, streamKey(appName, key)).Bytes()
	if err != nil {
		return nil, err
	}
	var info StreamInfo
	return &info, json.Unmarshal(data, &info)
}

// AddViewer - 시청자 세션 등록 (중복재생 차단)
func (s *Store) AddViewer(ctx context.Context, appName, streamKey, userID string) error {
	key := viewerKey(appName, streamKey, userID)
	// SET NX: 이미 있으면 실패 → 중복재생 차단
	ok, err := s.rdb.SetNX(ctx, key, time.Now().Unix(), viewerTTL).Result()
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("이미 재생 중인 세션이 있습니다: userID=%s", userID)
	}
	return nil
}

// RemoveViewer - 시청자 세션 해제
func (s *Store) RemoveViewer(ctx context.Context, appName, streamKey, userID string) error {
	return s.rdb.Del(ctx, viewerKey(appName, streamKey, userID)).Err()
}

// IsViewerActive - 시청자 세션 존재 여부
func (s *Store) IsViewerActive(ctx context.Context, appName, streamKey, userID string) (bool, error) {
	n, err := s.rdb.Exists(ctx, viewerKey(appName, streamKey, userID)).Result()
	return n > 0, err
}

func streamKey(appName, key string) string {
	return fmt.Sprintf("%s%s/%s", streamKeyPrefix, appName, key)
}

func viewerKey(appName, sKey, userID string) string {
	return fmt.Sprintf("%s%s/%s:%s", viewerKeyPrefix, appName, sKey, userID)
}
