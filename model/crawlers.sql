CREATE DATABASE wechat_crawlers;

CREATE TABLE `wechat_data` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `tag_id` int(11) NOT NULL DEFAULT '0' COMMENT '标签ID（推荐，搞笑...）',
  `tag_name` varchar(255) NOT NULL DEFAULT '' COMMENT '标签名称',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '文章标题',
  `content` varchar(500) NOT NULL DEFAULT '' COMMENT '文章内容',
  `image` varchar(255) NOT NULL DEFAULT '' COMMENT '图片',
  `source` varchar(255) NOT NULL DEFAULT '' COMMENT '公众号来源',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT NULL COMMENT '修改时间',
  `delete_time` datetime DEFAULT NULL COMMENT '删除时间',
  `url` varchar(255) NOT NULL DEFAULT '' COMMENT '链接',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dataun` (`tag_id`,`title`,`image`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='微信数据';