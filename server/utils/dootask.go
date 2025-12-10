package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	dootask "github.com/dootask/tools/server/go"
)

type DooTaskClient struct {
	Client *dootask.Client
}

// NewDooTaskClient 创建 DooTask 客户端
func NewDooTaskClient(token string) DooTaskClient {
	if token == "" {
		return DooTaskClient{Client: nil}
	}

	return DooTaskClient{
		Client: dootask.NewClient(token),
	}
}

// SendBotMessage 发送 DooTask 机器人通知
func (d *DooTaskClient) SendBotMessage(userID *uint, message string) error {
	if userID == nil || *userID == 0 {
		return errors.New("userID is required")
	}

	if d.Client == nil {
		return errors.New("DooTask client is not initialized")
	}

	req := dootask.SendBotMessageRequest{
		UserID:  int(*userID),
		Text:    message,
		BotType: "system-msg",
	}

	return d.Client.SendBotMessage(req)
}

// BuildKPIAppConfig 返回 KPI 微应用配置的 JSON 字符串（供消息中的 micro-app 链接使用）
// BuildKPIAppConfig 返回 KPI 微应用配置的 JSON 字符串（带考核ID的详情入口）
func BuildKPIAppConfig(evaluationID uint) string {
	q := url.Values{}
	if evaluationID > 0 {
		q.Set("evaluation_id", fmt.Sprintf("%d", evaluationID))
	}
	return buildKPIAppConfig("kpi-details", "/apps/kpi/evaluations", q)
}

// BuildKPIInvitationAppConfig 返回 KPI 邀请评分入口的微应用配置
func BuildKPIInvitationAppConfig(invitationID uint, evaluationID uint) string {
	q := url.Values{}
	if invitationID > 0 {
		q.Set("invitation_id", fmt.Sprintf("%d", invitationID))
	}
	if evaluationID > 0 {
		q.Set("evaluation_id", fmt.Sprintf("%d", evaluationID))
	}
	return buildKPIAppConfig("kpi-invitations", "/apps/kpi/invitations", q)
}

// buildKPIAppConfig 构造 KPI 微应用配置的 JSON 字符串
func buildKPIAppConfig(name, basePath string, query url.Values) string {
	fullURL := basePath + "?theme={system_theme}"
	if encoded := query.Encode(); encoded != "" {
		fullURL = fullURL + "&" + encoded
	}

	cfg := map[string]interface{}{
		"id":         "kpi",
		"name":       name,
		"immersive":  true,
		"keep_alive": false,
		"url_type":   "iframe",
		"url":        fullURL,
	}

	b, err := json.Marshal(cfg)
	if err != nil {
		return ""
	}
	return string(b)
}
