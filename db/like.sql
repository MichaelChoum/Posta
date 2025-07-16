create database posta_like;
use posta_like;

CREATE TABLE `like_record` (
                               `id` bigint(20) SIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                               `biz_id` int NOT NULL DEFAULT 0 COMMENT '业务ID',
                               `obj_id` bigint(20) SIGNED NOT NULL DEFAULT '0' COMMENT '点赞对象id',
                               `user_id` bigint(20) SIGNED NOT NULL DEFAULT '0' COMMENT '用户ID',
                               `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                               `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
                               PRIMARY KEY (`id`),
                               KEY `ix_update_time` (`update_time`),
                               UNIQUE KEY `uk_biz_obj_uid` (`biz_id`,`obj_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='点赞记录表';

CREATE TABLE `like_count` (
                              `id` bigint(20) SIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                              `biz_id` int NOT NULL DEFAULT 0 COMMENT '业务ID',
                              `obj_id` bigint(20) SIGNED NOT NULL DEFAULT '0' COMMENT '点赞对象id',
                              `like_num` int(11) NOT NULL DEFAULT '0' COMMENT '点赞数',
                              `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                              `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
                              PRIMARY KEY (`id`),
                              KEY `ix_update_time` (`update_time`),
                              UNIQUE KEY `uk_biz_obj` (`biz_id`,`obj_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='点赞计数表';