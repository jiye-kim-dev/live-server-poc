## live-server-poc
오픈소스 기반으로 라이브 서버 만들어보기 (PoC)

---
Ref) <br>
https://mediamtx.org/docs/kickoff/install
 
### 간략하게 전체흐름 정리
```text
외부 API 호출 (앱이름/스트림키 발급)
    ↓
   서버  → MediaMTX Control API로 path 등록
    ↓
OBS/FFmpeg → rtmp://host:1935/{앱이름}/{스트림키} push
    ↓
방송 종료 → 서버 → MediaMTX Control API로 path 삭제
```

### 실행방법
```shell
# 1. 전체 실행
docker-compose up -d

# 2. 스트림 등록
# 근데 이미 있는거면 먼저 해제 후 재등록 요망
curl -X DELETE "http://localhost:8085/v1/streams/test123?appName=live"
  
curl -X POST http://localhost:8085/v1/streams \
  -H "Content-Type: application/json" \
  -d '{"appName": "live", "streamKey": "test123"}'

# 3. FFmpeg으로 push
# 또는 OBS 로 직접 요청해도 됨
ffmpeg -re -f lavfi -i testsrc=size=1280x720:rate=30 \
  -c:v libx264 -preset ultrafast -b:v 1000k \
  -f flv rtmp://localhost:1935/live/test123

# 4. LL-HLS 재생
http://localhost:8888/live/test123/index.m3u8
```