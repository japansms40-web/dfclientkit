// Package account 管理"账号队列驱动"型客户端项目共用的账号模型：CK/UA/IP 三元组、
// 状态与累计统计、导入文本解析、磁盘持久化。不依赖任何 UI 或具体发布协议。
package account

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// 账号状态取值；与前端展示的中文文案保持一致。
const (
	StatusPending = "待发"
	StatusRunning = "发布中"
	StatusSuccess = "成功"
	StatusFailed  = "失败"
)

// Account 是队列中的一个账号。json tag 同时决定磁盘持久化格式与前端绑定收到的字段名。
type Account struct {
	CK      string `json:"ck"`
	UA      string `json:"ua"`
	IP      string `json:"ip"`
	Status  string `json:"status"`
	Success int    `json:"success"`
	Fail    int    `json:"fail"`
	Total   int    `json:"total"`
	Bad     bool   `json:"bad"` // 手动标记为坏号，批量处理时跳过
}

// ParseImportText 把粘贴/拖入的文本按 "----" 分隔解析为账号列表；
// 每段去除首尾空白后作为一个账号的 CK，空段落忽略。
func ParseImportText(text string) []Account {
	parts := strings.Split(text, "----")
	out := make([]Account, 0, len(parts))
	for _, p := range parts {
		ck := strings.TrimSpace(p)
		if ck == "" {
			continue
		}
		out = append(out, Account{CK: ck, Status: StatusPending})
	}
	return out
}

// Load 从 path 指向的 JSON 文件读取账号列表；文件不存在或内容损坏时返回空列表。
// 返回值始终是非 nil 的切片：Go 的 nil 切片编码成 JSON 会变成 null，传到前端会让
// `accounts.value` 变成 null 而不是空数组，导致模板里任何 accounts.length /
// accounts.map 之类的访问直接抛异常，把整个页面渲染搞崩——这里必须显式兜底，
// 不能只判断 len() 是否为 0。
func Load(path string) []Account {
	list := []Account{}
	b, err := os.ReadFile(path)
	if err != nil {
		return list
	}
	_ = json.Unmarshal(b, &list)
	if list == nil {
		list = []Account{}
	}
	return list
}

// Save 把账号列表写入 path 指向的 JSON 文件，按需创建父目录。
func Save(path string, list []Account) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}
