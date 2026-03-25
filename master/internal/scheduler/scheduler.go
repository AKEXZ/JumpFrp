package scheduler

import (
	"log"
	"time"

	"github.com/jumpfrp/master/config"
	"github.com/jumpfrp/master/internal/model"
	"github.com/jumpfrp/master/internal/service"
	"gorm.io/gorm"
)

type Scheduler struct {
	db      *gorm.DB
	cfg     *config.Config
	mailSvc *service.MailService
	stopCh  chan struct{}
}

func New(db *gorm.DB, cfg *config.Config, sysSvc *service.SystemService) *Scheduler {
	return &Scheduler{
		db:      db,
		cfg:     cfg,
		mailSvc: service.NewMailService(sysSvc),
		stopCh:  make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	go s.nodeOfflineChecker()
	go s.vipExpireChecker()
	log.Println("定时任务已启动")
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
}

// 每 60 秒检查节点是否离线（超过 90 秒无心跳则标记离线）
func (s *Scheduler) nodeOfflineChecker() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			threshold := time.Now().Add(-90 * time.Second)
			result := s.db.Model(&model.Node{}).
				Where("status = ? AND last_heartbeat < ?", model.NodeStatusOnline, threshold).
				Update("status", model.NodeStatusOffline)
			if result.RowsAffected > 0 {
				log.Printf("节点离线检测: %d 个节点已标记为离线", result.RowsAffected)
			}
		case <-s.stopCh:
			return
		}
	}
}

// 每天检查 VIP 到期（到期前 7/3/1 天发提醒邮件）
func (s *Scheduler) vipExpireChecker() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			vipNames := map[int]string{1: "Basic", 2: "Pro", 3: "Ultimate"}
			for _, days := range []int{7, 3, 1} {
				start := time.Now().AddDate(0, 0, days)
				end := start.Add(24 * time.Hour)
				var users []model.User
				s.db.Where("vip_level > 0 AND vip_expire_at BETWEEN ? AND ?", start, end).Find(&users)
				for _, u := range users {
					name := vipNames[u.VIPLevel]
					expireStr := u.VIPExpireAt.Format("2006-01-02 15:04")
					log.Printf("[VIP到期提醒] 用户 %s (%s) 将在 %d 天后到期", u.Username, name, days)
					go s.mailSvc.SendVIPExpiring(u.Email, u.Username, name, days, expireStr)
				}
			}
		case <-s.stopCh:
			return
		}
	}
}
