package pbgenerator

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Rule struct {
	paramType         ParamType
	paramDeclar       string
	paramDeclarPrefix string
	keyValMap         map[string]string
	defaultVal        *string
	isRepeated        bool
	isOk              bool
}

type ParamType = string

const (
	UnKnown  ParamType = ""
	String   ParamType = "string"
	Int64    ParamType = "int64"
	Repeated ParamType = "repeated"
)

// 占位符参数
var placeholderMap = make(map[string]string)

func init() {

	var err error
	placeholderMap, err = readConfig()
	if err != nil {
		panic("error: " + err.Error())
	}

}

func Do() error {

	commands, err := GetCommandParams()
	if err != nil {
		return err
	}
	if commands == nil { // 查看帮助
		return nil
	}

	err = WalkDir(commands.Dir)

	if err != nil {
		return err
	}

	cmd := exec.Command("gofmt", "-l", "-w", commands.Dir)
	err = cmd.Run()
	if err != nil {
		return errors.New("gofmt error")
	}
	return nil
}

func WalkDir(dir string) error {

	fmt.Println("Working dir: ", dir)
	suffix := ".tpl.proto"
	var protoFilePaths []map[string]string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return fmt.Errorf("error: %s", err.Error())
		}

		if strings.HasSuffix(path, suffix) {

			// 遍历每一个 .tpl.proto 文件，替换其中的校验规则
			// 替换文件操作
			protoTplMTime, _ := FileInfo(path)

			suffix := strings.TrimSuffix(path, suffix) // hellovalidator.tpl.proto -> hellovalidator
			pbGoFilePath := suffix + ".pb.go"
			pbGoMTime, fileExist := FileInfo(pbGoFilePath)

			if fileExist && protoTplMTime < pbGoMTime && !*commandForce {
				// 如果 .tpl.proto 的 修改时间小于 .pb.go 的修改时间，说明没有修改过文件，跳过
				// 不用 .proto 的修改时间判断的原因时：可能生成了 proto 文件，但是生成pb文件报错，此时应该支持重复执行
				// 如果没有修改过，但是填了强制执行的参数，也要执行
				return nil
			}

			fmt.Println(path)
			fileByte, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("读取.tpl.proto文件失败: %s", path)
			}
			fileContent := string(fileByte)
			fileContent = strings.ReplaceAll(fileContent, ".tpl.proto", ".proto")
			fileContent, err = doReplaceFileContent(fileContent)
			if err != nil {
				return fmt.Errorf("文件校验替换失败: %s", err.Error())
			}

			protoFilePath := suffix + ".proto" // hellovalidator.tpl.proto -> hellovalidator.proto
			err = ioutil.WriteFile(protoFilePath, []byte(fileContent), 0777)
			if err != nil {
				return fmt.Errorf("write file error: %s", protoFilePath)
			}

			m := map[string]string{
				"pbGoFilePath":  pbGoFilePath,
				"protoFilePath": protoFilePath,
			}
			protoFilePaths = append(protoFilePaths, m)

		}

		return nil

	})

	for _, protoFilePath := range protoFilePaths {

		pbGoFilePath := protoFilePath["pbGoFilePath"]
		// 执行protoc命令
		err = execCommand(protoFilePath["protoFilePath"])
		if err != nil {
			return err
		}

		// 替换pb文件的omitempty
		fileByte, err := ioutil.ReadFile(pbGoFilePath)
		if err != nil {
			return fmt.Errorf("读取.pb.go文件失败: %s", pbGoFilePath)
		}
		fileContent := string(fileByte)
		fileContent = strings.ReplaceAll(fileContent, ",omitempty", "")
		err = ioutil.WriteFile(pbGoFilePath, []byte(fileContent), 0777)
		if err != nil {
			return fmt.Errorf("write file error: %s", pbGoFilePath)
		}

	}

	return err
}

