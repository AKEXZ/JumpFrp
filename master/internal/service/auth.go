package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
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

	// 验证邮箱验证码（从 system_configs 表读取）
	// 格式: "code|expire_timestamp"
	key := "verify_code_" + input.Email
	var cfg model.SystemConfig
	result := s.db.Where("key = ?", key).First(&cfg)
	if result.Error != nil {
		return nil, errors.New("验证码无效或已过期")
	}
	
	// 解析验证码和过期时间
	parts := strings.Split(cfg.Value, "|")
	if len(parts) != 2 || parts[0] != input.Code {
		return nil, errors.New("验证码无效或已过期")
	}
	
	// 检查是否过期
	expireTs, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || time.Now().Unix() > expireTs {
		return nil, errors.New("验证码已过期")
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
	s.db.Where("key = ?", key).Delete(&model.SystemConfig{})

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

	// 存储验证码到 system_configs 表，格式: "code|expire_timestamp"
	// 不创建用户，避免测试邮件误创建用户
	key := "verify_code_" + email
	value := fmt.Sprintf("%s|%d", code, expire.Unix())
	
	var cfg model.SystemConfig
	result := s.db.Where("key = ?", key).First(&cfg)
	if result.Error != nil {
		// 创建新记录
		s.db.Create(&model.SystemConfig{
			Key:   key,
			Value: value,
		})
	} else {
		// 更新已有记录
		s.db.Model(&cfg).Update("value", value)
	}

	fmt.Printf("[DEBUG] 验证码: %s -> %s (5分钟有效)\n", email, code)
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
