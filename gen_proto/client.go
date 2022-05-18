package gen_proto

import (
	"bytes"
	"fmt"
	"os"

	"github.com/visonlv/protoc-gen-vison/logger"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func GenerateClient(plugin *protogen.Plugin, serverName string) error {
	fileFullPath := serverName + "/client.go"
	fileWriter, err := os.OpenFile(fileFullPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		logger.Infof("open file:%s failed, err:%s", fileFullPath, err)
		return nil
	}
	defer fileWriter.Close()

	WriteLine(fileWriter, versionName)
	WriteLine(fileWriter, "package ", serverName)
	WriteLine(fileWriter, `
import (
	context "context"

	"google.golang.org/grpc"
)`)

	var clientBuf bytes.Buffer
	var newClientBuf bytes.Buffer
	for _, file := range plugin.Files {
		if len(file.Services) == 0 {
			continue
		}

		for _, service := range file.Services {
			clientBuf.Write([]byte(fmt.Sprintf("%s   *%sClient\n\t", service.GoName, service.GoName)))
			newClientBuf.Write([]byte(fmt.Sprintf("client.%s = &%sClient{cc}\n\t", service.GoName, service.GoName)))

			WriteLine(fileWriter, fmt.Sprintf(`
type %sClient struct {
	cc grpc.ClientConnInterface
}`, service.GoName))
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

				if method.Desc.IsStreamingClient() && method.Desc.IsStreamingServer() {
					WriteLine(fileWriter, ReplaceList(allStreamClient, "${serviceName}", service.GoName, "${methodName}", method.GoName, "${req}", methodDesc.Request, "${resp}", methodDesc.Reply))
					WriteLine(fileWriter, ReplaceList(allStreamClientFunc, "${serviceName}", service.GoName, "${methodName}", method.GoName, "${methodPath}", methodDesc.Path))
				} else if method.Desc.IsStreamingClient() {
					WriteLine(fileWriter, ReplaceList(clientStreamClient, "${serviceName}", service.GoName, "${methodName}", method.GoName, "${req}", methodDesc.Request, "${resp}", methodDesc.Reply))
					WriteLine(fileWriter, ReplaceList(clientStreamClientFunc, "${serviceName}", service.GoName, "${methodName}", method.GoName, "${methodPath}", methodDesc.Path, "${req}", methodDesc.Request))
				} else if method.Desc.IsStreamingServer() {
					WriteLine(fileWriter, ReplaceList(serverStreamClient, "${serviceName}", service.GoName, "${methodName}", method.GoName, "${req}", methodDesc.Request, "${resp}", methodDesc.Reply))
					WriteLine(fileWriter, ReplaceList(serverStreamClientFunc, "${serviceName}", service.GoName, "${methodName}", method.GoName, "${methodPath}", methodDesc.Path))
				} else {
					WriteLine(fileWriter, ReplaceList(normalClientFunc, "${serviceName}", service.GoName, "${methodName}", method.GoName, "${req}", methodDesc.Request, "${resp}", methodDesc.Reply, "${methodPath}", methodDesc.Path))
				}
			}

			WriteLine(fileWriter, fmt.Sprintf(`
func New%sClient(cc grpc.ClientConnInterface) *%sClient {
	return &%sClient{cc}
}`, service.GoName, service.GoName, service.GoName))
		}
	}

	camelName := camelCase(serverName)
	WriteLine(fileWriter, fmt.Sprintf(`
type %sClient struct {
	%s
}`, camelName, clientBuf.String()))

	WriteLine(fileWriter, fmt.Sprintf(`
func New%sClient(cc grpc.ClientConnInterface) *%sClient {
	client := &%sClient{}
	%s
	return client
}`, camelName, camelName, camelName, newClientBuf.String()))

	fileWriter.Close()
	return nil
}
