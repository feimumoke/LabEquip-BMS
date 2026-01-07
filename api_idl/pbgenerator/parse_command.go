package pbgenerator

import (
	"errors"
	"flag"
	"fmt"
	"path"
)

type CommandParam struct {
	Dir   string
	force bool
}

var commandHelp = flag.Bool("h", false, "使用帮助")
var commandDir = flag.String("d", "", "指定某个目录, 必填")
var commandForce = flag.Bool("f", false, "强制重新生成校验文件：必须结合 -d 使用，避免整个项目被覆盖")

func GetCommandParams() (*CommandParam, error) {

	flag.Parse()

	if *commandHelp {
		flag.Usage()
		return nil, nil
	}

	dir := *commandDir
	force := *commandForce

	if len(dir) == 0 {
		return nil, errors.New("-d 必填")
	}

	if !force && len(dir) == 0 { // -f 要配合 -d 使用
		return nil, errors.New("-f 需配合 -d 使用")
	}
	if dir == "./" || dir == "." {
		return nil, fmt.Errorf("命令行参数不合法：请指定项目相对路径，请勿指定为项目根路径： %s", dir)
	}
	if len(dir) == 0 {
		dir = "./"
	}
	if path.IsAbs(dir) {
		return nil, fmt.Errorf("命令行参数不合法：请输入相对路径(请勿填写绝对路径)： %s", dir)
	}
	_, exist := FileInfo(dir)
	if !exist {
		return nil, fmt.Errorf("命令行参数不合法：目录不存在： %s", dir)
	}

	return &CommandParam{force: force, Dir: dir}, nil

}
