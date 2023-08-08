package main

import (
	"xyhelper-arkose-v2/api"
	"xyhelper-arkose-v2/config"
	"xyhelper-arkose-v2/har"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	ctx := gctx.New()
	// harFilePath := "temp/localhost1845.har"
	// request, err := har.Parse(ctx, harFilePath)
	// if err != nil {
	// 	panic(err)
	// }
	// if request == nil {
	// 	panic("request is nil")
	// }
	// config.Cache.Set(ctx, "request", request, 0)
	newrequest := &har.Request{}
	config.Cache.MustGet(ctx, "request").Scan(newrequest)
	g.Dump(newrequest)
	s := g.Server()
	s.SetPort(config.PORT)
	s.BindHandler("/token", api.GetToken)
	s.BindHandler("/upload", api.Upload)
	s.Run()

}
