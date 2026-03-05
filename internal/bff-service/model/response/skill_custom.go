package response

type CustomSkillDetail struct {
	SkillDetail
	ZipUrl string `json:"zipUrl"`
}

type CustomSkillIDResp struct {
	SkillId string `json:"skillId"`
}