// doReplaceFileContent 做文件替换
// fileContent 为原文件内容
// protoPlaceholderMap 为配置文件变量到替换值的映射
func doReplaceFileContent(fileContent string) (string, error) {

	regex := regexp.MustCompile(`(.*;).*// ?conf: ?(.*)`)
	content := regex.FindAllStringSubmatch(fileContent, -1)

	replaceMap := make(map[string]string)
	// 规范：
	// required string username = 1; //valid: (max_len=10, pattern="STRING_NOT_EMPTY")
	for _, oneParamLine := range content { // 获取多行
		rule := &Rule{paramType: UnKnown, isRepeated: false}

		for i, one := range oneParamLine { // 某行

			// 第一个参数为完整的声明语句
			if i == 0 {
				rule.paramDeclar = one
			}
			// 第二个参数为正则表达式的最左分组，如 paramDeclarPrefix 为 required string username = 1，用来做最后拼接字符串
			// 通过判断是否包含字符串判断类型，与 valid 绑定
			if i == 1 {

				rule.paramDeclarPrefix = one

				isString := strings.Contains(rule.paramDeclarPrefix, " string ")
				isInt64 := strings.Contains(rule.paramDeclarPrefix, " int64 ")
				if isInt64 {
					rule.paramType = Int64
				}
				if isString {
					rule.paramType = String
				}
				rule.isRepeated = strings.Contains(rule.paramDeclarPrefix, " repeated ")
				if rule.paramType == UnKnown {
					break
				}

			}
			// 第三个参数则为 校验规则
			if i == 2 {
				if rule.paramType == UnKnown { // 如果第一个参数没有获取到类型，报错
					return "", fmt.Errorf("参数类型不在(int64/string)类型内，暂不支持参数校验 : (%s)", rule.paramDeclar)
				}
				paramsStr := one[1 : len(one)-1]         // 括号去掉，(max_len=10, pattern="STRING_NOT_EMPTY") => max_len=10, pattern="STRING_NOT_EMPTY"
				params := strings.Split(paramsStr, ", ") // 按照 , 截取，得到 max_len=10 和 pattern="STRING_NOT_EMPTY"
				keyValMap := make(map[string]string)

				for _, oneParam := range params {

					keyVal := strings.Split(oneParam, "=") // 按照 = 截取出key val，也就是 max_len: 10, pattern: "STRING_NOT_EMPTY"，保存到 keyValMap 中
					if len(keyVal) < 2 {                   // 如果只有一个值，说明没有 = 号
						return "", fmt.Errorf("校验参数不合法 (%s)，请核对是否包含完整键值对", rule.paramDeclar)
					}
					if keyVal[0] == Default {
						rule.defaultVal = &keyVal[1]
						continue
					}
					keyValMap[keyVal[0]] = keyVal[1]

				}
				rule.keyValMap = keyValMap

			}
			rule.isOk = true
		}
		if rule.isOk {
			validateStr, err := appendValidateStr(rule)
			if err != nil {
				return "", err
			}
			defaultStr, err := appendDefaultStr(rule)
			if err != nil {
				return "", err
			}
			resultStr := ""
			if len(defaultStr) != 0 && len(validateStr) != 0 {
				resultStr = fmt.Sprintf(" [(validator.field) = {%s}, default = %s] ;", validateStr, defaultStr)
			} else if len(validateStr) != 0 {
				resultStr = fmt.Sprintf(" [(validator.field) = {%s}] ;", validateStr)
			} else {
				resultStr = fmt.Sprintf(" [default = %s]; ", defaultStr)
			}
			// paramDeclarPrefix 多了个分号这里去掉
			replaceMap[rule.paramDeclarPrefix] = strings.ReplaceAll(rule.paramDeclarPrefix, ";", "") + resultStr
		}
	}

	for key, val := range replaceMap {
		fileContent = strings.ReplaceAll(fileContent, key, val)
	}
	return fileContent, nil
}

func appendDefaultStr(rule *Rule) (string, error) {

	if rule.defaultVal != nil {
		if rule.isRepeated {
			return "", errors.New("repeated 类型不支持default")
		}
		return *rule.defaultVal, nil
	}
	return "", nil

}

// appendValidateStr 用来做 校验行的拼接
func appendValidateStr(rule *Rule) (string, error) {

	legalKeys, _ := legalParamsMap[rule.paramType]
	if rule.isRepeated { // 如果是repeated类型的，把repeated支持的校验参数也加进去
		legalKeys = append(legalKeys, legalParamsMap[Repeated]...)
	}
	// 为了让 key val 在拼接 validate 的时候不会因为 遍历 rule.keyValMap 的时候乱序，所以这里先放到 allKeys 的然后遍历以保证顺序性，以避免不同人跑同一个命令行导致校验参数顺序不一样
	allKeys := []string{}
	keyValMap := make(map[string]string)
	for key, val := range rule.keyValMap {
		allKeys = append(allKeys, key)
		keyValMap[key] = val
	}
	// 按照 key 的大小排序
	sort.Slice(allKeys, func(i, j int) bool {
		return allKeys[i] < allKeys[j]
	})

	validateStr := ""
	for _, key := range allKeys {
		val := keyValMap[key]
		// 判断 key 在对应类型中合法
		if !isInStringArray(key, legalKeys) {
			return "", fmt.Errorf("(%s) 校验参数(%s)不在合法参数内, 变量类型 (%s), 应为 (%v)", rule.paramDeclar, key, rule.paramType, legalKeys)
		}

		switch rule.paramType {
		case String:
			validateAppend, err := dealWithString(key, val, rule.paramDeclar)
			if err != nil {
				return "", err
			}
			validateStr += validateAppend
		case Int64:
			if rule.paramType == Int64 {
				validateAppend, err := dealWithInt64(key, val, rule.paramDeclar)
				if err != nil {
					return "", err
				}
				validateStr += validateAppend
				continue
			}
		}
	}

	return validateStr, nil
}

