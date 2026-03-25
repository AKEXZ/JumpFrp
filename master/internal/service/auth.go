package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jumpfrp/master/config"
	"github.com/jumpfrp/master/internal/middleware"
	"github.com/jumpfrp/master/internal/model"
	"gorm.io/gorm"
)

type AuthService struct {
	db      *gorm.DB
	cfg     *config.Config
	mailSvc *MailService
}

func NewAuthService(db *gorm.DB, cfg *config.Config, sysSvc *SystemService) *AuthService {
	return &AuthService{db: db, cfg: cfg, mailSvc: NewMailService(sysSvc)}
}

type RegisterInput struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Code     string `json:"code" binding:"required"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required"`  // 支持邮箱或用户名
	Password string `json:"password" binding:"required"`
}

func (s *AuthService) Register(input RegisterInput) (*model.User, error) {
	// 检查用户名是否存在
	var count int64
	s.db.Model(&model.User{}).Where("username = ?", input.Username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否存在
	s.db.Model(&model.User{}).Where("email = ?", input.Email).Count(&count)
	if count > 0 {
		return nil, errors.New("邮箱已被注册")
	}

	// 验证邮箱验证码
	var user model.User
	result := s.db.Where("email = ? AND verify_code = ? AND verify_expire > ?",
		input.Email, input.Code, time.Now()).First(&user)
	if result.Error != nil {
		return nil, errors.New("验证码无效或已过期")
	}

	// 创建用户
	newUser := &model.User{
		Username:      input.Username,
		Email:         input.Email,
		VIPLevel:      model.VIPFree,
		Status:        model.UserStatusActive,
		EmailVerified: true,
		APIToken:      generateToken(32),
	}
	if err := newUser.SetPassword(input.Password); err != nil {
		return nil, err
	}

	if err := s.db.Create(newUser).Error; err != nil {
		return nil, err
	}

	// 清除验证码
	s.db.Model(&user).Updates(map[string]interface{}{
		"verify_code":   "",
		"verify_expire": nil,
	})

	return newUser, nil
}

func (s *AuthService) Login(input LoginInput) (string, *model.User, error) {
	var user model.User
	// 支持邮箱或用户名登录
	result := s.db.Where("email = ? OR username = ?", input.Email, input.Email).First(&user)
	if result.Error != nil {
		return "", nil, errors.New("账号或密码错误")
	}

	if user.Status == model.UserStatusBanned {
		return "", nil, errors.New("账号已被封禁")
	}

	if !user.CheckPassword(input.Password) {
		return "", nil, errors.New("邮箱或密码错误")
	}

	// 判断是否是管理员（username == "admin"）
	isAdmin := user.Username == "admin"

	// 生成 JWT
	expire := time.Now().Add(time.Duration(s.cfg.JWT.ExpireHours) * time.Hour)
	claims := &middleware.Claims{
		UserID:   user.ID,
		Username: user.Username,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expire),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return "", nil, err
	}

	return tokenStr, &user, nil
}

func (s *AuthService) SendVerifyCode(email string) error {
	// 生成 6 位验证码
	code := fmt.Sprintf("%06d", randomInt(999999))
	expire := time.Now().Add(5 * time.Minute)

	// 存储验证码（不管用户是否存在，都存一条临时记录）
	var user model.User
	result := s.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		// 用户不存在，创建临时记录
		user = model.User{
			Email:        email,
			VerifyCode:   code,
			VerifyExpire: &expire,
			Status:       "pending",
		}
		s.db.Create(&user)
	} else {
		s.db.Model(&user).Updates(map[string]interface{}{
			"verify_code":   code,
			"verify_expire": expire,
		})
	}

	// TODO: 发送邮件（Phase 2 实现 SMTP）
	fmt.Printf("[DEBUG] 验证码: %s -> %s\n", email, code)
	// 发送验证码邮件
	go s.mailSvc.SendVerifyCode(email, email, code)
	return nil
}

func generateToken(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func randomInt(max int) int {
	b := make([]byte, 4)
	rand.Read(b)
	n := int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
	if n < 0 {
		n = -n
	}
	return n % (max + 1)
}
