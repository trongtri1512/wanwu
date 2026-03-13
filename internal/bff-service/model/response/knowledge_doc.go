package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type DocPageResult struct {
	List             []*ListDocResp    `json:"list"`
	Total            int64             `json:"total"`
	PageNo           int               `json:"pageNo"`
	PageSize         int               `json:"pageSize"`
	DocKnowledgeInfo *DocKnowledgeInfo `json:"docKnowledgeInfo"`
}

type DocConfigResult struct {
	DocImportType     int32       `json:"docImportType"`     //文档导入类型，0：文件上传，1：url上传，2.批量url上传
	DocSegment        *DocSegment `json:"docSegment"`        //分段信息配置
	DocAnalyzer       []string    `json:"docAnalyzer"`       //文档解析类型
	ParserModelId     string      `json:"parserModelId"`     //ocr模型id
	AsrModelId        string      `json:"asrModelId"`        //asr模型id
	MultimodalModelId string      `json:"multimodalModelId"` //多模态模型id
	DocPreprocess     []string    `json:"docPreprocess"`     //文本预处理规则
}

type DocSegment struct {
	SegmentMethod  string   `json:"segmentMethod" validate:"required"` // 分段方法 0：通用分段；1：父子分段
	SegmentType    string   `json:"segmentType"`                       // 分段方式，只有通用分段必填 0：自动分段；1：自定义分段
	Splitter       []string `json:"splitter,omitempty"`                // 分隔符（只有自定义分段必填）
	MaxSplitter    *int     `json:"maxSplitter,omitempty"`             // 可分隔最大值（只有自定义分段必填）
	Overlap        *float32 `json:"overlap,omitempty"`                 // 可重叠值（只有自定义分段必填）
	SubSplitter    []string `json:"subSplitter,omitempty"`             // 分隔符（只有父子分段必填）
	SubMaxSplitter *int     `json:"subMaxSplitter,omitempty"`          // 可分隔最大值（只有父子分段必填）
}

type DocKnowledgeInfo struct {
	KnowledgeId     string          `json:"knowledgeId"`
	KnowledgeName   string          `json:"knowledgeName"`
	GraphSwitch     int32           `json:"graphSwitch"`
	ShowGraphReport bool            `json:"showGraphReport"`
	Description     string          `json:"description"`
	Keywords        []*KeywordsInfo `json:"keywords"`
	EmbeddingModel  *ModelInfo      `json:"embeddingModel"`
	LlmModelId      string          `json:"llmModelId"`
	Category        int32           `json:"category"` // 0: 知识库 1: 问答库 2: 多模态知识库
	Avatar          request.Avatar  `json:"avatar"`   // 头像
}

type ListDocResp struct {
	DocId         string `json:"docId"`
	DocName       string `json:"docName"`       //文档名称
	DocType       string `json:"docType"`       //文档类型
	KnowledgeId   string `json:"knowledgeId"`   //知识库id
	UploadTime    string `json:"uploadTime"`    //上传时间
	Status        int    `json:"status"`        //处理状态
	ErrorMsg      string `json:"errorMsg"`      //解析错误信息，预留
	FileSize      int64  `json:"fileSize"`      //文件大小，单位字节(Byte)
	SegmentMethod string `json:"segmentMethod"` //分段模式 0:通用分段，1：父子分段
	Author        string `json:"author"`        //上传文档 作者
	GraphStatus   int32  `json:"graphStatus"`   //图谱状态 0:待处理，1.解析中，2.解析成功，3.解析失败 -1. 当文档状态为解析失败时，显示 -
	GraphErrMsg   string `json:"graphErrMsg"`   //图谱错误信息
	IsMultimodal  bool   `json:"isMultimodal"`  // 是否为多模态文件
}

type DocImportTipResp struct {
	Message       string `json:"msg"`
	UploadStatus  int32  `json:"uploadstatus"`  //上传状态
	KnowledgeId   string `json:"knowledgeId"`   //知识库id
	KnowledgeName string `json:"knowledgeName"` //知识库名称
}

type DocSegmentResp struct {
	FileName            string            `json:"fileName"`            //名称
	PageTotal           int               `json:"pageTotal"`           //总页数
	SegmentTotalNum     int               `json:"segmentTotalNum"`     //分段数量
	MaxSegmentSize      int               `json:"maxSegmentSize"`      //设置最大长度
	SegmentType         string            `json:"segmentType"`         //分段方式 0自动分段 1自定义分段
	UploadTime          string            `json:"uploadTime"`          //上传时间
	Splitter            string            `json:"splitter"`            //分隔符（只有自定义分段必填）
	MetaDataList        []*DocMetaData    `json:"metaDataList"`        //文档元数据
	SegmentContentList  []*SegmentContent `json:"contentList"`         //内容
	SegmentImportStatus string            `json:"segmentImportStatus"` //分段导入状态描述
	SegmentMethod       string            `json:"segmentMethod"`       //分段方式 父子分段/通用分段
	DocAnalyzerText     []string          `json:"docAnalyzerText"`     //文档解析类型 文字提取 / OCR解析  / 模型解析
}

type DocMetaData struct {
	MetaKey       string `json:"metaKey"`       // key
	MetaValue     string `json:"metaValue"`     // 确定值
	MetaValueType string `json:"metaValueType"` // number，time, string
	MetaRule      string `json:"metaRule"`      // 正则表达式
	MetaId        string `json:"metaId"`        // 元数据id
}

type SegmentContent struct {
	Content    string   `json:"content"`
	Available  bool     `json:"available"`
	ContentId  string   `json:"contentId"`
	ContentNum int      `json:"contentNum"`
	Labels     []string `json:"labels"`
	IsParent   bool     `json:"isParent"` // 父子分段/通用分段 true是父分段，false是通用分段
	ChildNum   int      `json:"childNum"` // 子分段数量
}

type ChildSegmentInfo struct {
	Content  string `json:"content"`  // 内容
	ChildId  string `json:"childId"`  // 子分段id
	ChildNum int    `json:"childNum"` // 子分段序号
	ParentId string `json:"parentId"` // 父分段id
}

type AnalysisDocUrlResp struct {
	UrlList []*DocUrl `json:"urlList"`
}

type DocUrl struct {
	Url      string `json:"url"`
	FileName string `json:"fileName"`
	FileSize int    `json:"fileSize"`
}

type DocChildSegmentResp struct {
	SegmentContentList []*ChildSegmentInfo `json:"contentList"` //内容
}

type DocUploadLimitResp struct {
	UploadLimitList []*DocUploadLimit `json:"uploadLimitList"`
}

type DocUploadLimit struct {
	FileType string   `json:"fileType"` // 文件类型 图片：image 视频：video
	MaxSize  int      `json:"maxSize"`  // 文件大小限制，单位MB
	ExtList  []string `json:"extList"`  // 文件后缀列表
}
