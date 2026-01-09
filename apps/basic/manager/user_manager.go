package manager

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/defines/entity"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/datasource"
	"github.com/feimumoke/labequipbms/framework/support/paginator"
	"github.com/feimumoke/labequipbms/framework/support/timeutil"
	"github.com/feimumoke/labequipbms/framework/support/util"
)

type UserManager struct {
	ds datasource.DataSource
}

func NewUserManager() *UserManager {
	return &UserManager{ds: datasource.DefaultBasicSource}
}
func (u *UserManager) GetUserRoleIdListMng(ctx context.Context, userID string) ([]string, *bmserror.BMSError) {
	return nil, nil
}

func (u *UserManager) IsUserHasPermission(ctx context.Context, userID, path string) bool {

	return true
}

func GenerateUserSessionCacheKey(userID, skey string) string {
	prefix := ":1"
	oldCacheKeyFormat := "%s:%s:%s"
	if len(skey) < 12 {
		return fmt.Sprintf(oldCacheKeyFormat, prefix, userID, skey)
	}
	return fmt.Sprintf(oldCacheKeyFormat, prefix, userID, skey[:12])
}

func (u *UserManager) GetUserInfo(ctx context.Context, userID, skey string) (*entity.AccountCacheInfo, *bmserror.BMSError) {
	sessionID := GenerateUserSessionCacheKey(userID, skey)
	// TODO get from cache
	loginSessionInfo, err := u.GetValidLoginSessionBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err.Mark()
	}
	if loginSessionInfo == nil {
		return nil, nil
	}

	accountInfo := &entity.AccountCacheInfo{}
	marshalErr := json.Unmarshal([]byte(loginSessionInfo.UserInfo), accountInfo)
	if marshalErr != nil {
		return nil, bmserror.NewError(constant.ErrInternalServer, marshalErr.Error())
	}
	return accountInfo, nil
}

// 通过session_id查询未过期的登录session
func (u *UserManager) GetValidLoginSessionBySessionID(ctx context.Context, sessionID string) (*entity.LoginSession, *bmserror.BMSError) {
	loginSessionList := make([]*entity.LoginSession, 0)
	err := u.ds.GetDataSource(ctx, nil).Table(entity.LoginSessionTabTableName).Where("session_id = ?", sessionID).
		Where("expire_time >= ?", timeutil.GetCurrentUnix()).Find(&loginSessionList).GetError()
	if err != nil {
		return nil, err.Mark()
	}
	if len(loginSessionList) == 0 {
		return nil, nil
	}
	return loginSessionList[0], nil
}

func (u *UserManager) CreateUser(ctx context.Context, user *entity.UserTab) *bmserror.BMSError {
	if err := u.ds.GetDataSource(ctx, nil).Table(entity.UserTabName).Create(user).GetError(); err != nil {
		return err.Mark()
	}
	return nil
}

type UserSearchParam struct {
	SearchKey string `json:"search_key"`
	PageIn    *paginator.PageIn
}

func (u *UserManager) SearchUserMng(ctx context.Context, params *UserSearchParam) ([]*entity.UserTab, int64, *bmserror.BMSError) {
	var userList []*entity.UserTab
	db := u.ds.GetDataSource(ctx, nil).Table(entity.UserTabName)
	if params.SearchKey != "" {
		db = db.Where("name LIKE ? OR email LIKE ? OR phone LIKE ?", "%"+params.SearchKey+"%", "%"+params.SearchKey+"%", "%"+params.SearchKey+"%")
	}
	total, err := paginator.Paginator(db, params.PageIn, &userList)
	if err != nil {
		return nil, 0, err.Mark()
	}
	return userList, total, nil
}

func (u *UserManager) GetUserByCode(ctx context.Context, code string) (*entity.UserTab, *bmserror.BMSError) {
	var user entity.UserTab
	db := u.ds.GetDataSource(ctx, nil).Table(entity.UserTabName).Where("email = ? OR phone = ?", code, code).First(&user)
	if db.RecordNotFound() {
		return nil, nil
	}
	if err := db.GetError(); err != nil {
		return nil, err.Mark()
	}
	return &user, nil
}

func (u *UserManager) GetUserByEmail(ctx context.Context, email string) (*entity.UserTab, *bmserror.BMSError) {
	var user entity.UserTab
	db := u.ds.GetDataSource(ctx, nil).Table(entity.UserTabName).Where("email = ?", email).First(&user)
	if db.RecordNotFound() {
		return nil, nil
	}
	if err := db.GetError(); err != nil {
		return nil, err.Mark()
	}
	return &user, nil
}

func HashPassword(password, salt string) (string, *bmserror.BMSError) {
	hashStr := password + salt
	hash, err := util.GetMD5Encode(hashStr)
	if err != nil {
		return "", err.Mark()
	}
	return hash, nil
}

func (u *UserManager) CreateLoginSession(ctx context.Context, session *entity.LoginSession) *bmserror.BMSError {
	if err := u.ds.GetDataSource(ctx, nil).Table(entity.LoginSessionTabTableName).Create(session).GetError(); err != nil {
		return err.Mark()
	}
	return nil
}
