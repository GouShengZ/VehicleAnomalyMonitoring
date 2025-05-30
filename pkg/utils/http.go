package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// HTTPClient 增强的HTTP客户端
type HTTPClient struct {
	client  *http.Client
	timeout time.Duration
}

// NewHTTPClient 创建新的HTTP客户端
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// DefaultHTTPClient 默认HTTP客户端
var DefaultHTTPClient = NewHTTPClient(30 * time.Second)

// CallAPI 向指定的 URL 发送 HTTP 请求，并处理响应。
// urlString: 目标 API 的 URL。
// method: HTTP 方法 (例如 "GET", "POST")。
// params: 要添加到 URL 查询字符串的参数。
// body: 用于 POST 请求的请求体（对于 GET 请求应为 nil）。
// result: 用于解码 JSON 响应的目标接口（应为指针）。
func CallAPI(urlString string, method string, params map[string]string, body []byte, result interface{}) error {
	// 解析 URL
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("error parsing URL: %w", err)
	}

	// 添加查询参数
	query := parsedURL.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	parsedURL.RawQuery = query.Encode()

	var req *http.Request

	// 根据方法创建请求
	switch method {
	case http.MethodGet: // 使用 http 常量更佳
		req, err = http.NewRequest(method, parsedURL.String(), nil)
	case http.MethodPost: // 使用 http 常量更佳
		reqBody := bytes.NewBuffer(body)
		req, err = http.NewRequest(method, parsedURL.String(), reqBody)
		if err == nil { // 仅在创建成功时设置 Header
			req.Header.Set("Content-Type", "application/json")
		}
	default:
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}
	if err != nil {
		return fmt.Errorf("error creating %s request: %w", method, err)
	}

	// 发送请求
	// 注意：建议复用 http.Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close() // 确保响应体被关闭

	// 读取响应体
	responseBody, err := io.ReadAll(resp.Body) // <--- 修改了变量名
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		// 注意：在错误中包含 responseBody 可能暴露敏感信息
		return fmt.Errorf("API request failed with status code: %d, body: %s", resp.StatusCode, string(responseBody))
	}

	// 解码 JSON 响应
	// 如果响应体为空，Unmarshal 会报错，需要根据 API 情况判断是否需要处理空响应体
	if len(responseBody) > 0 && result != nil {
		err = json.Unmarshal(responseBody, result)
		if err != nil {
			// 注意：在错误中包含 responseBody 可能暴露敏感信息
			return fmt.Errorf("error unmarshaling response: %w, body: %s", err, string(responseBody))
		}
	} else if result == nil && len(responseBody) > 0 {
		// 如果调用者不关心结果 (result is nil)，但有响应体，可能需要记录或警告
		// fmt.Printf("Warning: Received response body but no result variable provided to unmarshal into.\n")
	}

	return nil // 成功
}

// DownloadFile 从指定URL下载文件并保存到本地
// url: 下载文件的URL
// filepath: 保存文件的本地路径
func DownloadFile(url string, filepath string) error {
	// 发送GET请求
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 创建文件
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 将响应内容写入文件
	_, err = io.Copy(out, resp.Body)
	return err
}
