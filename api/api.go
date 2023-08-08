package api

import (
	"log"
	"math/rand"
	"strconv"
	"strings"
	"xyhelper-arkose-v2/config"
	"xyhelper-arkose-v2/har"

	"gitee.com/baixudong/gospider/ja3"
	"gitee.com/baixudong/gospider/requests"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
)

// GetToken 获取token
func GetToken(r *ghttp.Request) {
	ctx := r.Context()
	harRequest := &har.Request{}
	config.Cache.MustGet(ctx, "request").Scan(harRequest)
	url := harRequest.URL
	headers := harRequest.Headers
	Headers := g.Map{}
	for _, v := range headers {
		// 如果v.Name 以 : 开头，那么就是一个特殊的请求头，需要特殊处理
		if gstr.HasPrefix(v.Name, ":") {
			continue
		}
		Headers[v.Name] = v.Value
	}
	// g.Dump(Headers)
	payload := harRequest.PostData.Text
	// 以&分割转换为数组
	payloadArray := gstr.Split(payload, "&")
	// 移除最后一个元素
	payloadArray = payloadArray[:len(payloadArray)-1]
	// 将 rnd=0.3046791926621015 添加到数组最后
	payloadArray = append(payloadArray, "rnd="+strconv.FormatFloat(rand.Float64(), 'f', -1, 64))
	// 以&连接数组
	payload = strings.Join(payloadArray, "&")
	// 生成指纹
	Ja3Spec, err := ja3.CreateSpecWithId(ja3.HelloFirefox_Auto) //根据id 生成指纹
	if err != nil {
		log.Panic(err)
	}
	reqCli, err := requests.NewClient(ctx, requests.ClientOption{
		Ja3Spec: Ja3Spec,
		H2Ja3:   true,
		Proxy:   config.PROXY,
	})
	if err != nil {
		log.Panic(err)
	}
	defer reqCli.Close()
	response, err := reqCli.Request(ctx, "post", url, requests.RequestOption{
		Headers: Headers,
		Data:    payload,
		Cookies: harRequest.Cookies,
	})
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 0,
			"msg":  err.Error(),
		})
		return
	}
	defer response.Close()
	text := response.Text()
	// 如果不包含 sup=1|rid= 的字符串,那么就是失败了
	if !gstr.Contains(text, "sup=1|rid=") {
		r.Response.WriteJsonExit(g.Map{
			"code": 0,
			"msg":  "获取token失败",
		})
		return
	}
	r.Response.Status = response.StatusCode()
	r.Response.WriteJsonExit(text)

}

// Upload 上传har文件
func Upload(r *ghttp.Request) {
	ctx := r.Context()
	// 启用认证

	if r.Method == "GET" {
		r.Response.WriteTpl("upload.html")
		return
	}
	// 上传文件
	files := r.GetUploadFiles("harFile")
	names, err := files.Save("./temp/")
	if err != nil {
		r.Response.WriteExit(err)
	}
	harRequset, err := har.Parse(ctx, "./temp/"+names[0])
	if err != nil {
		r.Response.WriteTpl("error.html", g.Map{
			"error": err.Error(),
		})
		return
	}
	err = config.Cache.Set(ctx, "request", harRequset, 0)
	if err != nil {
		r.Response.WriteTpl("error.html", g.Map{
			"error": err.Error(),
		})
		return
	}
	r.Response.WriteTpl("success.html")
}
