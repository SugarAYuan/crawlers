package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"crawlers/model"
	"crawlers/service"

	"github.com/PuerkitoBio/goquery"
)

var idMaps map[string]string

func WeiChatHandler(context *service.CrawlersContext) {
	CrawlersContext = context
	idMaps := getTagMaps()
	for {
		select {
		case <-time.Tick(1 * time.Hour):
			if len(idMaps) > 0 {
				for k, v := range idMaps {
					doWeiChat(v, k, 0)
				}
			}
		}
	}

}

func getTagMaps() map[string]string {
	res, err := http.Get(`http://weixin.sogou.com/`)
	idMaps := make(map[string]string, 0)
	if err != nil {
		CrawlersContext.Logrus.Error("http get weixin.sogou.com " + err.Error())
		return idMaps
	}

	defer res.Body.Close()

	root, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		CrawlersContext.Logrus.Error("goquery new document err" + err.Error())
		return idMaps
	}

	root.Find("#type_tab>.fieed-box>a").Each(func(i int, selection *goquery.Selection) {
		id, _ := selection.Attr("id")
		if id == "more_anchor" || id == "" {
			return
		}
		idMaps[strings.TrimLeft(id, "pc_")] = selection.Text()
	})

	root.Find("#hide_tab>a").Each(func(i int, selection *goquery.Selection) {
		id, _ := selection.Attr("id")
		if id == "more_anchor" || id == "" {
			return
		}
		idMaps[strings.TrimLeft(id, "pc_")] = selection.Text()
	})

	CrawlersContext.Logrus.Info(idMaps)
	return idMaps
}

func doWeiChat(tgName, id string, page int) {
	var url string
	if page == 0 {
		url = fmt.Sprintf("http://weixin.sogou.com/pcindex/pc/pc_%v/pc_%v.html", id, id)
	} else {
		url = fmt.Sprintf("http://weixin.sogou.com/pcindex/pc/pc_%v/%v.html", id, page)
	}

	CrawlersContext.Logrus.Info(url)

	page++
	if page == 16 {
		return
	}

	res, err := http.Get(url)

	if err != nil {
		CrawlersContext.Logrus.Error("http get " + url + err.Error())
		return
	}

	defer res.Body.Close()

	root, err := goquery.NewDocumentFromReader(res.Body)

	if err != nil {
		CrawlersContext.Logrus.Error("goquery new document err" + err.Error())
		return
	}

	root.Find("li").Each(func(i int, selection *goquery.Selection) {

		data := new(model.WechatData)
		data.Title = selection.Find("h3").Text()
		data.Url, _ = selection.Find("h3>a").Attr("href")
		data.Content = selection.Find(".txt-info").Text()
		data.Source = selection.Find(".account").Text()
		data.Image, _ = selection.Find("a>.img-box>img").Attr("src")
		data.TagId = id
		data.TagName = tgName
		CrawlersContext.Logrus.Info(data.Title)
		inserSql := `
replace into wechat_data(tag_id,tag_name,title,content,image,source,url) values('%v','%v','%v','%v','%v','%v','%v')
`
		_, err = CrawlersContext.MysqlClient.Exec(fmt.Sprintf(inserSql,
			data.TagId,
			data.TagName,
			strings.Replace(data.Title, "'", `\'`, -1),
			strings.Replace(data.Content, "'", `\'`, -1),
			data.Image,
			strings.Replace(data.Source, "'", `\'`, -1),
			data.Url))

		if err != nil {
			CrawlersContext.Logrus.Error("insert data err" + err.Error())
			return
		}

	})

	doWeiChat(tgName, id, page)

}
