# crawlers

> 这是一个基于goquery简单的爬虫小框架

```bash
#简单构建

go build

#action为启动项，目前就一个，以后会慢慢加
#config为配置文件切换
./crawlers -action="wechat" -config="dev"

#2018-08-27 add
./crawlers -action="yuqing" -config="dev"

```

> model里存放着数据库文件需要运行

> 目前只是简单的写了一个搜狗上微信公众号内容的抓取，将来会不断增加。

> 如果你有什么想要爬取的页面，可以留言给我我可以顺便写上
 