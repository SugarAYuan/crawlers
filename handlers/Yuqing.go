package handlers

import (
	"crawlers/service"
	"net/http"
	"strings"
	"github.com/PuerkitoBio/goquery"
	//"io/ioutil"
	"crawlers/model"
	"strconv"
	"time"
	"regexp"
	"fmt"
)

func YuqingHandler (context *service.CrawlersContext) {

	CrawlersContext = context
	isM := false
	for {
		select {
		case <- time.Tick(1 * time.Minute) :
			timeN := time.Now()
			if timeN.Minute() < 10 && !isM {
				timeLine , err := time.ParseInLocation("2006-01-02 15:04:05" ,timeN.Format("2006-01-02 15:00:00") , time.Local)
				if err != nil {
					CrawlersContext.Logrus.Warn(err)
					return
				}
				go shenma(timeLine , timeN)
				go weibo(timeLine , timeN)
				go baidu(timeLine , timeN)
				go wechat(timeLine , timeN)
				go zhihu(timeLine , timeN)
				go sogou(timeLine , timeN)
				isM = true
			} else if timeN.Minute() > 10 {
				isM = false
			}

		}
	}

}

func shenma (timeLine , timeN time.Time) {
	url := `https://m.sm.cn/s?q=%E7%A5%9E%E9%A9%AC%E7%83%AD%E6%90%9C&from=smor&by=submit&snum=1`
	userAgent := `Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Mobile Safari/537.36`
	host := `m.sm.cn`
	root , err := sendHotHttp(url , userAgent , host)
	if err != nil {
		CrawlersContext.Logrus.Warn(err.Error())
		return
	}
	reg := regexp.MustCompile(`\d+`)
	words := make([]*model.HotWords , 0)
	var hotMax float64
	root.Find(".news-top-list-content-li").Each(func(i int, selection *goquery.Selection) {
		word := new(model.HotWords)
		word.Word = selection.Find("div>.x-line-clamp-1").Text()

		CrawlersContext.Logrus.Info(word.Word , "shenma")
		word.Media = "shenma"
		t , err := strconv.Atoi(reg.FindString(selection.Find(".news-info-txt-icons>span").Text()))
		if err != nil {
			CrawlersContext.Logrus.Warn(err)
			return
		}
		word.HotNumber = int64(t)
		if hotMax == 0 && t > 0 {
			hotMax = float64(t)
			word.Heat = 1
		} else {
			word.Heat = float64(t) / hotMax
		}

		word.TimeLine  = timeLine
		word.CreateTime = timeN
		if word.Word == "" {
			return
		}

		words = append(words, word)

	})

	if len(words) < 1 {
		return
	}

	n , err := CrawlersContext.XormSession.NewSession().Insert(&words)

	if err != nil {
		CrawlersContext.Logrus.Error("Insert Db error " , err.Error())
		return
	}
	CrawlersContext.Logrus.Info("Insert data length:", n)
}

func weibo (timeLine , timeN time.Time) {
	url := `http://s.weibo.com/top/summary?cate=realtimehot`
	//userAgent := `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36`
	userAgent := `Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Mobile Safari/537.36`
	host := `s.weibo.com`
	root , err := sendHotHttp(url , userAgent , host)
	if err != nil {
		CrawlersContext.Logrus.Warn(err.Error())
		return
	}

	words := make([]*model.HotWords , 0)
	var hotMax float64
	root.Find(".list>.list_a>li>a").Each(func(i int, selection *goquery.Selection) {

		if i == 0 {
			return
		}
		reg := regexp.MustCompile(`\<em\S+`)
		tmpWords , err := selection.Find("span").Html()
		if err != nil {
			CrawlersContext.Logrus.Warn(err)
			return
		}
		tmpWords = reg.ReplaceAllString(tmpWords , "")
		reg = regexp.MustCompile(`\n|\r|\s`)

		word := new(model.HotWords)
		word.Media = "weibo"
		word.Word = reg.ReplaceAllString(tmpWords , "")
		CrawlersContext.Logrus.Info(word.Word , "weibo")
		t , err := strconv.Atoi(selection.Find("span>em>em").Text())
		if err != nil {
			CrawlersContext.Logrus.Warn(err)
			return
		}
		word.HotNumber = int64(t)
		if hotMax == 0 && t > 0 {
			hotMax = float64(t)
			word.Heat = 1
		} else {
			word.Heat = float64(t) / hotMax
		}
		word.TimeLine = timeLine
		word.CreateTime = timeN
		if word.Word == "" {
			return
		}

		words = append(words, word)

	})

	if len(words) < 1 {
		return
	}

	n , err := CrawlersContext.XormSession.NewSession().Insert(&words)

	if err != nil {
		CrawlersContext.Logrus.Error("Insert Db error " , err.Error())
		return
	}
	CrawlersContext.Logrus.Info("Insert data length:", n)
}

