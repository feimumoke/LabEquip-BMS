package web

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/defines/entity"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/datasource"
)

// AuthService 认证服务
type AuthService struct {
	ds datasource.DataSource
}

var authServiceInstance *AuthService

func init() {
	authServiceInstance = &AuthService{
		ds: datasource.DefaultBasicSource,
	}
}

// GetAuthService 获取认证服务实例
func GetAuthService() *AuthService {
	return authServiceInstance
}

// ValidateToken 验证 token 并返回用户信息
func (a *AuthService) ValidateToken(ctx context.Context, token string) (*entity.AccountCacheInfo, *bmserror.BMSError) {
	if token == "" {
		return nil, bmserror.NewError(constant.ErrParam, "token is empty")
	}

	// 查询未过期的 session
	loginSessionList := make([]*entity.LoginSession, 0)
	err := a.ds.GetDataSource(ctx, nil).
		Table(entity.LoginSessionTabTableName).
		Where("session_id = ?", token).
		//Where("expire_time >= ?", timeutil.GetCurrentUnix()).
		Find(&loginSessionList).GetError()

	if err != nil {
		return nil, bmserror.NewError(constant.ErrDB, err.Error())
	}

	if len(loginSessionList) == 0 {
		return nil, bmserror.NewError(constant.ErrAuth, "token expired or invalid")
	}

	// 解析用户信息
	accountInfo := &entity.AccountCacheInfo{}
	marshalErr := json.Unmarshal([]byte(loginSessionList[0].UserInfo), accountInfo)
	if marshalErr != nil {
		return nil, bmserror.NewError(constant.ErrInternalServer, marshalErr.Error())
	}

	return accountInfo, nil
}

// GetUserByID 根据 user_id 获取用户信息
func (a *AuthService) GetUserByID(ctx context.Context, userID string) (*entity.UserTab, *bmserror.BMSError) {
	user := &entity.UserTab{}
	err := a.ds.GetDataSource(ctx, nil).
		Table(entity.UserTabName).
		Where("user_id = ?", userID).
		First(user).GetError()

	if err != nil {
		return nil, bmserror.NewError(constant.ErrDB, err.Error())
	}

	return user, nil
}

// 白名单路径 - 这些路径不需要登录验证
var whiteListPaths = []string{
	"/apps/basic/user/user_login",  // 用户登录
	"/apps/basic/user/create_user", // 用户注册
	"/apps/common/enums",           // 获取枚举值
	"/health",                      // 健康检查
	"/ping",                        // Ping
}

// IsWhiteListPath 判断是否是白名单路径
func IsWhiteListPath(path string) bool {
	for _, whitePath := range whiteListPaths {
		// 使用 Contains 而不是 HasPrefix，因为路径可能包含前缀如 /api
		if strings.Contains(path, whitePath) {
			return true
		}
	}
	return false
}
