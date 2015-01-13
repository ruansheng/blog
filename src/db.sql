DROP DATABASE IF EXISTS `blog`;
CREATE DATABASE `blog`;
USE blog;

/*文章 表*/
DROP TABLE IF EXISTS `article`;
CREATE TABLE `article` (
  `article_id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '文章id',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '文章标题',
  `content` varchar(1024) NOT NULL DEFAULT '' COMMENT '文章内容',
  `create_time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `update_time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `delete_time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '软删除时间',
  `is_del` tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '软删除',
  `show_num` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '被查看的次数',
  PRIMARY KEY (`article_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT '文章 表';

/*管理员 表*/
DROP TABLE IF EXISTS `manage`;
CREATE TABLE `manage` (
  `manage_id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '管理员id',
  `create_time` int(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `update_time` int(11) unsigned DEFAULT '0' COMMENT '更新时间',
  `delete_time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '软删除时间',
  `is_del` tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '软删除',
  `manage_name` varchar(60) NOT NULL DEFAULT '' COMMENT '管理员name',
  `password` varchar(100) NOT NULL DEFAULT '' COMMENT '密码',
  PRIMARY KEY (`manage_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT '管理员 表';