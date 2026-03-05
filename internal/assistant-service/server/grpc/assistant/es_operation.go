package assistant

import (
	"context"
	"encoding/json"
	"fmt"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/pkg/es"
	"google.golang.org/protobuf/types/known/emptypb"
)

// SaveToES saves a document to ES.
func (s *Service) SaveToES(ctx context.Context, req *assistant_service.SaveToESReq) (*emptypb.Empty, error) {
	if req.IndexName == "" {
		return nil, fmt.Errorf("index name is empty")
	}

	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(req.DocJson), &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal doc json: %v", err)
	}

	if err := es.Assistant().IndexDocument(ctx, req.IndexName, doc); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// DeleteFromES deletes documents from ES by conditions.
func (s *Service) DeleteFromES(ctx context.Context, req *assistant_service.DeleteFromESReq) (*emptypb.Empty, error) {
	if req.IndexName == "" {
		return nil, fmt.Errorf("index name is empty")
	}

	conditions := make(map[string]interface{})
	for k, v := range req.Conditions {
		conditions[k] = v
	}

	if err := es.Assistant().DeleteByFields(ctx, req.IndexName, conditions); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// SearchFromES searches documents in ES by conditions.
func (s *Service) SearchFromES(ctx context.Context, req *assistant_service.SearchFromESReq) (*assistant_service.SearchFromESResp, error) {
	if req.IndexName == "" {
		return nil, fmt.Errorf("index name is empty")
	}

	conditions := make(map[string]interface{})
	for k, v := range req.Conditions {
		conditions[k] = v
	}

	from := int((req.PageNo - 1) * req.PageSize)
	size := int(req.PageSize)

	docs, total, err := es.Assistant().SearchByFields(ctx, req.IndexName, conditions, from, size, req.SortOrder)
	if err != nil {
		return nil, err
	}

	docJsonList := make([]string, 0, len(docs))
	for _, doc := range docs {
		docJsonList = append(docJsonList, string(doc))
	}

	return &assistant_service.SearchFromESResp{
		DocJsonList: docJsonList,
		Total:       total,
	}, nil
}
