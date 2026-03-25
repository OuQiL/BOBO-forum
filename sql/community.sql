USE bobo_db;

CREATE TABLE `communities` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '社区ID',
    `name` VARCHAR(100) NOT NULL COMMENT '社区名称',
    `description` VARCHAR(500) DEFAULT '' COMMENT '社区详情',
    `post_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '帖子数目',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 0-禁用 1-启用',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name` (`name`) COMMENT '社区名称唯一索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='社区表';

INSERT INTO `communities` (`name`, `description`, `post_count`, `status`) VALUES
('BOBO广场', '社区公告、活动通知、新手报到', 0, 1),
('代码工坊', '踩坑记录、最佳实践、架构探讨', 0, 1),
('数码前线', '新品资讯、开箱评测、选购建议', 0, 1),
('影视天地', '追剧打卡、电影推荐、观后感', 0, 1),
('运动打卡', '健身日志、跑步骑行、运动装备', 0, 1);
