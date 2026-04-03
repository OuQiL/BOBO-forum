USE bobo_db;

CREATE TABLE `follows` (
    `id` BIGINT UNSIGNED NOT NULL COMMENT '关注业务ID (雪花算法)',
    `follower_id` BIGINT UNSIGNED NOT NULL COMMENT '关注者ID',
    `following_id` BIGINT UNSIGNED NOT NULL COMMENT '被关注者ID',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '关注时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_follower_following` (`follower_id`, `following_id`) COMMENT '防止重复关注',
    KEY `idx_follower` (`follower_id`, `created_at` DESC) COMMENT '关注列表查询',
    KEY `idx_following` (`following_id`, `created_at` DESC) COMMENT '粉丝列表查询',
    CONSTRAINT `chk_not_self` CHECK (`follower_id` != `following_id`) COMMENT '防止自己关注自己'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='关注关系表';
