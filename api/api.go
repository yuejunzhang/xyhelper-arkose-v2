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
	"github.com/gogf/gf/v2/encoding/gjson"
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
	if url == "" {
		r.Response.WriteJsonExit(g.Map{
			"code": 0,
			"msg":  "url不能为空",
		})
		return
	}

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
		g.Log().Error(ctx, getRealIP(r), err.Error())
		r.Response.WriteJsonExit(g.Map{
			"code": 0,
			"msg":  err.Error(),
		})
		return
	}
	defer reqCli.Close()
	response, err := reqCli.Request(ctx, "post", url, requests.RequestOption{
		Headers: Headers,
		Data:    payload,
		Cookies: harRequest.Cookies,
	})
	if err != nil {
		g.Log().Error(ctx, getRealIP(r), err.Error())
		r.Response.WriteJsonExit(g.Map{
			"code": 0,
			"msg":  err.Error(),
		})
		return
	}
	defer response.Close()
	text := response.Text()
	token := gjson.New(text).Get("token").String()
	if token == "" {
		g.Log().Error(ctx, getRealIP(r), text)
		r.Response.WriteJsonExit(g.Map{
			"code": 0,
			"msg":  "获取token失败: " + text,
		})
		return
	}
	// 如果不包含 sup=1|rid= 的字符串,那么就是失败了
	if !gstr.Contains(token, "sup=1|rid=") {
		g.Log().Error(ctx, getRealIP(r), token)
		r.Response.WriteJsonExit(g.Map{
			"code": 0,
			"msg":  "获取token失败: " + token,
		})
		return
	}
	g.Log().Info(ctx, getRealIP(r), token)

	r.Response.Status = response.StatusCode()
	r.Response.WriteJsonExit(text)

}

// Upload 上传har文件
func Upload(r *ghttp.Request) {
	ctx := r.Context()
	// 启用认证
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		r.Response.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		r.Response.Status = 401
		return
	}
	auth := strings.SplitN(authHeader, " ", 2)
	if len(auth) != 2 || auth[0] != "Basic" {
		r.Response.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		r.Response.Status = 401
		return
	}
	g.Dump(auth)

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
func getRealIP(req *ghttp.Request) string {
	// 优先获取Cf-Connecting-Ip
	if ip := req.Header.Get("Cf-Connecting-Ip"); ip != "" {
		return ip
	}

	// 优先获取X-Real-IP
	if ip := req.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// 其次获取X-Forwarded-For
	if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	// 最后获取RemoteAddr
	ip := req.RemoteAddr
	// 处理端口
	if index := strings.Index(ip, ":"); index != -1 {
		ip = ip[0:index]
	}
	if ip == "[" {
		ip = req.GetClientIp()
	}
	return ip
}
