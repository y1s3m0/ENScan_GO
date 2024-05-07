package shangjibao

/* shangjibao By dot5
 */
import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/outputfile"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"strings"
)

// pageParseJson 提取页面中的JSON字段
func pageParseJson(content string) gjson.Result {

	tag1 := "window.pageData ="
	tag2 := "window.isSpider ="
	//tag2 := "/* eslint-enable */</script><script data-app"
	idx1 := strings.Index(content, tag1)
	idx2 := strings.Index(content, tag2)
	if idx2 > idx1 {
		str := content[idx1+len(tag1) : idx2]
		str = strings.Replace(str, "\n", "", -1)
		str = strings.Replace(str, " ", "", -1)
		str = str[:len(str)-1]
		return gjson.Get(string(str), "result")
	} else {
		gologger.Errorf("无法解析信息错误信息%s\n", content)
	}
	return gjson.Result{}
}

// GetInfoByKeyword 获取公司信息及备案
// options options
func GetInfoByKeyword(options *common.ENOptions)(ensInfos *common.EnInfos, ensOutMap map[string]*outputfile.ENSMap) {
	ensInfos = &common.EnInfos{}
	ensInfos.Infos = make(map[string][]gjson.Result)
	ensOutMap = make(map[string]*outputfile.ENSMap)
	ensInfos.Name = options.KeyWord
	// 获取初始化API数据
	ensInfoMap := getENMap()

	//获取数据
	s := ensInfoMap["icp"]

	gologger.Infof("SJB api查询 %s\n", options.KeyWord)
	dataList := getInfoList(s.api, options)
	//判断下网站备案，然后提取出来，处理下数据
	var tmp []gjson.Result
	for _, d := range dataList {
		entName:=strings.Replace(strings.Replace(d.Get("entName").String(), "<em>", "", -1), "</em>", "", -1)
		recordTime:=d.Get("recordTime").String()
		for _, y := range d.Get("icpInfoDetails").Array() {
			for _, o := range y.Get("domainName").Array() {
				valueTmp, _ := sjson.Set(y.Raw, "domainName", o.String())
				valueTmp, _ = sjson.Set(valueTmp, "siteUrl", y.Get("siteUrl").Array()[0].String())
				valueTmp, _ = sjson.Set(valueTmp, "enName", entName)
				valueTmp, _ = sjson.Set(valueTmp, "recordTime", recordTime)
				tmp = append(tmp, gjson.Parse(valueTmp))
			}
		}
	}
	dataList = tmp

	// 添加来源信息，并把信息存储到数据里面
	for _, y := range dataList {
		valueTmp, _ := sjson.Set(y.Raw, "inFrom", options.KeyWord)
		ensInfos.Infos["icp"] = append(ensInfos.Infos["icp"], gjson.Parse(valueTmp))
	}

	//命令输出展示
	var data [][]string
	for _, y := range dataList {
		results := gjson.GetMany(y.Raw, ensInfoMap["icp"].field...)
		var str []string
		for _, ss := range results {
			str = append(str, ss.String())
		}
		data = append(data, str)
	}
	common.TableShow(ensInfoMap["icp"].keyWord, data, options)
	for k, v := range getENMap() {
		ensOutMap[k] = &outputfile.ENSMap{Name: v.name, Field: v.field, KeyWord: v.keyWord}
	}
	return ensInfos, ensOutMap
}

// getInfoList 获取信息列表
func getInfoList(types string, options *common.ENOptions) []gjson.Result {
	urls := "https://shangjibao.baidu.com/" + types
	params :=`{"param":{"unlockedRange":1,"page":{"currPage":1,"pageSize":50},"district":[],"sort":[],"query":"`+options.KeyWord+`","industry":[],"scopes":[1]}}`
	content := common.GetSjbReq(urls, params,options)
	var listData []gjson.Result
	if gjson.Get(string(content), "code").String() == "0" {
		data := gjson.Get(string(content), "data")
		listData = data.Get("dataList").Array()
	}
	return listData

}
