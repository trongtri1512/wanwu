package request

type CreateCustomSkillReq struct {
	Name       string `json:"name" form:"name"`
	Avatar     Avatar `json:"avatar" form:"avatar"`
	Author     string `json:"author" form:"author" validate:"required"`
	Desc       string `json:"desc" form:"desc"`
	ZipUrl     string `json:"zipUrl" form:"zipUrl" validate:"required"`
	SaveId     string `json:"saveId" form:"saveId"`
	SourceType string `json:"sourceType" form:"sourceType"`
}

func (c *CreateCustomSkillReq) Check() error {
	return nil
}

type CustomSkillIDReq struct {
	SkillId string `json:"skillId" validate:"required"`
}

func (c *CustomSkillIDReq) Check() error {
	return nil
}

type DeleteCustomSkillReq struct {
	SkillId string `json:"skillId" validate:"required"`
}

func (c *DeleteCustomSkillReq) Check() error {
	return nil
}
