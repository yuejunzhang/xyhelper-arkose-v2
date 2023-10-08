package har

import (
	"encoding/base64"
	"net/url"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

type HAR struct {
}
type Bda struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}
type Entries struct {
	Request Request `json:"request"`

	StartedDateTime string `json:"startedDateTime"`
}

type Request struct {
	Method      string `json:"method"`
	URL         string `json:"url"`
	HTTPVersion string `json:"httpVersion"`
	Headers     []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"headers"`
	QueryString []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"queryString"`
	Cookies []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"cookies"`
	HeadersSize int `json:"headersSize"`
	BodySize    int `json:"bodySize"`
	PostData    struct {
		MimeType string `json:"mimeType"`
		Text     string `json:"text"`
		Params   []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"params"`
	} `json:"postData"`
	BX              string `json:"bx"`
	BV              string `json:"bv"`
	StartedDateTime string `json:"startedDateTime"`
}

// Parse parses the given HAR file and returns the parsed result.
func Parse(ctx g.Ctx, harFilePath string) (harRequest *Request, err error) {
	if !gfile.Exists(harFilePath) {
		return nil, gerror.Newf(`har file "%s" does not exist`, harFilePath)
	}
	content := gfile.GetContents(harFilePath)
	hdrJson := gjson.New(content)
	entries := hdrJson.GetJson("log.entries")
	// entries.Dump()

	var entriesArray []Entries
	entries.Scan(&entriesArray)

	// 遍历entriesArray 查找URL包含 “fc/gt2/public_key” 的
	for _, entry := range entriesArray {
		if gstr.Contains(entry.Request.URL, "fc/gt2/public_key") {
			params := entry.Request.PostData.Params
			arkBody := make(url.Values)

			var bda string
			var arkBx string

			for _, param := range params {
				if param.Name == "bda" {
					bda, err = url.QueryUnescape(param.Value)
					if err != nil {
						return nil, err
					}

				}
				if param.Name != "rnd" && param.Name != "bda" {
					query, err := url.QueryUnescape(param.Value)
					if err != nil {
						return nil, err
					}
					arkBody.Set(param.Name, query)
				}
			}
			if bda == "" {
				return nil, gerror.New("fc/gt2/public_key bda not found")
			}
			entry.Request.StartedDateTime = entry.StartedDateTime
			t, err := time.Parse(time.RFC3339, entry.StartedDateTime)
			if err != nil {
				return nil, err
			}
			bw := getBw(t.Unix())
			arkBx = Decrypt(bda, arkBody.Get("userbrowser")+bw)
			entry.Request.BX = arkBx
			entry.Request.BV = arkBody.Get("userbrowser")
			g.Dump(entry)
			gjson.New(entry.Request.BX).Dump()
			return &entry.Request, nil
		}
	}

	return nil, gerror.New("fc/gt2/public_key not found")
}

func getBt() int64 {
	return time.Now().UnixMicro() / 1000000
}
func getBw(bt int64) string {
	return strconv.FormatInt(bt-(bt%21600), 10)
}

func GetBdaWitBx(bx, bv string) string {
	bt := getBt()
	bw := getBw(bt)
	bxjson := gjson.New(bx)
	var bxArray []Bda
	bxjson.Scan(&bxArray)
	// 遍历数组
	for i := 0; i < len(bxArray); i++ {
		// n 时间戳
		if bxArray[i].Key == "n" {
			n := gbase64.EncodeString(gconv.String(gtime.Now().Unix()))
			bxjson.Set(gconv.String(i)+".value", n)
		}

	}
	bx = bxjson.MustToJsonString()
	bda := Encrypt(bx, bv+bw)
	return base64.StdEncoding.EncodeToString([]byte(bda))
}
