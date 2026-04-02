USE bobo_db;

CREATE TABLE `likes` (
    `id` BIGINT UNSIGNED NOT NULL COMMENT '点赞ID (雪花算法)',
    `type` TINYINT NOT NULL DEFAULT 0 COMMENT '点赞类型：0-帖子点赞，1-评论点赞',
    `target_id` BIGINT UNSIGNED NOT NULL COMMENT '目标ID（帖子ID或评论ID）',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：0-已取消，1-点赞中',
    `liketime` TIMESTAMP COMMENT '点赞时间',
    `unliketime` TIMESTAMP COMMENT '取消点赞时间',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_target_user` (`target_id`, `user_id`) COMMENT '确保一个用户对同一目标只能点赞一次',
    KEY `idx_type_target` (`type`, `target_id`) COMMENT '目标点赞查询',
    KEY `idx_user_id` (`user_id`) COMMENT '用户点赞查询',
    KEY `idx_user_liketime` (`user_id`, `liketime` DESC) COMMENT '用户点赞历史查询（按点赞时间排序）',
    KEY `idx_type_status` (`type`, `status`) COMMENT '按类型和状态查询',
    KEY `idx_status_liketime` (`status`, `liketime` DESC) COMMENT '状态+点赞时间复合索引'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='点赞表（支持帖子和评论）';