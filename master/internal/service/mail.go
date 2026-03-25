package service

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"strconv"
)

type MailService struct {
	sysSvc *SystemService
}

func NewMailService(sysSvc *SystemService) *MailService {
	return &MailService{sysSvc: sysSvc}
}

type MailData struct {
	Username   string
	Content    template.HTML
	ActionURL  string
	ActionText string
	Year       int
}

const mailTpl = `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><style>
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;background:#f5f7fa;margin:0;padding:0}
.wrap{max-width:560px;margin:40px auto;background:white;border-radius:12px;overflow:hidden;box-shadow:0 4px 20px rgba(0,0,0,.08)}
.header{background:linear-gradient(135deg,#1a1a2e,#0f3460);padding:32px;text-align:center}
.header h1{color:#409eff;margin:0;font-size:24px}
.header p{color:#aaa;margin:8px 0 0;font-size:14px}
.body{padding:32px}
.body p{color:#444;line-height:1.8;margin:0 0 16px}
.btn{display:inline-block;background:#409eff;color:white;padding:12px 32px;border-radius:8px;text-decoration:none;font-size:15px;margin:8px 0}
.footer{background:#f5f7fa;padding:16px 32px;text-align:center;color:#999;font-size:12px}
.code{font-size:36px;font-weight:bold;color:#409eff;letter-spacing:8px;text-align:center;padding:16px;background:#f0f7ff;border-radius:8px;margin:16px 0}
</style></head>
<body>
<div class="wrap">
  <div class="header">
    <h1>⚡ JumpFrp</h1>
    <p>高速内网穿透服务</p>
  </div>
  <div class="body">
    <p>Hi <strong>{{.Username}}</strong>，</p>
    {{.Content}}
    {{if .ActionURL}}<p><a class="btn" href="{{.ActionURL}}">{{.ActionText}}</a></p>{{end}}
  </div>
  <div class="footer">© {{.Year}} JumpFrp · jumpfrp.top · 如非本人操作请忽略</div>
</div>
</body></html>`

func (m *MailService) Send(to, subject string, data MailData) error {
	cfg := m.sysSvc.GetSMTPConfig()

	if !cfg.Enabled || cfg.Host == "" {
		// SMTP 未配置或未启用，仅打印日志
		fmt.Printf("[MAIL] To: %s | Subject: %s\n", to, subject)
		return nil
	}

	// 渲染 HTML
	tpl := template.Must(template.New("mail").Parse(mailTpl))
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return err
	}

	msg := "From: JumpFrp <" + cfg.From + ">\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		buf.String()

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	addr := cfg.Host + ":" + strconv.Itoa(cfg.Port)

	if cfg.SSL {
		tlsCfg := &tls.Config{ServerName: cfg.Host}
		conn, err := tls.Dial("tcp", addr, tlsCfg)
		if err != nil {
			return smtp.SendMail(addr, auth, cfg.From, []string{to}, []byte(msg))
		}
		defer conn.Close()
		client, err := smtp.NewClient(conn, cfg.Host)
		if err != nil {
			return err
		}
		defer client.Close()
		if err = client.Auth(auth); err != nil {
			return err
		}
		if err = client.Mail(cfg.From); err != nil {
			return err
		}
		if err = client.Rcpt(to); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(msg))
		if err != nil {
			return err
		}
		return w.Close()
	}

	return smtp.SendMail(addr, auth, cfg.From, []string{to}, []byte(msg))
}

// 发送验证码邮件
func (m *MailService) SendVerifyCode(to, username, code string) error {
	content := fmt.Sprintf(`
<p>您正在注册 JumpFrp 账号，验证码为：</p>
<div class="code">%s</div>
<p style="color:#999;font-size:13px">验证码 5 分钟内有效，请勿泄露给他人。</p>`, code)

	return m.Send(to, "【JumpFrp】邮箱验证码", MailData{
		Username: username,
		Content:  template.HTML(content),
	})
}

// 发送重置密码邮件
func (m *MailService) SendResetPassword(to, username, resetURL string) error {
	content := `<p>您申请了重置密码，点击下方按钮完成操作：</p>
<p style="color:#999;font-size:13px">链接 30 分钟内有效，如非本人操作请忽略。</p>`

	return m.Send(to, "【JumpFrp】重置密码", MailData{
		Username:   username,
		Content:    template.HTML(content),
		ActionURL:  resetURL,
		ActionText: "重置密码",
	})
}

// 发送 VIP 开通成功邮件
func (m *MailService) SendVIPGranted(to, username, vipName, expireAt string) error {
	content := fmt.Sprintf(`
<p>您的 VIP 已成功开通，详情如下：</p>
<p>套餐：<strong>%s</strong></p>
<p>到期时间：<strong>%s</strong></p>
<p>感谢您对 JumpFrp 的支持！</p>`, vipName, expireAt)

	return m.Send(to, "【JumpFrp】VIP 开通成功", MailData{
		Username:   username,
		Content:    template.HTML(content),
		ActionURL:  "https://jumpfrp.top/dashboard",
		ActionText: "进入控制台",
	})
}

// 发送 VIP 到期提醒邮件
func (m *MailService) SendVIPExpiring(to, username, vipName string, daysLeft int, expireAt string) error {
	content := fmt.Sprintf(`
<p>您的 VIP 套餐即将到期，请及时续费：</p>
<p>套餐：<strong>%s</strong></p>
<p>到期时间：<strong>%s</strong>（还剩 <strong style="color:#f05a28">%d 天</strong>）</p>
<p>到期后将自动降级为免费用户，隧道数量和功能将受到限制。</p>`, vipName, expireAt, daysLeft)

	return m.Send(to, fmt.Sprintf("【JumpFrp】VIP 将在 %d 天后到期", daysLeft), MailData{
		Username:   username,
		Content:    template.HTML(content),
		ActionURL:  "https://jumpfrp.top/vip",
		ActionText: "立即续费",
	})
}
