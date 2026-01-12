package view

import (
	pbbasic "github.com/feimumoke/labequipbms/api_idl/apps/basic"
	"github.com/feimumoke/labequipbms/framework/asynctask"
	"github.com/feimumoke/labequipbms/framework/web"
)

func InitBasicView(s *web.BasicServer, r *asynctask.AsyncRunner) {
	initEquip(s, r)
	initLab(s, r)
	initUser(s, r)
}

func initEquip(s *web.BasicServer, r *asynctask.AsyncRunner) {
	h := NewEquipHandler()
	s.RegisterPOST("/apps/basic/equip/create_equip", h.CreateEquipHandler, &pbbasic.CreateEquipRequest{})
	s.RegisterPOST("/apps/basic/equip/search_equip", h.SearchEquipHandler, &pbbasic.SearchEquipRequest{})
}

func initLab(s *web.BasicServer, r *asynctask.AsyncRunner) {
	h := NewLabHandler()
	s.RegisterPOST("/apps/basic/lab/search_lab", h.SearchLabHandler, &pbbasic.SearchLabRequest{})
}

func initUser(s *web.BasicServer, r *asynctask.AsyncRunner) {
	h := NewUserHandler()
	s.RegisterPOST("/apps/common/enums", h.GetEnumsView, &struct{}{})
	s.RegisterPOST("/apps/basic/user/create_user", h.CreateUserHandler, &pbbasic.CreateUserRequest{})
	s.RegisterPOST("/apps/basic/user/search_user", h.SearchUserHandler, &pbbasic.SearchUserRequest{})
	s.RegisterPOST("/apps/basic/user/user_login", h.UserLoginHandler, &pbbasic.UserLoginRequest{})

}
