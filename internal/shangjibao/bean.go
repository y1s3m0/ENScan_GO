package shangjibao

import (
	"github.com/tidwall/gjson"
)

type EnBen struct {
	Pid           string `json:"pid"`
	EntName       string `json:"entName"`
	EntType       string `json:"entType"`
	ValidityFrom  string `json:"validityFrom"`
	Domicile      string `json:"domicile"`
	EntLogo       string `json:"entLogo"`
	OpenStatus    string `json:"openStatus"`
	LegalPerson   string `json:"legalPerson"`
	LogoWord      string `json:"logoWord"`
	TitleName     string `json:"titleName"`
	TitleLegal    string `json:"titleLegal"`
	TitleDomicile string `json:"titleDomicile"`
	RegCap        string `json:"regCap"`
	Scope         string `json:"scope"`
	RegNo         string `json:"regNo"`
	PersonTitle   string `json:"personTitle"`
	PersonID      string `json:"personId"`
}

type EnsGo struct {
	name      string
	total     int64
	available int64
	api       string   //API 地址
	gNum      string   //判断数量大小的关键词
	field     []string //获取的字段名称 看JSON
	keyWord   []string //关键词
}

type EnInfo struct {
	Pid         string `json:"pid"`
	EntName     string `json:"entName"`
	legalPerson string
	openStatus  string
	email       string
	telephone   string
	branchNum   int64
	investNum   int64
	//info
	Infos  map[string][]gjson.Result
	ensMap map[string]*EnsGo
	//other
	investInfos map[string]EnInfo
	branchInfos map[string]EnInfo
}

type EnInfos struct {
	Name        string
	Pid         string
	legalPerson string
	openStatus  string
	email       string
	telephone   string
	branchNum   int64
	investNum   int64
	Infos       map[string][]gjson.Result
}

func getENMap() map[string]*EnsGo {
	ensInfoMap := make(map[string]*EnsGo)
	ensInfoMap = map[string]*EnsGo{
		"enterprise_info": {
			name:    "企业信息",
			field:   []string{"entName", "legalPerson", "openStatus", "", "", "", "startDate", "", "", "", ""},
			keyWord: []string{"企业名称", "法人代表", "经营状态", "电话", "邮箱", "注册资本", "成立日期", "注册地址", "经营范围", "统一社会信用代码", "PID"},
		},
		"icp": {
			name:    "ICP备案",
			api:     "crm/web/sjb/toker/queryenterpriserecommendlistwithicpinfo",
			field:   []string{"siteName", "siteUrl", "domainName", "icpNo", "enName","recordTime"},
			keyWord: []string{"网站名称", "网址", "域名", "网站备案/许可证号", "公司名称","备案时间"},
		},
	}
	for k, _ := range ensInfoMap {
		ensInfoMap[k].keyWord = append(ensInfoMap[k].keyWord, "数据关联  ")
		ensInfoMap[k].field = append(ensInfoMap[k].field, "inFrom")
	}
	return ensInfoMap

}

