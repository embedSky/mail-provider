package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/open-falcon/mail-provider/config"
	"github.com/toolkits/web/param"
)

// 短信模板结构
type SmsTemplate struct {
	Time     string `json:"time"`
	Dev_name string `json:"dev_name"`
	Content  string `json:"content"`
}

func configSmsProcRoutes() {
	http.HandleFunc("/sender/sms", func(w http.ResponseWriter, r *http.Request) {
		cfg := config.Config()
		token := param.String(r, "token", "")
		if cfg.Http.Token != token {
			http.Error(w, "no privilege", http.StatusForbidden)
			return
		}

		content := param.MustString(r, "content")

		tos := param.MustString(r, "tos")
		phones := strings.Split(tos, ",")
		phonesJSONBytes, err := json.Marshal(phones)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		signString := cfg.Sms.SignName
		var signArr []string
		for index := 0; index < len(phones); index++ {
			signArr = append(signArr, signString)
		}
		signJSONBytes, err := json.Marshal(signArr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		tpl := SmsTemplate{
			Content:  content,
			Time:     time.Now().Format("2016-01-02 15:04:05"),
			Dev_name: "矿机",
		}
		var tplArr []SmsTemplate
		for index := 0; index < len(phones); index++ {
			tplArr = append(tplArr, tpl)
		}
		tplJSONBytes, err := json.Marshal(tplArr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		// 调用阿里云短息服务发送sms
		client, err := sdk.NewClientWithAccessKey("default", cfg.Sms.AccessKeyId, cfg.Sms.AccessKeySecret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		request := requests.NewCommonRequest()
		request.Method = "POST"
		request.Scheme = "https" // https | http
		request.Domain = "dysmsapi.aliyuncs.com"
		request.Version = "2017-05-25"
		request.ApiName = "SendBatchSms"
		request.QueryParams["PhoneNumberJson"] = string(phonesJSONBytes)
		request.QueryParams["SignNameJson"] = string(signJSONBytes)
		request.QueryParams["TemplateCode"] = cfg.Sms.TemplateCode
		request.QueryParams["TemplateParamJson"] = string(tplJSONBytes)

		response, err := client.ProcessCommonRequest(request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			fmt.Print(response.GetHttpContentString())
			http.Error(w, "success", http.StatusOK)
		}

	})
}
