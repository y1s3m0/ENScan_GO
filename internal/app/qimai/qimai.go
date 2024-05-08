package qimai

import (
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/outputfile"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"strconv"
)

func GetInfoByKeyword(options *common.ENOptions) (ensInfos *common.EnInfos, ensOutMap map[string]*outputfile.ENSMap) {
	ensInfos = &common.EnInfos{}
	ensInfos.Infos = make(map[string][]gjson.Result)
	ensOutMap = make(map[string]*outputfile.ENSMap)
	ensInfos.Name = options.KeyWord
	params := map[string]string{
		"page":   "1",
		"search": options.KeyWord,
		"market": "1", //默认用360
	}
	res := gjson.Parse(GetReq("search/android", params, options)).Get("appList").Array()
	if len(res) == 0 {
		if options.IsDebug {
			gologger.Debugf("【查询错误信息】\n%s\n", ensInfos.Name)
		}
		gologger.Errorf("没有查询到关键词 “%s” \n", ensInfos.Name)
	} else {
		gologger.Infof("七麦关键词：“%s” 查询到 %d 个结果\n", ensInfos.Name, len(res))
		if list, ok := options.CompanyCheckList["qimai"]; ok {
			// 遍历列表，检查是否存在目标 value
			valueExists := false
			for i:=0;i<len(res);i++{
				if res[i].Get("company.id").String()==""&&res[i].Get("company.id").String()=="0"{
					continue
				}
				for _, v := range list {
					if v == res[i].Get("company.id").String() {
						gologger.Infof("已查询过 '%s'\n", res[i].Get("company.name"))
						valueExists = true
						break
					}
				}
				if !valueExists &&res[i].Get("company.id").Int()!=0{
					gologger.Infof("'%s'\n",res[i].Get("company.name").String())
					ensInfos.Infos = GetInfoByCompanyId(res[i].Get("company.id").Int(), options)
					break
				}
			}
		} else {
			gologger.Errorf("options.CompanyCheckList 七麦不存在\n")
		}
	}
	for k, v := range getENMap() {
		ensOutMap[k] = &outputfile.ENSMap{Name: v.name, Field: v.field, KeyWord: v.keyWord}
	}
	return ensInfos, ensOutMap
}

func GetInfoByCompanyId(companyId int64, options *common.ENOptions) (data map[string][]gjson.Result) {
	gologger.Infof("GetInfoByCompanyId: %d\n", companyId)
	data = map[string][]gjson.Result{}
	ensMap := getENMap()
	params := map[string]string{
		"id": strconv.Itoa(int(companyId)),
	}
	searchInfo := "enterprise_info"
	//gjson.GetMany(gjson.Get(GetReq(ensMap[searchInfo].api, params, options), "data").Raw, ensMap[searchInfo].field...)
	r, err := sjson.Set(gjson.Get(GetReq(ensMap[searchInfo].api, params, options), "data").Raw, "id", companyId)
	if err != nil {
		gologger.Errorf("Set pid error: %s", err.Error())
	}
	rs := gjson.Parse(r)
	data[searchInfo] = append(data[searchInfo], rs)
	params["page"] = "1"
	params["apptype"] = "2"
	searchInfo = "app"
	data[searchInfo] = append(data[searchInfo], getInfoList(ensMap["app"].api, params, options)...)
	//安卓
	params["page"] = "1"
	params["apptype"] = "3"
	data[searchInfo] = append(data[searchInfo], getInfoList(ensMap["app"].api, params, options)...)
	//命令输出展示
	var tdata [][]string
	for _, y := range data[searchInfo] {
		results := gjson.GetMany(y.Raw, ensMap[searchInfo].field...)
		var str []string
		for _, ss := range results {
			str = append(str, ss.String())
		}
		tdata = append(tdata, str)
	}
	common.TableShow(ensMap[searchInfo].keyWord, tdata, options)
	return data
}

func getInfoList(types string, params map[string]string, options *common.ENOptions) (listData []gjson.Result) {
	data := gjson.Parse(GetReq(types, params, options))
	if data.Get("code").String() == "10000" {
		getPath := "appList"
		getPage := "maxPage"
		if types == "company/getCompanyApplist" {
			if params["apptype"] == "2" {
				getPath = "ios"
			} else if params["apptype"] == "3" {
				getPath = "android"
			}
			getPage = getPath + "PageInfo.pageCount"
			getPath += "AppInfo"
			data = data.Get("data")
		}

		listData = append(listData, data.Get(getPath).Array()...)
		if data.Get(getPage).Int() <= 1 {
			return listData
		} else {
			for i := 2; i <= int(data.Get(getPage).Int()); i++ {
				gologger.Infof("getInfoList: %s %d\n", types, i)
				params["page"] = fmt.Sprintf("%d", i)
				listData = append(listData, gjson.Parse(GetReq(types, params, options)).Get("data."+getPath).Array()...)
			}
		}
		if len(listData) == 0 {
			gologger.Errorf("没有数据")
		}
	} else {
		gologger.Errorf("获取数据失败,请检查是否登陆\n")
		gologger.Debugf(data.Raw + "\n")
	}
	return listData
}
