package gen_proto

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/visonlv/protoc-gen-vison/logger"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func GenerateServer(plugin *protogen.Plugin, serverName string, handlePath string) error {
	modUrl := ReadModUrl(handlePath + "/..")
	filePath := handlePath
	CreateDir(filePath)
	configFile, err := os.OpenFile(filePath+"/handler.go", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		logger.Infof("read file fail:%s", err)
		return err
	}
	WriteLine(configFile, versionName)
	WriteLine(configFile, "package handler")

	var importBuf bytes.Buffer
	var serverListBuf bytes.Buffer
	var urlsBuf bytes.Buffer
	var serverBuf bytes.Buffer
	for _, file := range plugin.Files {
		if len(file.Services) == 0 {
			continue
		}

		fileName := file.Desc.Path()
		nameArr := strings.Split(fileName, "/")
		if len(nameArr) <= 0 {
			continue
		}
		fileName = nameArr[len(nameArr)-1]
		fileName = strings.ReplaceAll(fileName, ".proto", "")
		// source/flow_log.proto
		filePath := fmt.Sprintf("%s/%s", handlePath, fileName)

		//server.go
		importBuf.Write([]byte(fmt.Sprintf("%s_handler \"%s/handler/%s\"\n\t", fileName, modUrl, fileName)))

		newFile := false
		if !CheckFileIsExist(filePath) {
			os.MkdirAll(filePath, os.ModePerm)
			newFile = true
		}

		var fileWriter *os.File
		var err error
		if newFile {
			fileWriter, err = os.OpenFile(fmt.Sprintf("%s/%s.go", filePath, fileName), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		} else {
			fileWriter, err = os.OpenFile(fmt.Sprintf("%s/%s.go", filePath, fileName), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		}
		if err != nil {
			logger.Infof("open file failed, err:%s", err)
			return nil
		}
		defer fileWriter.Close()

		if newFile {
			WriteLine(fileWriter, versionName)
			WriteLine(fileWriter, "package ", fileName)

			str, _ := os.Getwd()
			logger.Infof("%s", str)
			importPath := fmt.Sprintf("%s/proto/%s", ReadModUrl(handlePath+"/.."), serverName)
			index := strings.Index(str, "/proto/")
			if index > 0 {
				logger.Infof("%s", str[:index])
				logger.Infof("%s", str[index:])
				importPath = fmt.Sprintf("%s%s/%s", ReadModUrl(str[:index]), str[index:], serverName)
			}

			WriteLine(fileWriter, fmt.Sprintf(`
import (
	context "context"

	pb "%s"
)`, importPath))
		}

		serviceExistMap := make(map[string]bool)
		input := bufio.NewScanner(fileWriter)
		for input.Scan() {
			lineText := input.Text()
			if strings.HasPrefix(lineText, "func ") {
				lineText = lineText[strings.Index(lineText, "*")+1:]
				lineText = lineText[:strings.Index(lineText, "(")]
				lineText = strings.ReplaceAll(lineText, " ", "")
				serviceExistMap[lineText] = true
			} else if strings.HasPrefix(lineText, "type ") {
				lineText = strings.ReplaceAll(lineText, "struct {", "")
				lineText = strings.ReplaceAll(lineText, "type", "")
				lineText = strings.ReplaceAll(lineText, " ", "")
				serviceExistMap[lineText] = true
			}
		}

		for _, service := range file.Services {
			if ok := serviceExistMap[service.GoName]; !ok {
				WriteLine(fileWriter, fmt.Sprintf(`
type %s struct {
}`, service.GoName))
			}

			serverListBuf.Write([]byte(fmt.Sprintf("list = append(list, &%s_handler.%s{})\n\t", fileName, service.GoName)))

			sd := &serviceDesc{
				ServiceType: service.GoName,
				ServiceName: string(service.Desc.FullName()),
				Metadata:    file.Desc.Path(),
			}
			for _, method := range service.Methods {
				rule, ok := proto.GetExtension(method.Desc.Options(), annotations.E_Http).(*annotations.HttpRule)
				var methodDesc *methodDesc
				if rule != nil && ok {
					methodDesc = buildHTTPRule(serverName, method, rule)
				} else {
					path := fmt.Sprintf("/%s/%s", service.Desc.FullName(), method.Desc.Name())
					methodDesc = buildMethodDesc(serverName, method, "POST", path)
				}
				sd.Methods = append(sd.Methods, methodDesc)

				methodKey := fmt.Sprintf("%s)%s", service.GoName, methodDesc.Name)
				serviceMethodKey := service.GoName + "." + methodDesc.Name
				urlsBuf.Write([]byte(fmt.Sprintf("methodDescMap[\"%s\"] = []string{\"%s\", \"%t\", \"%t\"}\n\t", serviceMethodKey, methodDesc.Path, method.Desc.IsStreamingClient(), method.Desc.IsStreamingServer())))

				if method.Desc.IsStreamingClient() && method.Desc.IsStreamingServer() {
					serverBuf.Write([]byte(ReplaceList(allStreamServer, "${serviceName}", service.GoName, "${methodName}", method.GoName, "${req}", methodDesc.Request, "${resp}", methodDesc.Reply, "${methodPath}", methodDesc.Path)))
				} else if method.Desc.IsStreamingClient() {
					serverBuf.Write([]byte(ReplaceList(clientStreamServer, "${serviceName}", service.GoName, "${methodName}", method.GoName, "${req}", methodDesc.Request, "${resp}", methodDesc.Reply, "${methodPath}", methodDesc.Path)))
				} else if method.Desc.IsStreamingServer() {
					serverBuf.Write([]byte(ReplaceList(serverStreamServer, "${serviceName}", service.GoName, "${methodName}", method.GoName, "${req}", methodDesc.Request, "${resp}", methodDesc.Reply, "${methodPath}", methodDesc.Path)))
				}
				if ok := serviceExistMap[methodKey]; !ok {
					if method.Desc.IsStreamingClient() && method.Desc.IsStreamingServer() {
						WriteLine(fileWriter, ReplaceList(allStreamServerFunc, "${serviceName}", service.GoName, "${methodName}", method.GoName, "${req}", methodDesc.Request, "${resp}", methodDesc.Reply, "${methodPath}", methodDesc.Path))
					} else if method.Desc.IsStreamingClient() {
						WriteLine(fileWriter, ReplaceList(clientStreamServerFunc, "${serviceName}", service.GoName, "${methodName}", method.GoName, "${req}", methodDesc.Request, "${resp}", methodDesc.Reply, "${methodPath}", methodDesc.Path))
					} else if method.Desc.IsStreamingServer() {
						WriteLine(fileWriter, ReplaceList(serverStreamServerFunc, "${serviceName}", service.GoName, "${methodName}", method.GoName, "${req}", methodDesc.Request, "${resp}", methodDesc.Reply, "${methodPath}", methodDesc.Path))
					} else {
						WriteLine(fileWriter, ReplaceList(nolmalServer, "${serviceName}", service.GoName, "${methodName}", method.GoName, "${req}", methodDesc.Request, "${resp}", methodDesc.Reply, "${methodPath}", methodDesc.Path))
					}
				}
			}
		}
	}

	WriteLine(configFile, fmt.Sprintf(`
import (
	%s
)`, importBuf.String()))

	WriteLine(configFile, fmt.Sprintf(`
func GetLsit() []interface{} {
	list := make([]interface{}, 0)
	%s
	return list
}`, serverListBuf.String()))

	WriteLine(configFile, fmt.Sprintf(`
func GetUrlMap() map[string][]string {
	methodDescMap := make(map[string][]string)
	%s
	return methodDescMap
}`, urlsBuf.String()))

	logger.Infof("len %d", serverBuf.Len())
	if serverBuf.Len() > 0 {
		serverFullPath := serverName + "/server.go"
		serverWriter, err := os.OpenFile(serverFullPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			logger.Infof("open file:%s failed, err:%s", serverFullPath, err)
			return nil
		}
		defer serverWriter.Close()

		WriteLine(serverWriter, versionName)
		WriteLine(serverWriter, "package ", serverName)
		WriteLine(serverWriter, `
import (
	"google.golang.org/grpc"
)
		`)

		WriteLine(serverWriter, serverBuf.String())
	}

	return nil
}
