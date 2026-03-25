package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/jumpfrp/master/internal/model"
	"gorm.io/gorm"
)

type VIPService struct {
	db *gorm.DB
}

func NewVIPService(db *gorm.DB) *VIPService {
	return &VIPService{db: db}
}

// VIP 套餐定义
type VIPPlan struct {
	Level       int
	Name        string
	Price       float64
	DurationDays int
	Description string
}

var Plans = []VIPPlan{
	{
		Level:        model.VIPBasic,
		Name:         "Basic",
		Price:        9.9,
		DurationDays: 30,
		Description:  "5条隧道 / 10个端口 / 5Mbps / TCP+UDP",
	},
	{
		Level:        model.VIPPro,
		Name:         "Pro",
		Price:        29.9,
		DurationDays: 30,
		Description:  "20条隧道 / 50个端口 / 20Mbps / 全协议 / 子域名",
	},
	{
		Level:        model.VIPUltimate,
		Name:         "Ultimate",
		Price:        99.9,
		DurationDays: 30,
		Description:  "无限隧道 / 200个端口 / 100Mbps / 全协议 / 固定端口",
	},
}

// 获取所有套餐
func (s *VIPService) GetPlans() []VIPPlan {
	return Plans
}

// 获取用户当前 VIP 信息
func (s *VIPService) GetUserVIP(userID uint) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// 手动开通 VIP（管理员操作）
func (s *VIPService) AdminGrantVIP(userID uint, vipLevel int, durationDays int) error {
	if vipLevel < model.VIPBasic || vipLevel > model.VIPUltimate {
		return errors.New("VIP 等级无效")
	}
	if durationDays <= 0 {
		return errors.New("有效期必须大于 0")
	}

	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 计算新的过期时间
	var expireAt time.Time
	if user.VIPExpireAt != nil && user.VIPExpireAt.After(time.Now()) {
		// 已有有效 VIP，在现有基础上延期
		expireAt = user.VIPExpireAt.AddDate(0, 0, durationDays)
	} else {
		// 新开通或已过期，从现在开始计算
		expireAt = time.Now().AddDate(0, 0, durationDays)
	}

	// 更新用户 VIP
	if err := s.db.Model(&user).Updates(map[string]interface{}{
		"vip_level":     vipLevel,
		"vip_expire_at": expireAt,
	}).Error; err != nil {
		return err
	}

	// 创建订单记录
	order := &model.VIPOrder{
		UserID:       userID,
		VIPLevel:     vipLevel,
		DurationDays: durationDays,
		Price:        0, // 管理员手动开通，价格为 0
		Status:       "paid",
		ExpireAt:     &expireAt,
	}
	s.db.Create(order)

	return nil
}

// 用户自助购买 VIP（Phase 5 接入支付）
func (s *VIPService) UserBuyVIP(userID uint, vipLevel int, durationDays int) (*model.VIPOrder, error) {
	if vipLevel < model.VIPBasic || vipLevel > model.VIPUltimate {
		return nil, errors.New("VIP 等级无效")
	}

	// 查找对应套餐
	var plan *VIPPlan
	for i := range Plans {
		if Plans[i].Level == vipLevel {
			plan = &Plans[i]
			break
		}
	}
	if plan == nil {
		return nil, errors.New("套餐不存在")
	}

	// 计算价格（按天数比例）
	price := plan.Price * float64(durationDays) / float64(plan.DurationDays)

	// 创建订单（待支付）
	expireAt := time.Now().AddDate(0, 0, durationDays)
	order := &model.VIPOrder{
		UserID:       userID,
		VIPLevel:     vipLevel,
		DurationDays: durationDays,
		Price:        price,
		Status:       "pending",
		ExpireAt:     &expireAt,
	}

	if err := s.db.Create(order).Error; err != nil {
		return nil, err
	}

	return order, nil
}

// 订单支付成功回调（Phase 5 支付网关调用）
func (s *VIPService) OrderPaid(orderID uint) error {
	var order model.VIPOrder
	if err := s.db.First(&order, orderID).Error; err != nil {
		return errors.New("订单不存在")
	}

	if order.Status != "pending" {
		return errors.New("订单状态不允许支付")
	}

	// 更新订单状态
	s.db.Model(&order).Update("status", "paid")

	// 更新用户 VIP
	var user model.User
	s.db.First(&user, order.UserID)

	var expireAt time.Time
	if user.VIPExpireAt != nil && user.VIPExpireAt.After(time.Now()) {
		expireAt = user.VIPExpireAt.AddDate(0, 0, order.DurationDays)
	} else {
		expireAt = time.Now().AddDate(0, 0, order.DurationDays)
	}

	s.db.Model(&user).Updates(map[string]interface{}{
		"vip_level":     order.VIPLevel,
		"vip_expire_at": expireAt,
	})

	return nil
}

// 获取用户订单列表
func (s *VIPService) GetUserOrders(userID uint) ([]model.VIPOrder, error) {
	var orders []model.VIPOrder
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&orders).Error
	return orders, err
}

// 获取所有订单（管理员）
func (s *VIPService) GetAllOrders(status string) ([]model.VIPOrder, error) {
	var orders []model.VIPOrder
	query := s.db.Preload("User")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").Find(&orders).Error
	return orders, err
}

// 取消订单
func (s *VIPService) CancelOrder(orderID uint, userID uint) error {
	var order model.VIPOrder
	if err := s.db.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		return errors.New("订单不存在")
	}

	if order.Status != "pending" {
		return fmt.Errorf("只能取消待支付订单，当前状态: %s", order.Status)
	}

	return s.db.Model(&order).Update("status", "cancelled").Error
}
