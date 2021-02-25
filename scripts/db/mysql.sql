/**
//手动执行脚本,创建数据库,账号.
create database `uploadserver` CHARACTER SET utf8mb4;
use uploadserver;
CREATE USER 'uploadserver'@'127.0.0.1' IDENTIFIED BY '123456';
CREATE USER 'uploadserver'@'localhost' IDENTIFIED BY '123456';
CREATE USER 'uploadserver'@'%' IDENTIFIED BY '123456';

grant DELETE,EXECUTE,INSERT,SELECT,UPDATE on uploadserver.* to 'uploadserver'@'127.0.0.1';
grant DELETE,EXECUTE,INSERT,SELECT,UPDATE on uploadserver.* to 'uploadserver'@'localhost';
grant DELETE,EXECUTE,INSERT,SELECT,UPDATE on uploadserver.* to 'uploadserver'@'%';

FLUSH PRIVILEGES;
**/

use uploadserver;

create table if not exists `uploadfile` (
    `rid` varchar(64) not null primary key comment '资源ID',
    `app_id` varchar(64) null comment '资源所属app_id',
    hash varchar(64) not null comment '资源HASH',
    size int default 0 comment '资源大小',
    path varchar(256) not null comment '资源URL',
    width int default 0 comment '图片视频宽度(像素)',
    height int default 0 comment '图片视频高度(像素)',
    duration int default 0 comment '视频时长(秒)',
    extinfo text null comment '扩展信息，一般是json格式',
    create_time int not null comment '创建时间',
    update_time int not null comment '更新时间',
    KEY(`hash`),
    KEY(`path`(128)),
    KEY(`create_time`)
)ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE utf8mb4_bin comment '文件信息表';