// String 类型处理
func dealWithString(key, val, paramDeclar string) (string, error) {

	internalKey, ok := legalParamsToInternalMap[key]
	if !ok {
		internalKey = key
	}

	validateStr := ""
	if isInStringArray(key, []string{StringMinLen, StringMaxLen}) { // 部分为int类型转为int
		intVal, err := strconv.Atoi(val)
		if err != nil {
			return "", fmt.Errorf("(%s)校验参数转为整型错误，校验值为(%s)", paramDeclar, val)
		}
		if key == StringMinLen {
			intVal = intVal - 1
		}
		if key == StringMaxLen {
			intVal = intVal + 1
		}
		validateStr = fmt.Sprintf("%s: %d, ", internalKey, intVal)
		return validateStr, nil
	}
	if key == StringRegex { //  遇到 regex 的，说明是正则表达式，此时判断是否在参数替换map中，如果是的话则替换
		if valReplace, ok := placeholderMap[strings.Trim(val, "\"")]; ok {
			validateStr = fmt.Sprintf("%s: %s, ", internalKey, "\""+valReplace+"\"")
			return validateStr, nil
		}
	}
	validateStr = fmt.Sprintf("%s: %s, ", legalParamsToInternalMap[key], val)
	return validateStr, nil
}

// Int64类型处理
func dealWithInt64(key, val, paramDeclar string) (string, error) {

	internalKey, ok := legalParamsToInternalMap[key]
	if !ok {
		internalKey = key
	}

	validateStr := ""
	if isInStringArray(key, []string{IntGTE, IntLTE, IntLT, IntGT}) { // 部分为int类型的转为int

		intVal, err := strconv.Atoi(val)
		if err != nil {
			return "", fmt.Errorf("(%s)校验参数转为整型错误，校验值为(%s)", paramDeclar, val)
		}

		if key == IntGTE {
			intVal -= 1
			internalKey = IntGTInternal
		}
		if key == IntLTE {
			intVal += 1
			internalKey = IntLTInternal
		}

		validateStr = fmt.Sprintf("%s: %d, ", internalKey, intVal)
		return validateStr, nil
	} else {
		validateStr = fmt.Sprintf("%s: %s, ", internalKey, val)
		return validateStr, nil
	}
}

// isInStringArray 判断 str 是否在 array中
func isInStringArray(str string, array []string) bool {
	for _, one := range array {
		if one == str {
			return true
		}
	}
	return false
}

// readConfig 读取 variables.config 配置文件
func readConfig() (map[string]string, error) {

	fileBytes, err := ioutil.ReadFile("variables.config")
	if err != nil {
		return nil, err
	}
	fileContent := string(fileBytes)
	variables := strings.Split(fileContent, "\n")

	protoPlaceholderMap := make(map[string]string)
	for _, variable := range variables {
		variableMap := strings.Split(variable, "=")
		varKey := variableMap[0]
		if len(variableMap) == 1 || len(varKey) == 0 { // 过滤空行或没有=的
			continue
		}
		varVal := variableMap[1]
		protoPlaceholderMap[varKey] = varVal
		//fmt.Println(varKey + " => " + varVal)
	}
	return protoPlaceholderMap, nil
}

// execCommand 用来执行 protoc 命令
func execCommand(newFilePath string) error {

	cmd := exec.Command("protoc", "-I", ".", "--govalidators_out=paths=source_relative:./", "--go_out=plugins=grpc,paths=source_relative:./", newFilePath)

	//fmt.Println("Cmd", cmd.Args)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(stderr.String() + err.Error())
	}
	return nil
}

func FileInfo(_path string) (int64, bool) {

	fi, err := os.Stat(_path)
	if err != nil && os.IsNotExist(err) {
		return 0, false
	}
	modTime := fi.ModTime().Unix()
	return modTime, true
}
