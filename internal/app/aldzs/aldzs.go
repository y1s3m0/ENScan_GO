package aldzs

import (
	"crypto/tls"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/outputfile"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"net/http"
	"time"
	"os"
	"strconv"
)

func getReq(searchType string, data map[string]string,options *common.ENOptions) gjson.Result {
	//安全延时
	time.Sleep(time.Duration(options.DelayTime) * time.Second)

	//计算签名
	//构造ChinaZ请求
	client := resty.New()
	client.SetTimeout(time.Duration(options.TimeOut) * time.Minute)
	if options.Proxy != "" {
		client.SetProxy(options.Proxy)
	}
	url := fmt.Sprintf("https://zhishuapi.aldwx.com/Main/action/%s", searchType)
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	client.Header = http.Header{
		"User-Agent":   {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36"},
		"Accept":       {"text/html, application/xhtml+xml, image/jxr, */*"},
		"Content-Type": {"application/x-www-form-urlencoded; charset=UTF-8"},
		"Referer":      {"https://www.aldzs.com"},
	}
	clientR := client.R()
	clientR.Method = "POST"
	clientR.SetFormData(data)
	clientR.URL = url
	resp, err := clientR.Send()
	if err != nil {
		fmt.Println(err)
	}
	res := gjson.Parse(string(resp.Body()))
	if res.Get("code").String() != "200" {
		gologger.Errorf("【aldzs】似乎出了点问题 %s \n", res.Get("msg"))
	}
	return res.Get("data")
}

func GetInfoByKeyword(options *common.ENOptions) (ensInfos *common.EnInfos, ensOutMap map[string]*outputfile.ENSMap) {
	ensInfos = &common.EnInfos{}
	ensInfos.Infos = make(map[string][]gjson.Result)
	ensOutMap = make(map[string]*outputfile.ENSMap)

	keyword := options.KeyWord
	//拿到Token信息
	token := options.CookieInfo
	gologger.Infof("查询关键词 %s 的小程序\n", keyword)
	appList := getReq("Search/Search/search", map[string]string{
		"appName":    keyword,
		"page":       "1",
		"token":      token,
		"visit_type": "1",
	},options).Array()
	if len(appList) == 0 {
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NO", "ID", "小程序名称", "所属公司", "描述"})
	for k, v := range appList {
		table.Append([]string{
			strconv.Itoa(k),
			v.Get("id").String(),
			v.Get("name").String(),
			v.Get("company").String(),
			v.Get("desc").String(),
		})
	}
	table.Render()
	//默认取第一个进行查询
	var appKey string
	if list, ok := options.CompanyCheckList["aldzs"]; ok {
		// 遍历列表，检查是否存在目标 value
		valueExists := false
		for i:=0;i<len(appList);i++{
			if appList[i].Get("appKey").String()==""&&appList[i].Get("appKey").String()=="0"{
				continue
			}
			for _, v := range list {
				if v == appList[i].Get("appKey").String() {
					gologger.Infof("已查询过 '%s'\n", appList[i].Get("company"))
					valueExists = true
					break
				}
			}
			if !valueExists {
				gologger.Infof("查询 %s 开发的相关小程序 【默认取100个】\n", appList[i].Get("company"))
				appKey = appList[i].Get("appKey").String()
				break
			}
		}
	} else {
		gologger.Errorf("options.CompanyCheckList aldzs不存在\n")
	}

	sAppList := getReq("Miniapp/App/sameBodyAppList", map[string]string{
		"appKey": appKey,
		"page":   "1",
		"size":   "100",
		"token":  token,
	},options).Array()
	ensInfos.Infos["wx_app"] = sAppList
	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NO", "ID", "小程序名称", "描述"})
	for k, v := range sAppList {
		table.Append([]string{
			strconv.Itoa(k),
			v.Get("id").String(),
			v.Get("name").String(),
			v.Get("desc").String(),
		})
	}
	table.Render()

	for k, v := range getENMap() {
		ensOutMap[k] = &outputfile.ENSMap{Name: v.name, Field: v.field, KeyWord: v.keyWord}
	}
	return ensInfos, ensOutMap
}

type EnsGo struct {
	name     string
	api      string
	fids     string
	params   map[string]string
	field    []string
	keyWord  []string
	typeInfo []string
}

func getENMap() map[string]*EnsGo {
	ensInfoMap := make(map[string]*EnsGo)
	ensInfoMap = map[string]*EnsGo{
		"wx_app": {
			name:    "微信小程序",
			field:   []string{"name", "categoryTitle", "logo", "", ""},
			keyWord: []string{"名称", "分类", "头像", "二维码", "阅读量"},
		},
	}
	for k, _ := range ensInfoMap {
		ensInfoMap[k].keyWord = append(ensInfoMap[k].keyWord, "数据关联  ")
		ensInfoMap[k].field = append(ensInfoMap[k].field, "inFrom")
	}
	return ensInfoMap
}
