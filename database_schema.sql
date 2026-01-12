-- =============================================
-- 实验设备借记管理系统 - 数据库表结构
-- LabEquip-BMS Database Schema
-- =============================================

CREATE
DATABASE IF NOT EXISTS bms_basic_db
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

use
bms_basic_db;

-- 用户表
DROP TABLE IF EXISTS `user_tab`;
CREATE TABLE `user_tab`
(
    `id`        BIGINT(20)   NOT NULL AUTO_INCREMENT COMMENT '主键',
    `user_id`   VARCHAR(64)  NOT NULL COMMENT '用户ID',
    `email`     VARCHAR(128) NOT NULL COMMENT '邮箱',
    `name`      VARCHAR(64)  NOT NULL COMMENT '姓名',
    `avatar`    VARCHAR(255) DEFAULT '' COMMENT '头像URL',
    `phone`     VARCHAR(32)  DEFAULT '' COMMENT '手机号',
    `status`    BIGINT(20)   DEFAULT 1 COMMENT '状态: 1-正常 2-禁用',
    `passwd`    VARCHAR(128) NOT NULL COMMENT '密码(MD5加密)',
    `salt`      VARCHAR(64)  NOT NULL COMMENT '密码盐值',
    `gender`    BIGINT(20)   DEFAULT 0 COMMENT '性别: 0-未知 1-男 2-女',
    `introduce` TEXT COMMENT '个人介绍',
    `role`      BIGINT(20)   NOT NULL COMMENT '角色: 1-超级管理员 2-管理员 3-教师 4-学生',
    `ctime`     BIGINT(20)   NOT NULL COMMENT '创建时间(时间戳)',
    `mtime`     BIGINT(20)   NOT NULL COMMENT '修改时间(时间戳)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_id` (`user_id`),
    UNIQUE KEY `uk_email` (`email`),
    KEY         `idx_phone` (`phone`),
    KEY         `idx_role` (`role`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='用户表';

-- 用户登录记录表
DROP TABLE IF EXISTS `user_login_tab`;
CREATE TABLE `user_login_tab`
(
    `id`          BIGINT(20)  NOT NULL AUTO_INCREMENT COMMENT '主键',
    `user_id`     VARCHAR(64) NOT NULL COMMENT '用户ID',
    `login_time`  BIGINT(20)  NOT NULL COMMENT '登录时间(时间戳)',
    `expire_time` BIGINT(20)  NOT NULL COMMENT '过期时间(时间戳)',
    `login_ip`    VARCHAR(64) DEFAULT '' COMMENT '登录IP',
    `login_type`  BIGINT(20)  DEFAULT 1 COMMENT '登录类型: 1-Web 2-Mobile',
    `result`      BIGINT(20)  DEFAULT 1 COMMENT '登录结果: 0-失败 1-成功',
    PRIMARY KEY (`id`),
    KEY           `idx_user_id` (`user_id`),
    KEY           `idx_login_time` (`login_time`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='用户登录记录表';

-- 登录Session表
DROP TABLE IF EXISTS `login_session_tab`;
CREATE TABLE `login_session_tab`
(
    `id`          BIGINT(20)   NOT NULL AUTO_INCREMENT COMMENT '主键',
    `user_id`     VARCHAR(64)  NOT NULL COMMENT '用户ID',
    `session_id`  VARCHAR(128) NOT NULL COMMENT 'Session ID',
    `user_info`   TEXT         NOT NULL COMMENT '用户信息JSON',
    `expire_time` BIGINT(20)   NOT NULL COMMENT '过期时间(时间戳)',
    `ctime`       BIGINT(20)   NOT NULL COMMENT '创建时间(时间戳)',
    `mtime`       BIGINT(20)   NOT NULL COMMENT '修改时间(时间戳)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_session_id` (`session_id`),
    KEY           `idx_user_id` (`user_id`),
    KEY           `idx_expire_time` (`expire_time`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='登录Session表';

-- =============================================
-- 2. 基础信息表
-- =============================================

-- 实验室表
DROP TABLE IF EXISTS `lab_tab`;
CREATE TABLE `lab_tab`
(
    `id`          BIGINT(20)   NOT NULL AUTO_INCREMENT COMMENT '主键',
    `lab_code`    VARCHAR(64)  NOT NULL COMMENT '实验室编码',
    `lab_name`    VARCHAR(128) NOT NULL COMMENT '实验室名称',
    `address`     VARCHAR(255) DEFAULT '' COMMENT '地址',
    `manager_id`  BIGINT(20)   DEFAULT 0 COMMENT '管理员ID',
    `description` TEXT COMMENT '描述',
    `ctime`       BIGINT(20)   NOT NULL COMMENT '创建时间(时间戳)',
    `mtime`       BIGINT(20)   NOT NULL COMMENT '修改时间(时间戳)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_lab_code` (`lab_code`),
    KEY           `idx_manager_id` (`manager_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='实验室表';

-- 设备表
DROP TABLE IF EXISTS `equip_tab`;
CREATE TABLE `equip_tab`
(
    `id`            BIGINT(20)   NOT NULL AUTO_INCREMENT COMMENT '主键',
    `equip_id`      VARCHAR(64)  NOT NULL COMMENT '设备ID',
    `category_id`   BIGINT(20)   NOT NULL COMMENT '分类ID: 1-通用 2-化学 3-生物 4-物理 5-电子 6-计算机 7-安全 8-特殊',
    `category_name` VARCHAR(64)  NOT NULL COMMENT '分类名称',
    `equip_name`    VARCHAR(128) NOT NULL COMMENT '设备名称',
    `model`         VARCHAR(128) DEFAULT '' COMMENT '规格型号',
    `creator`       VARCHAR(64)  DEFAULT '' COMMENT '创建人',
    `description`   TEXT COMMENT '描述',
    `ctime`         BIGINT(20)   NOT NULL COMMENT '创建时间(时间戳)',
    `mtime`         BIGINT(20)   NOT NULL COMMENT '修改时间(时间戳)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_equip_id` (`equip_id`),
    KEY             `idx_category_id` (`category_id`),
    KEY             `idx_equip_name` (`equip_name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='设备表';

CREATE TABLE `distributed_idcreator_tab`
(
    `id`          bigint(20) unsigned                      NOT NULL AUTO_INCREMENT,
    `id_type`     int(10) unsigned                         NOT NULL,
    `date_period` bigint(20) unsigned                      NOT NULL,
    `pt_no`       int(10) unsigned                         NOT NULL,
    `id_value`    bigint(20) unsigned                      NOT NULL,
    `description` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL,
    `mtime`       int(10) unsigned                         NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id_type_date_period_pt_no` (`id_type`, `date_period`, `pt_no`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

-- 插入默认实验室
INSERT INTO `lab_tab` (`lab_code`, `lab_name`, `address`, `manager_id`, `description`, `ctime`, `mtime`)
VALUES ('LAB001', '物理实验室', '教学楼A座3楼', 0, '主要用于物理实验教学', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
       ('LAB002', '化学实验室', '教学楼B座2楼', 0, '主要用于化学实验教学', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
       ('LAB003', '生物实验室', '教学楼C座1楼', 0, '主要用于生物实验教学', UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 插入默认管理员用户
-- 密码: admin123 (需要后端加盐加密后再插入)
-- 注意: 实际使用时需要通过注册接口创建用户，这里仅作示例
INSERT INTO `user_tab` (`user_id`, `email`, `name`, `phone`, `status`, `passwd`, `salt`, `role`, `ctime`, `mtime`)
VALUES ('admin_001', 'admin@example.com', '系统管理员', '13800138000', 1, 'to_be_encrypted', 'random_salt', 1,
        UNIX_TIMESTAMP(), UNIX_TIMESTAMP());


CREATE
DATABASE IF NOT EXISTS bms_inv_db
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

use
bms_inv_db;

-- =============================================
-- 3. 库存相关表
-- =============================================

-- 库存表
DROP TABLE IF EXISTS `inventory_tab`;
CREATE TABLE `inventory_tab`
(
    `id`            BIGINT(20)  NOT NULL AUTO_INCREMENT COMMENT '主键',
    `lab_id`        VARCHAR(64) NOT NULL COMMENT '实验室ID',
    `equip_id`      VARCHAR(64) NOT NULL COMMENT '设备ID',
    `total_qty`     BIGINT(20)  NOT NULL DEFAULT 0 COMMENT '总库存数量',
    `on_hand_qty`   BIGINT(20)  NOT NULL DEFAULT 0 COMMENT '在手库存 = total - borrowed',
    `available_qty` BIGINT(20)  NOT NULL DEFAULT 0 COMMENT '可用库存 = total - borrowed - allocated',
    `borrowed_qty`  BIGINT(20)  NOT NULL DEFAULT 0 COMMENT '已借出数量(已拿走)',
    `allocated_qty` BIGINT(20)  NOT NULL DEFAULT 0 COMMENT '已分配数量(未拿走)',
    `operator`      VARCHAR(64) DEFAULT '' COMMENT '操作人',
    `ctime`         BIGINT(20)  NOT NULL COMMENT '创建时间(时间戳)',
    `mtime`         BIGINT(20)  NOT NULL COMMENT '修改时间(时间戳)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_lab_equip` (`lab_id`, `equip_id`),
    KEY             `idx_equip_id` (`equip_id`),
    KEY             `idx_available_qty` (`available_qty`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='库存表';

-- 库存任务表
DROP TABLE IF EXISTS `inventory_task_tab`;
CREATE TABLE `inventory_task_tab`
(
    `id`          BIGINT(20)  NOT NULL AUTO_INCREMENT COMMENT '主键',
    `task_id`     VARCHAR(64) NOT NULL COMMENT '任务ID',
    `task_type`   BIGINT(20)  NOT NULL COMMENT '任务类型: 1-增加 2-减少',
    `task_status` BIGINT(20)  NOT NULL COMMENT '任务状态: 1-进行中 2-已完成',
    `lab_id`      VARCHAR(64) NOT NULL COMMENT '实验室ID',
    `equip_id`    VARCHAR(64) NOT NULL COMMENT '设备ID',
    `total_qty`   BIGINT(20)  NOT NULL COMMENT '操作数量',
    `operator`    VARCHAR(64) NOT NULL COMMENT '操作人',
    `remark`      TEXT COMMENT '备注',
    `ctime`       BIGINT(20)  NOT NULL COMMENT '创建时间(时间戳)',
    `mtime`       BIGINT(20)  NOT NULL COMMENT '修改时间(时间戳)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_task_id` (`task_id`),
    KEY           `idx_lab_equip` (`lab_id`, `equip_id`),
    KEY           `idx_task_status` (`task_status`),
    KEY           `idx_operator` (`operator`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='库存任务表';

-- 库存任务日志表
DROP TABLE IF EXISTS `inventory_task_log_tab`;
CREATE TABLE `inventory_task_log_tab`
(
    `id`          BIGINT(20)  NOT NULL AUTO_INCREMENT COMMENT '主键',
    `task_id`     VARCHAR(64) NOT NULL COMMENT '任务ID',
    `task_status` BIGINT(20)  NOT NULL COMMENT '任务状态: 1-进行中 2-已完成',
    `remark`      TEXT COMMENT '备注',
    `operator`    VARCHAR(64) NOT NULL COMMENT '操作人',
    PRIMARY KEY (`id`),
    KEY           `idx_task_id` (`task_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='库存任务日志表';


-- =============================================
-- 5. 交易日志表(三级账)
-- =============================================

-- 交易日志表(三级账)
DROP TABLE IF EXISTS `transaction_log_tab`;
CREATE TABLE `transaction_log_tab`
(
    `id`             BIGINT(20)  NOT NULL AUTO_INCREMENT COMMENT '主键',
    `transaction_id` VARCHAR(64) NOT NULL COMMENT '交易ID',
    `sheet_id`       VARCHAR(64) NOT NULL COMMENT '单据ID(任务ID)',
    `equip_id`       VARCHAR(64) NOT NULL COMMENT '设备ID',
    `lab_id`         VARCHAR(64) NOT NULL COMMENT '实验室ID',
    `operator`       VARCHAR(64) NOT NULL COMMENT '操作人',
    `remark`         TEXT COMMENT '备注',
    `total_qty`      BIGINT(20)  NOT NULL COMMENT '总库存',
    `on_hand_qty`    BIGINT(20)  NOT NULL COMMENT '在手库存',
    `available_qty`  BIGINT(20)  NOT NULL COMMENT '可用库存',
    `borrowed_qty`   BIGINT(20)  NOT NULL COMMENT '已借出数量',
    `allocated_qty`  BIGINT(20)  NOT NULL COMMENT '已分配数量',
    `op_qty`         BIGINT(20)  NOT NULL COMMENT '操作数量',
    `trans_type`     BIGINT(20)  NOT NULL COMMENT '交易类型: 1-增加 2-减少 3-分配 4-借出 5-归还 6-拒绝',
    `sheet_type`     BIGINT(20)  NOT NULL COMMENT '单据类型: 1-库存单 2-借记单',
    `ctime`          BIGINT(20)  NOT NULL COMMENT '创建时间(时间戳)',
    PRIMARY KEY (`id`),
    KEY              `idx_transaction_id` (`transaction_id`),
    KEY              `idx_sheet_id` (`sheet_id`),
    KEY              `idx_equip_lab` (`equip_id`, `lab_id`),
    KEY              `idx_trans_type` (`trans_type`),
    KEY              `idx_sheet_type` (`sheet_type`),
    KEY              `idx_ctime` (`ctime`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='交易日志表(三级账)';


CREATE
DATABASE IF NOT EXISTS bms_task_db
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

use
bms_task_db;

-- =============================================
-- 4. 借记相关表
-- =============================================

-- 借记任务表
DROP TABLE IF EXISTS `borrow_task_tab`;
CREATE TABLE `borrow_task_tab`
(
    `id`          BIGINT(20)  NOT NULL AUTO_INCREMENT COMMENT '主键',
    `task_id`     VARCHAR(64) NOT NULL COMMENT '任务ID',
    `equip_id`    VARCHAR(64) NOT NULL COMMENT '设备ID',
    `lab_id`      VARCHAR(64) NOT NULL COMMENT '实验室ID',
    `borrow_qty`  BIGINT(20)  NOT NULL COMMENT '借记数量',
    `task_status` BIGINT(20)  NOT NULL COMMENT '任务状态: 1-待分配 2-已分配 3-已审批 4-进行中 7-已拒绝 8-已归还 9-已取消',
    `creator`     VARCHAR(64) NOT NULL COMMENT '创建人',
    `approval`    VARCHAR(64) DEFAULT '' COMMENT '审批人',
    `ctime`       BIGINT(20)  NOT NULL COMMENT '创建时间(时间戳)',
    `mtime`       BIGINT(20)  NOT NULL COMMENT '修改时间(时间戳)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_task_id` (`task_id`),
    KEY           `idx_equip_lab` (`equip_id`, `lab_id`),
    KEY           `idx_task_status` (`task_status`),
    KEY           `idx_creator` (`creator`),
    KEY           `idx_approval` (`approval`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='借记任务表';

-- 借记任务日志表
DROP TABLE IF EXISTS `borrow_task_log_tab`;
CREATE TABLE `borrow_task_log_tab`
(
    `id`          BIGINT(20)  NOT NULL AUTO_INCREMENT COMMENT '主键',
    `task_id`     VARCHAR(64) NOT NULL COMMENT '任务ID',
    `task_status` BIGINT(20)  NOT NULL COMMENT '任务状态: 1-待分配 2-已分配 3-已审批 4-进行中 7-已拒绝 8-已归还 9-已取消',
    `remark`      TEXT COMMENT '备注',
    `operator`    VARCHAR(64) NOT NULL COMMENT '操作人',
    `ctime`       BIGINT(20)  NOT NULL COMMENT '创建时间(时间戳)',
    PRIMARY KEY (`id`),
    KEY           `idx_task_id` (`task_id`),
    KEY           `idx_operator` (`operator`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='借记任务日志表';

CREATE TABLE `process_message_v2_tab`
(
    `id`           bigint(20) unsigned                     NOT NULL AUTO_INCREMENT,
    `message_uuid` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
    `task_name`    varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
    `message`      mediumtext COLLATE utf8mb4_unicode_ci,
    `ctime`        int(10) unsigned                        NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_message_uuid` (`message_uuid`),
    KEY            `idx_task_name` (`task_name`),
    KEY            `idx_ctime` (`ctime`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;


CREATE TABLE `process_message_consumer_v2_tab`
(
    `id`             bigint(20) unsigned                     NOT NULL AUTO_INCREMENT,
    `message_uuid`   varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
    `task_name`      varchar(128) COLLATE utf8mb4_unicode_ci          DEFAULT '',
    `handler_name`   varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
    `handler_status` tinyint(4)                              NOT NULL DEFAULT '0',
    `handler_times`  tinyint(4)                              NOT NULL DEFAULT '0',
    `ctime`          int(10) unsigned                        NOT NULL DEFAULT '0',
    `mtime`          int(10) unsigned                        NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    KEY              `idx_message_uuid` (`message_uuid`),
    KEY              `idx_handler_name` (`handler_name`),
    KEY              `idx_task_name` (`task_name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;