func baidu (timeLine , timeN time.Time) {
	url := `http://top.baidu.com/buzz?b=1&c=513&fr=topbuzz_b341&charset=utf8`
	userAgent := `Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Mobile Safari/537.36`
	root , err := sendHotHttp(url , userAgent , "top.baidu.com")
	if err != nil {
		CrawlersContext.Logrus.Warn(err.Error())
		return
	}

	words := make([]*model.HotWords , 0)
	var hotMax float64
	root.Find(".list-table>tbody>tr").Each(func(i int, selection *goquery.Selection) {

		word := new(model.HotWords)
		word.Media = "baidu"
		word.Word = selection.Find(".keyword>.list-title").Text()
		if word.Word == "" {
			return
		}
		word.Word = ConvertToString(word.Word , "gbk" , "utf8")
		t , err := strconv.Atoi(selection.Find(".last>span").Text())
		if err != nil {
			CrawlersContext.Logrus.Warn(err)
			return
		}
		word.HotNumber = int64(t)
		if hotMax == 0 && t > 0 {
			hotMax = float64(t)
			word.Heat = 1
		} else {
			word.Heat = float64(t) / hotMax
		}

		word.TimeLine = timeLine
		word.CreateTime = timeN
		if word.Word == "" {
			return
		}

		words = append(words, word)
	})


	if len(words) < 1 {
		return
	}

	n , err := CrawlersContext.XormSession.NewSession().Insert(&words)

	if err != nil {
		CrawlersContext.Logrus.Error("Insert Db error " , err.Error())
		return
	}
	CrawlersContext.Logrus.Info("Insert data length:", n)

}

func wechat (timeLine , timeN time.Time) {
	url := `http://weixin.sogou.com/`
	userAgent := `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36`
	root , err := sendHotHttp(url , userAgent , "weixin.sogou.com")
	if err != nil {
		CrawlersContext.Logrus.Warn(err.Error())
		return
	}

	words := make([]*model.HotWords , 0)
	var hotMax float64
	//CrawlersContext.Logrus.Info(root.Find("#topwords>li").Html())
	root.Find("#topwords>li").Each(func(i int, selection *goquery.Selection) {
		word := new(model.HotWords)
		word.Media = "weixin"
		tmpStr , _ := selection.Find(".lan-line>span").Attr("style")
		t , err := strconv.Atoi(strings.TrimLeft(strings.TrimRight(tmpStr , "%") , "width:"))
		word.Word = selection.Find("a").Text()
		CrawlersContext.Logrus.Info(word.Word , "weixin")
		if err != nil {
			CrawlersContext.Logrus.Warn(err)
			return
		}
		word.HotNumber = int64(t)
		if hotMax == 0 && t > 0 {
			hotMax = float64(t)
			word.Heat = 1
		} else {
			word.Heat = float64(t) / hotMax
		}

		word.TimeLine = timeLine
		word.CreateTime = timeN
		if word.Word == "" {
			return
		}
		//CrawlersContext.Logrus.Info(word.Word , t , tmpStr)
		words = append(words, word)
	})


	if len(words) < 1 {
		return
	}

	n , err := CrawlersContext.XormSession.NewSession().Insert(&words)

	if err != nil {
		CrawlersContext.Logrus.Error("Insert Db error " , err.Error())
		return
	}
	CrawlersContext.Logrus.Info("Insert data length:", n)
}

