package gen_proto

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/visonlv/protoc-gen-vison/logger"
	"google.golang.org/protobuf/types/pluginpb"
)

//合并proto到文件
func MergeProto(req *pluginpb.CodeGeneratorRequest, serverName string) error {
	filePaths := req.FileToGenerate
	if len(filePaths) <= 0 {
		return nil
	}
	proName := serverName
	fileWriter, err := os.OpenFile(proName+".proto", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		logger.Infof("open file failed, err:%s", err)
		return nil
	}
	defer fileWriter.Close()

	isFirstProto := true
	for _, filePath := range filePaths {
		fileReader, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			logger.Infof("open file failed, err:%s", err)
			return nil
		}
		defer fileReader.Close()
		input := bufio.NewScanner(fileReader)
		startRead := false
		for input.Scan() {
			lineText := input.Text()
			if isFirstProto {
				if strings.HasPrefix(lineText, "package ") {
					lineText = fmt.Sprintf("package %s;", proName)
					WriteLine(fileWriter, lineText)
				} else if strings.HasPrefix(lineText, "option go_package ") {
					lineText = fmt.Sprintf("option go_package = \"./%s\";", proName)
					WriteLine(fileWriter, lineText)
				} else {
					WriteLine(fileWriter, lineText)
				}

			} else {
				if strings.HasPrefix(lineText, "service ") {
					startRead = true
				}
				if strings.HasPrefix(lineText, "message ") {
					startRead = true
				}
				if startRead {
					WriteLine(fileWriter, lineText)
				}
			}
		}
		isFirstProto = false
	}
	return nil
}
