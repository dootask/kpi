---
id: kpi.plugin.concept
title: KPI 插件元信息
type: concept
feature: kpi
scope: end-user
locale: zh
aliases:
  - KPI 插件
  - kpi 怎么装
  - 绩效插件
  - 绩效考核插件
  - kpi 应用 id
related_tools: []
related_pages: [application]
prerequisites: []
negative:
  - KPI 不是主程序内置功能，未装插件时入口不会出现
  - 插件升级不通过 git pull，需要在应用市场更新
  - 主程序版本必须高于 1.4.67，否则插件无法安装
last_verified: v1.7.90
---

# KPI 插件元信息

## 定义
KPI 绩效考核在 DooTask 中由社区插件提供，应用市场 app id 为 `kpi`，feature 短名 `kpi`。主程序不内置任何 KPI 代码，所有绩效逻辑都跑在独立 Docker 容器中，作为应用插件挂载到 DooTask 界面。版本号随应用市场发布的版本走，不固定。

## 关键属性
- **作者**：DooTask 官方
- **要求**：主程序版本 > 1.4.67（依赖新 API 能力）
- **运行形态**：单个 Docker 容器（镜像 `dootask/kpi:<version>`）
- **数据存储**：独立 SQLite 数据库，本地卷 `kpi_data` 挂载到 `/web/db`，不入主库
- **菜单注入**：安装后在「应用中心」注册「绩效考核」入口
- **重启策略**：`unless-stopped`，主机重启容器自动恢复

## 账号创建与角色分配（首次登录时）
KPI 账号在用户**首次进入插件（用 DooTask 令牌登录）时**自动建立并分配角色：

- 部门不存在则自动创建该部门
- DooTask 管理员 → KPI 内部 hr 角色
- 部门负责人 → KPI 内部 manager 角色
- 其余 → employee 角色

## 用户生命周期钩子
插件订阅了 DooTask 的用户事件，但只做在职状态同步，不在此建号：

- `user_onboard`：DooTask 重新启用某用户时，把 KPI 中**已存在**的对应账号标记为在职；KPI 中尚无此人时不处理（建号仍发生在首次登录）
- `user_offboard`：DooTask 删除 / 离职用户时，KPI 同步把对应账号标记为离职

## 信息同步规则
用户每次登录 KPI 时：

- 自动更新姓名、职位
- **角色保持不变**（不会因为 DooTask 角色变化而重新分配 KPI 角色）

## 不支持
- 不能离线安装到不联网的环境（需访问应用市场镜像源）
- 不能在主程序 < 1.4.67 的环境上安装

## 相关
- 是什么：[[kpi.concept]]
- 入口在哪：[[kpi.entry.menu-map]]
- 评分机制：[[kpi.scoring.concept]]
