package har

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"
)

type HAR struct {
}

type Entries struct {
	Request Request `json:"request"`
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
}

// Parse parses the given HAR file and returns the parsed result.
func Parse(ctx g.Ctx, harFilePath string) (*Request, error) {
	if !gfile.Exists(harFilePath) {
		return nil, gerror.Newf(`har file "%s" does not exist`, harFilePath)
	}
	content := gfile.GetContents(harFilePath)
	hdrJson := gjson.New(content)
	entries := hdrJson.GetJson("log.entries")
	var entriesArray []Entries
	entries.Scan(&entriesArray)

	// 遍历entriesArray 查找URL包含 “fc/gt2/public_key” 的
	for _, entry := range entriesArray {
		if gstr.Contains(entry.Request.URL, "fc/gt2/public_key") {
			// g.Dump(entry)
			return &entry.Request, nil

		}
	}

	return nil, gerror.New("fc/gt2/public_key not found")
}