func zhihu (timeLine , timeN time.Time) {
	url := `http://zhihu.sogou.com/`
	userAgent := `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36`
	root , err := sendHotHttp(url , userAgent , "weixin.sogou.com")
	if err != nil {
		CrawlersContext.Logrus.Warn(err.Error())
		return
	}

	words := make([]*model.HotWords , 0)
	var hotMax float64

	root.Find(".hot-news>li").Each(func(i int, selection *goquery.Selection) {
		word := new(model.HotWords)
		word.Media = "zhihu"
		tmpStr , _ := selection.Find(".lan-line>span").Attr("style")
		t , err := strconv.Atoi(strings.TrimLeft(strings.TrimRight(tmpStr , "%") , "width:"))
		word.Word = selection.Find("a").Text()
		CrawlersContext.Logrus.Info(word.Word , "zhihu")
		if err != nil {
			CrawlersContext.Logrus.Warn(err)
			return
		}
		word.HotNumber = int64(t)
		if hotMax == 0 && t > 0 {
			hotMax = float64(t)
			word.Heat = 1
		} else {
			word.Heat = float64(t) / hotMax
		}

		word.TimeLine = timeLine
		word.CreateTime = timeN
		if word.Word == "" {
			return
		}
		//CrawlersContext.Logrus.Info(word.Word , t , tmpStr)
		words = append(words, word)
	})


	if len(words) < 1 {
		return
	}

	n , err := CrawlersContext.XormSession.NewSession().Insert(&words)

	if err != nil {
		CrawlersContext.Logrus.Error("Insert Db error " , err.Error())
		return
	}
	CrawlersContext.Logrus.Info("Insert data length:", n)
}

func sogou (timeLine , timeN time.Time) {
	page := 1
SOGOPAGE:
	if page > 3 {
		return
	}
	url := fmt.Sprintf(`http://top.sogou.com/hot/shishi_%d.html` , page)
	userAgent := `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36`
	root , err := sendHotHttp(url , userAgent , "weixin.sogou.com")
	if err != nil {
		CrawlersContext.Logrus.Warn(err.Error())
		return
	}

	words := make([]*model.HotWords , 0)
	var hotMax float64

	root.Find(".pub-list>li").Each(func(i int, selection *goquery.Selection) {
		word := new(model.HotWords)
		word.Media = "sogou"
		if page == 1 && i < 3 {
			word.Word = selection.Find(".s2>.p1>a" ).Text()
		} else {
			word.Word = selection.Find(".s2>.p3>a").Text()
		}

		t , err := strconv.Atoi(selection.Find(".s3").Text())
		if err != nil {
			CrawlersContext.Logrus.Warn(err)
			return
		}
		CrawlersContext.Logrus.Info(word.Word , "sogou")
		if err != nil {
			CrawlersContext.Logrus.Warn(err)
			return
		}
		word.HotNumber = int64(t)
		if hotMax == 0 && t > 0 {
			hotMax = float64(t)
			word.Heat = 1
		} else {
			word.Heat = float64(t) / hotMax
		}

		word.TimeLine = timeLine
		word.CreateTime = timeN
		if word.Word == "" {
			return
		}
		//CrawlersContext.Logrus.Info(word.Word , t , tmpStr)
		words = append(words, word)
	})


	if len(words) < 1 {
		return
	}

	n , err := CrawlersContext.XormSession.NewSession().Insert(&words)

	if err != nil {
		CrawlersContext.Logrus.Error("Insert Db error " , err.Error())
		return
	}
	CrawlersContext.Logrus.Info("Insert data length:", n)
	page = page + 1
	goto SOGOPAGE
}


func sendHotHttp (url ,userAgent,host string) (*goquery.Document , error) {

	var root *goquery.Document

	client := &http.Client{}

	req , err := http.NewRequest("GET" , url , strings.NewReader(""))

	if err != nil {
		return root , err
	}

	req.Header.Set("User-Agent" , userAgent)
	//req.Header.Set("Accept-Charset" , "UTF-8")
	req.Header.Set("Host" , host)

	res , err := client.Do(req)

	if err != nil {
		return root , err
	}

	defer res.Body.Close()

	root , err = goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return root , err
	}

	return root , nil
}