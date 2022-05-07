package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/visonlv/protoc-gen-vison/gen_proto"
	"github.com/visonlv/protoc-gen-vison/logger"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	// protoc --vison_out=./speech ./source/*.proto
	//1.读取标准输入，接收proto 解析的文件内容，并解析成结构体
	input, _ := ioutil.ReadAll(os.Stdin)
	var req pluginpb.CodeGeneratorRequest
	proto.Unmarshal(input, &req)
	if req.Parameter == nil {
		logger.Info("req.Parameter empty")
		panic(errors.New("req.Parameter empty"))
	}
	paramStr := *req.Parameter
	paramMap := make(map[string]string)
	for _, param := range strings.Split(paramStr, "--") {
		if param != "" {
			kv := strings.Split(param, "=")
			paramMap[kv[0]] = kv[1]
		}
	}

	serverName, b := paramMap["serverName"]
	if !b {
		logger.Info("serverName not exit")
		panic(errors.New("serverName not exit"))
	}

	handlePath := ""
	h, b := paramMap["handlePath"]
	if b {
		handlePath = h
		gen_proto.CreateDir(handlePath)
	}
	gen_proto.CreateDir("./" + serverName)

	//2、合并proto
	err := gen_proto.MergeProto(&req, serverName)
	if err != nil {
		logger.Infof("MergeProto err:%s", err.Error())
		panic(err)
	}
	// 3.生成插件
	opts := protogen.Options{}
	plugin, err := opts.New(&req)
	if err != nil {
		panic(err)
	}

	gen_proto.GenerateClient(plugin, serverName)
	if handlePath != "" {
		gen_proto.GenerateServer(plugin, serverName, handlePath)
	}

	// 生成响应
	stdout := plugin.Response()
	out, err := proto.Marshal(stdout)
	if err != nil {
		panic(err)
	}

	// 将响应写回 标准输入, protoc会读取这个内容
	fmt.Fprintf(os.Stdout, string(out))
}
