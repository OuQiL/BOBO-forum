USE bobo_db;

CREATE TABLE `posts` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '帖子ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `community_id` BIGINT UNSIGNED NOT NULL COMMENT '社区ID',
    `username` VARCHAR(50) NOT NULL COMMENT '用户名',
    `title` VARCHAR(200) NOT NULL COMMENT '帖子标题',
    `content` TEXT NOT NULL COMMENT '帖子内容',
    `tags` VARCHAR(500) DEFAULT NULL COMMENT '标签，多个标签用逗号分隔',
    `like_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '点赞数',
    `comment_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '评论数',
    `view_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '浏览数',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_community_created` (`community_id`, `created_at` DESC) COMMENT '社区帖子列表查询',
    KEY `idx_user_id` (`user_id`) COMMENT '用户帖子查询',
    KEY `idx_created_at` (`created_at` DESC) COMMENT '最新帖子查询'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帖子表';

CREATE TABLE `comments` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '评论ID',
    `post_id` BIGINT UNSIGNED NOT NULL COMMENT '帖子ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `username` VARCHAR(50) NOT NULL COMMENT '用户名',
    `content` TEXT NOT NULL COMMENT '评论内容',
    `parent_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '父评论ID，0表示顶级评论',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_post_created` (`post_id`, `created_at` ASC) COMMENT '帖子评论列表查询',
    KEY `idx_user_id` (`user_id`) COMMENT '用户评论查询',
    KEY `idx_parent_id` (`parent_id`) COMMENT '子评论查询'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表';
