package view

import (
	"context"
	"encoding/json"

	pbbasic "github.com/feimumoke/labequipbms/api_idl/apps/basic"
	"github.com/feimumoke/labequipbms/apps/basic/manager"
	"github.com/feimumoke/labequipbms/apps/common/idutil"
	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/defines/entity"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/support/convert"
	"github.com/feimumoke/labequipbms/framework/support/paginator"
	"github.com/feimumoke/labequipbms/framework/support/timeutil"
	"github.com/feimumoke/labequipbms/framework/support/util"
	"github.com/feimumoke/labequipbms/framework/transaction"
	"github.com/feimumoke/labequipbms/framework/web"
	"github.com/google/uuid"
)

type UserHandler struct {
	userMng *manager.UserManager
}

func NewUserHandler() *UserHandler {
	return &UserHandler{userMng: manager.NewUserManager()}
}

func (h UserHandler) CreateUserHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	req := i.(*pbbasic.CreateUserRequest)
	if req.GetName() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "name is empty")
	}
	if req.GetEmail() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "email is empty")
	}
	if req.GetPassword() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "password is empty")
	}
	if req.GetPhone() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "phone is empty")
	}

	now := timeutil.GetCurrentUnix()
	var userNo string
	transactionErr := transaction.PropagationRequired(ctx, func(ctx context.Context) *bmserror.BMSError {
		// 检查用户是否已存在
		existingUser, bmsError := h.userMng.GetUserByCode(ctx, req.GetEmail())
		if bmsError != nil {
			return bmsError.Mark()
		}
		if existingUser != nil {
			return bmserror.NewError(constant.ErrParam, "user with email %v already exists", req.GetEmail())
		}

		// 检查手机号是否已存在
		existingUser, bmsError = h.userMng.GetUserByCode(ctx, req.GetPhone())
		if bmsError != nil {
			return bmsError.Mark()
		}
		if existingUser != nil {
			return bmserror.NewError(constant.ErrParam, "user with phone %v already exists", req.GetPhone())
		}

		// 生成用户ID
		var bmsError2 *bmserror.BMSError
		userNo, bmsError2 = idutil.GenUserNo(ctx)
		if bmsError2 != nil {
			return bmsError2.Mark()
		}

		// 生成salt
		salt, bmsError2 := util.RandStr(16)
		if bmsError2 != nil {
			return bmsError2.Mark()
		}

		// 加密密码
		hashedPassword, bmsError2 := manager.HashPassword(req.GetPassword(), salt)
		if bmsError2 != nil {
			return bmsError2.Mark()
		}

		user := &entity.UserTab{
			UserID:    userNo,
			Email:     req.GetEmail(),
			Name:      req.GetName(),
			Phone:     req.GetPhone(),
			Passwd:    hashedPassword,
			Salt:      salt,
			Gender:    req.GetGender(),
			Role:      constant.UserRoleType(req.GetRole()),
			Introduce: req.GetIntroduce(),
			Status:    1,
			Ctime:     now,
			Mtime:     now,
		}

		cErr := h.userMng.CreateUser(ctx, user)
		if cErr != nil {
			return cErr.Mark()
		}
		return nil
	})
	if transactionErr != nil {
		return nil, transactionErr.Mark()
	}
	return &pbbasic.CreateUserResponse{UserId: &userNo}, nil
}

func (h UserHandler) SearchUserHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	req := i.(*pbbasic.SearchUserRequest)
	pageIn := &paginator.PageIn{
		Pageno:     req.GetPageno(),
		Count:      req.GetCount(),
		IsGetTotal: true,
	}
	userList, total, bmsError := h.userMng.SearchUserMng(ctx, &manager.UserSearchParam{
		SearchKey: req.GetSearchKey(),
		PageIn:    pageIn,
	})
	if bmsError != nil {
		return nil, bmsError.Mark()
	}
	var userInfoList []*pbbasic.UserInfo
	for _, user := range userList {
		userInfo := &pbbasic.UserInfo{
			UserId:   &user.UserID,
			UserName: &user.Name,
			Role:     convert.Int64(user.Role),
			Avatar:   &user.Avatar,
			Phone:    &user.Phone,
			Email:    &user.Email,
			Ctime:    convert.Int64(user.Ctime),
			Gender:   convert.Int64(user.Gender),
		}
		userInfoList = append(userInfoList, userInfo)
	}
	totalPtr := total
	return &pbbasic.SearchUserResponse{
		Total: &totalPtr,
		List:  userInfoList,
	}, nil
}

func (h UserHandler) UserLoginHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	req := i.(*pbbasic.UserLoginRequest)
	if req.GetCode() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "code is empty")
	}
	if req.GetPasswd() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "passwd is empty")
	}

	// 根据 code (email 或 phone) 查找用户
	user, bmsError := h.userMng.GetUserByCode(ctx, req.GetCode())
	if bmsError != nil {
		return nil, bmsError.Mark()
	}
	if user == nil {
		loginStatus := int64(0)
		return &pbbasic.UserLoginResponse{
			LoginStatus: &loginStatus,
			UserName:    nil,
		}, nil
	}

	// 验证密码
	hashedPassword, bmsError := manager.HashPassword(req.GetPasswd(), user.Salt)
	if bmsError != nil {
		return nil, bmsError.Mark()
	}
	if hashedPassword != user.Passwd {
		loginStatus := int64(0)
		return &pbbasic.UserLoginResponse{
			LoginStatus: &loginStatus,
			UserName:    nil,
		}, nil
	}

	// 创建登录session
	now := timeutil.GetCurrentUnix()
	sessionID := uuid.New().String()
	sessionID = sessionID[:32]    // 使用32位session ID
	expireTime := now + 7*24*3600 // 7天过期

	accountInfo := &entity.AccountCacheInfo{
		UserID:    user.UserID,
		UserName:  user.Name,
		Email:     user.Email,
		LoginTime: now,
		LoginIp:   header.ClientIP,
		Skey:      sessionID,
	}
	userInfoJson, marshalErr := json.Marshal(accountInfo)
	if marshalErr != nil {
		return nil, bmserror.NewError(constant.ErrInternalServer, marshalErr.Error())
	}

	loginSession := &entity.LoginSession{
		UserID:     user.UserID,
		SessionID:  manager.GenerateUserSessionCacheKey(user.UserID, sessionID),
		UserInfo:   string(userInfoJson),
		ExpireTime: expireTime,
		Ctime:      now,
		Mtime:      now,
	}

	transactionErr := transaction.PropagationRequired(ctx, func(ctx context.Context) *bmserror.BMSError {
		cErr := h.userMng.CreateLoginSession(ctx, loginSession)
		if cErr != nil {
			return cErr.Mark()
		}
		return nil
	})
	if transactionErr != nil {
		return nil, transactionErr.Mark()
	}

	loginStatus := int64(1)
	userName := user.Name
	return &pbbasic.UserLoginResponse{
		LoginStatus: &loginStatus,
		UserName:    &userName,
	}, nil
}
