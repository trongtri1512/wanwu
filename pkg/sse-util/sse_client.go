package sse_util

import (
	"fmt"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

const DONE_MSG = "data: [DONE]\n"

const DONE_EMPTY = ""

type SSEWriterClient[Res any] interface {
	Header()
	Write(p Res) error
	Flush()
	ToRes(data string) Res
}

// GrpcStreamWriter grpc 流式写入器
type GrpcStreamWriter[Res any] struct {
	streamServer grpc.ServerStreamingServer[Res]
}

func (gsw *GrpcStreamWriter[Res]) Header() {
}

func (gsw *GrpcStreamWriter[Res]) Write(data Res) error {
	return gsw.streamServer.Send(&data)
}

func (gsw *GrpcStreamWriter[Res]) Flush() {
}

func (gsw *GrpcStreamWriter[Res]) ToRes(data string) Res {
	var zero Res
	return zero
}

type HttpSSEWriter[Res string] struct {
	ctx *gin.Context
}

func (hsw *HttpSSEWriter[string]) Header() {
	hsw.ctx.Header("Cache-Control", "no-cache")
	hsw.ctx.Header("Connection", "keep-alive")
	hsw.ctx.Header("Content-Type", "text/event-stream; charset=utf-8")
}

func (hsw *HttpSSEWriter[string]) Write(lineText string) error {
	_, err := hsw.ctx.Writer.Write([]byte(lineText))
	return err
}

func (hsw *HttpSSEWriter[Res]) Flush() {
	hsw.ctx.Writer.Flush()
}

func (hsw *HttpSSEWriter[string]) ToRes(data string) string {
	return data
}

// SSEWriter 设计sse writer 目标可以规范化统一标准输出方法（所有sse 返回都能用），同时与业务尽可能解耦
type SSEWriter[Res any] struct {
	client  SSEWriterClient[Res]
	label   string // 用于SSE日志中的标记
	doneMsg *Res   // SSE结束时，发送给前端的结束消息，空不发送；一般为 "data: [DONE]\n"
}

func NewSSEWriter(c *gin.Context, label string, doneMsg string) *SSEWriter[string] {
	http := &HttpSSEWriter[string]{ctx: c}
	http.Header()

	var msg *string
	if doneMsg != DONE_EMPTY {
		msg = &doneMsg
	}
	return &SSEWriter[string]{
		client:  http,
		label:   label,
		doneMsg: msg,
	}
}

func NewGrpcSSEWriter[Res any](streamServer grpc.ServerStreamingServer[Res], label string, doneMsg *Res) *SSEWriter[Res] {
	grpcStreamServer := &GrpcStreamWriter[Res]{streamServer: streamServer}
	return &SSEWriter[Res]{
		client:  grpcStreamServer,
		label:   label,
		doneMsg: doneMsg,
	}
}

// WriteStream 流式写入，识别channel 循环写入给前端
func (sw *SSEWriter[Res]) WriteStream(sseCh <-chan string, streamContextParams interface{},
	lineBuilder func(SSEWriterClient[Res], string, interface{}) (Res, bool, error),
	doneProcessor func(SSEWriterClient[Res], interface{}) error) error {
	for s := range sseCh {
		var lineText Res
		if lineBuilder != nil {
			line, skip, err := lineBuilder(sw.client, s, streamContextParams)
			if err != nil {
				log.Errorf("[SSE]%v line %v build err: %v", sw.label, err)
				return err
			}
			if skip {
				continue
			}
			lineText = line
		}
		if err := sw.WriteLine(lineText, false, streamContextParams, doneProcessor); err != nil {
			return err
		}
	}
	var zero Res
	return sw.WriteLine(zero, true, streamContextParams, doneProcessor)
}

// WriteLine 写入一行给客户端
func (sw *SSEWriter[Res]) WriteLine(lineText Res, done bool, streamProcessParams interface{},
	doneProcessor func(SSEWriterClient[Res], interface{}) error) error {

	var err error
	defer func() {
		if err != nil {
			log.Errorf("[SSE]%v err: %v", sw.label, err)
		} else if done {
			log.Debugf("[SSE]%v done", sw.label)
		} else {
			return
		}
		// err 或 done 执行 doneProcessor
		if doneProcessor != nil {
			if err := doneProcessor(sw.client, streamProcessParams); err != nil {
				log.Errorf("[SSE]%v doneProcessor err: %v", sw.label, err)
			}
		}
	}()

	if done && sw.doneMsg != nil {
		lineText = sw.client.ToRes(fmt.Sprintf("%v%v", lineText, *sw.doneMsg))
	}
	// 写入数据
	// log.Debugf("[SSE]%v write: %v", sw.label, lineText)
	if !EmptyValue(lineText) {
		err = sw.client.Write(lineText)
		if err != nil {
			err = fmt.Errorf("connection closed by web: %v", err)
			return err
		}
		sw.client.Flush()
	}
	return nil
}

// EmptyValue 空值判断
func EmptyValue[Res any](result Res) bool {
	if any(result) == nil {
		return true
	}
	switch v := any(result).(type) {
	case string:
		return v == ""
	case *string:
		return v == nil || *v == ""
	default:
		return false
	}
}
