package mp_common

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

// --- openapi request ---

type OcrReq struct {
	Files   *multipart.FileHeader `form:"file" json:"file" `
	OcrData *string               `form:"data" json:"data"`
	Url     *string               `form:"url" json:"url"`
}

func (req *OcrReq) Check() error {
	nonEmptyCount := 0
	if req.Files != nil {
		nonEmptyCount++
	}
	if req.OcrData != nil && *req.OcrData != "" {
		nonEmptyCount++
	}
	if req.Url != nil && *req.Url != "" {
		nonEmptyCount++
	}
	if nonEmptyCount != 1 {
		return fmt.Errorf("参数错误：Files、OcrData、Url 必须且只能传入一个有效参数")
	}
	return nil
}

func (req *OcrReq) Data() (map[string]interface{}, error) {
	m := make(map[string]interface{})
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// --- openapi response ---

type OcrResp struct {
	Code      int       `json:"code"`
	Message   string    `json:"message" validate:"required"`
	Version   string    `json:"version"`
	TimeStamp string    `json:"timestamp"`
	Id        string    `json:"id"`
	Sha1      string    `json:"sha1"`
	TimeCost  float64   `json:"time_cost"`
	FileName  string    `json:"filename"`
	OcrData   []OcrData `json:"data" validate:"required,dive"`
}

type OcrData struct {
	PageNum []int  `json:"page_num" validate:"required,min=1"`
	Type    string `json:"type" validate:"required"`
	Text    string `json:"text"`
	Length  int    `json:"length"`
}

// --- request ---

type IOcrReq interface {
	Data() *OcrReq
}

// ocrReq implementation of IOcrReq
type ocrReq struct {
	data *OcrReq
}

func NewOcrReq(data *OcrReq) IOcrReq {
	return &ocrReq{data: data}
}

func (req *ocrReq) Data() *OcrReq {
	return req.data
}

// --- response ---

type IOcrResp interface {
	String() string
	Data() (interface{}, bool)
	ConvertResp() (*OcrResp, bool)
}

// ocrResp implementation of IOcrResp
type ocrResp struct {
	raw string
}

func NewOcrResp(raw string) IOcrResp {
	return &ocrResp{raw: raw}
}

func (resp *ocrResp) String() string {
	return resp.raw
}

func (resp *ocrResp) Data() (interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("ocr resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *ocrResp) ConvertResp() (*OcrResp, bool) {
	var ret *OcrResp
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("ocr resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}

	if err := util.Validate(ret); err != nil {
		log.Errorf("ocr resp validate err: %v", err)
		return nil, false
	}
	return ret, true
}

// --- ocr ---

func Ocr(ctx *gin.Context, provider, apiKey, url string, req *OcrReq, headers ...Header) ([]byte, error) {
	if apiKey != "" {
		headers = append(headers, Header{
			Key:   "Authorization",
			Value: "Bearer " + apiKey,
		})
	}

	request := resty.New().
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}). // 关闭证书校验
		SetTimeout(0).                                             // 关闭请求超时
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "multipart/form-data").
		SetHeader("Accept", "application/json").
		SetDoNotParseResponse(true)
	for _, header := range headers {
		request.SetHeader(header.Key, header.Value)
	}
	// 根据不同参数类型，构建对应的请求
	var resp *resty.Response
	var err error
	switch {
	// 传入 Files（文件）
	case req.Files != nil:
		file, err := req.Files.Open()
		if err != nil {
			return nil, fmt.Errorf("request %v %v ocr err: %v", url, provider, err)
		}
		defer func() { _ = file.Close() }()
		request.SetFileReader("file", req.Files.Filename, file)

	// 传入 OcrData（base64 编码）
	case req.OcrData != nil && *req.OcrData != "":
		// 验证 base64 合法性
		if _, err := base64.StdEncoding.DecodeString(*req.OcrData); err != nil {
			return nil, fmt.Errorf("request %v %v ocr err: base64 编码无效 - %v", url, provider, err)
		}

		request.SetFormData(map[string]string{
			"data": *req.OcrData,
		})

	// 传入 Url（公网地址）
	case req.Url != nil && *req.Url != "":
		request.SetFormData(map[string]string{
			"url": *req.Url,
		})
	}
	resp, err = request.Post(url)
	if err != nil {
		return nil, fmt.Errorf("request %v %v ocr err: %v", url, provider, err)
	}
	b, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("request %v %v ocr read response body err: %v", url, provider, err)
	}
	if resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("request %v %v ocr http status %v msg: %v", url, provider, resp.StatusCode(), string(b))
	}
	return b, nil
}
