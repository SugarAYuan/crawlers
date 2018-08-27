package handlers

import (
	"crawlers/service"
	"github.com/axgle/mahonia"
)

var CrawlersContext *service.CrawlersContext

func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcDnnResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcDnnResult), true)
	result := string(cdata)
	return result
}