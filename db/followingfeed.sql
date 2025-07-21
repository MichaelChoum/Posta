create database posta_followingfeed;
use posta_followingfeed;

CREATE TABLE user_inbox (
                            id bigint(20) NOT NULL AUTO_INCREMENT,
                            user_id bigint(20) NOT NULL COMMENT '收件人',
                            article_id bigint(20) NOT NULL COMMENT '推送动态的文章ID',
                            sender_id bigint(20) NOT NULL COMMENT '动态作者ID',
                            status tinyint(4) NOT NULL DEFAULT '0' COMMENT '状态 0:待审核 1:审核不通过 2:可见 3:用户删除或者取消关注而删除',
                            publish_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '动态推入时间',
                            is_read TINYINT(1) NOT NULL DEFAULT 0 COMMENT '0未读、1已读',
                            PRIMARY KEY(id),
                            -- 确保一个用户对同一up主的同一篇动态（文章）只能进一次inbox，不能重复插入。
                            UNIQUE KEY uk_user_sender_article(user_id, sender_id, article_id),
                            KEY ix_user_publish_time(user_id, publish_time DESC),
                            -- 方便通过一篇文章ID，快速定位它被推送到了哪些用户的inbox。比如删除文章
                            KEY ix_article_id(article_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT '动态功能收信箱表';
