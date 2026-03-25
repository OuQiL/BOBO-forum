USE bobo_db;

CREATE TABLE `posts` (
    `id` BIGINT UNSIGNED NOT NULL COMMENT '帖子ID (雪花算法)',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `community_id` BIGINT UNSIGNED NOT NULL COMMENT '社区ID',
    `username` VARCHAR(50) NOT NULL COMMENT '用户名',
    `title` VARCHAR(200) NOT NULL COMMENT '帖子标题',
    `content` TEXT NOT NULL COMMENT '帖子内容',
    `tags` VARCHAR(500) DEFAULT NULL COMMENT '标签，多个标签用逗号分隔',
    `like_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '点赞数',
    `comment_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '评论数',
    `view_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '浏览数',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0-草稿 1-审核中 2-已发布 3-已删除',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_community_created` (`community_id`, `created_at` DESC) COMMENT '社区帖子列表查询',
    KEY `idx_user_id` (`user_id`) COMMENT '用户帖子查询',
    KEY `idx_created_at` (`created_at` DESC) COMMENT '最新帖子查询'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帖子表';

CREATE TABLE `comments` (
    `id` BIGINT UNSIGNED NOT NULL COMMENT '评论ID (雪花算法)',
    `post_id` BIGINT UNSIGNED NOT NULL COMMENT '帖子ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `username` VARCHAR(50) NOT NULL COMMENT '用户名',
    `content` TEXT NOT NULL COMMENT '评论内容',
    `parent_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '父评论ID，0表示一级评论，非0表示二级评论',
    `reply_to_user_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '回复目标用户ID，二级评论时使用',
    `reply_to_username` VARCHAR(50) DEFAULT '' COMMENT '回复目标用户名，二级评论时使用',
    `like_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '点赞数',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0-正常  1-已删除',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_post_created` (`post_id`, `created_at` ASC) COMMENT '帖子评论最新排序',
    KEY `idx_post_like` (`post_id`, `like_count` DESC) COMMENT '帖子评论最热排序',
    KEY `idx_user_id` (`user_id`) COMMENT '用户评论查询',
    KEY `idx_parent_id` (`parent_id`, `created_at` ASC) COMMENT '一级评论下的回复列表'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表（两级结构）';
