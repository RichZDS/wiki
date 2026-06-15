-- 用户表
CREATE TABLE IF NOT EXISTS `user` (
    `id`         BIGINT          NOT NULL                COMMENT '主键（雪花算法生成）',
    `name`       VARCHAR(64)     NOT NULL                COMMENT '用户名',
    `password`   VARCHAR(256)    NOT NULL                COMMENT '密码（bcrypt 哈希）',
    `quota`      BIGINT          NOT NULL DEFAULT 0      COMMENT '配额，单位字节',
    `remark`     VARCHAR(512)    NOT NULL DEFAULT ''     COMMENT '备注',
    `is_deleted` TINYINT(1)      NOT NULL DEFAULT 0      COMMENT '软删除标记 0-正常 1-已删除',
    `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';


-- AI 模型状态表
CREATE TABLE IF NOT EXISTS `ai_model` (
    `id`          BIGINT       NOT NULL COMMENT '主键（雪花算法生成）',
    `model_name`  VARCHAR(128) NOT NULL COMMENT '模型名称',
    `model_id`    VARCHAR(128) NOT NULL DEFAULT '' COMMENT '模型 ID（API 调用用）',
    `api_key`     VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'API 密钥，为空时回退环境变量',
    `is_used`     TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '是否可用：0-不可用 1-可用',
    `fail_reason` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '不可用原因',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_model_name` (`model_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 模型状态表';
