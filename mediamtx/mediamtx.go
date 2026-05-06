package mediamtx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Control API 전용
const (
	ResourceRtmpConns      = "rtmpconns"
	ResourceRtspSessions   = "rtspsessions"
	ResourceSrtConns       = "srtconns"
	ResourceWebrtcSessions = "webrtcsessions"
)

// MtxClient : MediaMTX Control API 클라이언트
type MtxClient struct {
	baseURL    string
	httpClient *http.Client
}

type PathConfig struct {
	Source         string `json:"source,omitempty"`
	MaxReaders     int    `json:"max_readers,omitempty"`
	RunOnPublish   string `json:"run_on_publish,omitempty"`
	RunOnUnpublish string `json:"run_on_unpublish,omitempty"`
}

type PathList struct {
	ItemCount int    `json:"itemCount"`
	PageCount int    `json:"pageCount"`
	Items     []Path `json:"items"`
}

type Path struct {
	Name          string   `json:"name"`
	Source        *Source  `json:"source"`
	Tracks        []string `json:"tracks"`
	BytesReceived int64    `json:"bytesReceived"`
	BytesSent     int64    `json:"bytesSent"`
	Readers       []Reader `json:"readers"`
}

type Source struct {
	Type string `json:"type"`
	Id   string `json:"id"`
}

func (v *Source) Resource() string {
	return toResource(v.Type)
}

type Reader struct {
	Type string `json:"type"`
	Id   string `json:"id"`
}

func (v *Reader) Resource() string {
	return toResource(v.Type)
}

func NewClient(baseURL string) *MtxClient {
	return &MtxClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// AddPath : 방송 스트림 동적 등록 (wowza 에서 appName, streamKey 조합해서 요청하는 그 경로 맞음)
func (v *MtxClient) AddPath(pathName string, config *PathConfig) error {
	body, err := json.Marshal(config)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v3/config/paths/add/%s", v.baseURL, pathName)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[MediaMTX] AddPath fail : status=%d, path=%s", resp.StatusCode, pathName)
	}

	return nil
}

// RemovePath : 방송 스트림 종료
func (v *MtxClient) RemovePath(pathName string) error {
	url := fmt.Sprintf("%s/v3/config/paths/delete/%s", v.baseURL, pathName)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil // 멱등성 보장
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[MediaMTX] RemovePath fail : status=%d, path=%s", resp.StatusCode, pathName)
	}

	return nil
}

// KickPublisher : 송출자 강제종료
func (v *MtxClient) KickPublisher(source *Source) error {
	return v.deleteConn(source.Resource(), source.Id)
}

// KickViewer : 시청자 강제종료
func (v *MtxClient) KickViewer(reader *Reader) error {
	return v.deleteConn(reader.Resource(), reader.Id)
}

// deleteConn : 연결 제거
func (v *MtxClient) deleteConn(resource, id string) error {
	url := fmt.Sprintf("%s/v3/%s/delete/%s", v.baseURL, resource, id)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil // 멱등성 보장
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[deleteConn] kick fail : status=%d, resource=%s, id=%s", resp.StatusCode, resource, id)
	}
	return nil
}

// ListPaths : 활성중인 스트림 목록 조회
func (v *MtxClient) ListPaths() (*PathList, error) {
	url := fmt.Sprintf("%s/v3/paths/list", v.baseURL)
	resp, err := v.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PathList
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func toResource(apiType string) string {
	switch apiType {
	case "rtmpConn":
		return ResourceRtmpConns
	case "srtConn":
		return ResourceSrtConns
	case "rtspSession":
		return ResourceRtspSessions
	case "webrtcSession":
		return ResourceWebrtcSessions
	default:
		return apiType
	}
}
