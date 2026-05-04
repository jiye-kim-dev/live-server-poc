## live-server-poc
오픈소스 기반으로 라이브 서버 만들어보기 (PoC)

---
Ref) <br>
https://mediamtx.org/docs/kickoff/install
 
### 전체 흐름(간략)
```text
외부 API 호출 (앱이름/스트림키 발급)
    ↓
   서버  → MediaMTX Control API로 path 등록
    ↓
OBS/FFmpeg → rtmp://host:1935/{앱이름}/{스트림키} push
    ↓
방송 종료 → 서버 → MediaMTX Control API로 path 삭제
```