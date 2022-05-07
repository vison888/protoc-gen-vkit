package gen_proto

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/visonlv/protoc-gen-vison/logger"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
)

var methodSets = make(map[string]int)

func CreateDir(filePath string) error {
	if !CheckFileIsExist(filePath) {
		err := os.MkdirAll(filePath, os.ModePerm)
		return err
	}
	return nil
}

func CheckFileIsExist(filePath string) bool {
	var exist = true
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func WriteLine(writer io.Writer, v ...interface{}) {
	for _, x := range v {
		switch x := x.(type) {
		case string:
			writer.Write([]byte(x))
		}
	}
	writer.Write([]byte("\n"))
}

func LineExitIndex() int {
	return 0
}

func buildHTTPRule(serverName string, m *protogen.Method, rule *annotations.HttpRule) *methodDesc {
	var (
		path         string
		method       string
		body         string
		responseBody string
	)

	switch pattern := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		path = pattern.Get
		method = "GET"
	case *annotations.HttpRule_Put:
		path = pattern.Put
		method = "PUT"
	case *annotations.HttpRule_Post:
		path = pattern.Post
		method = "POST"
	case *annotations.HttpRule_Delete:
		path = pattern.Delete
		method = "DELETE"
	case *annotations.HttpRule_Patch:
		path = pattern.Patch
		method = "PATCH"
	case *annotations.HttpRule_Custom:
		path = pattern.Custom.Path
		method = pattern.Custom.Kind
	}
	body = rule.Body
	responseBody = rule.ResponseBody
	md := buildMethodDesc(serverName, m, method, path)
	if method == "GET" || method == "DELETE" {
		if body != "" {
			_, _ = fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: %s %s body should not be declared.\n", method, path)
		}
	} else {
		if body == "" {
			_, _ = fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: %s %s does not declare a body.\n", method, path)
		}
	}
	if body == "*" {
		md.HasBody = true
		md.Body = ""
	} else if body != "" {
		md.HasBody = true
		md.Body = "." + camelCaseVars(body)
	} else {
		md.HasBody = false
	}
	if responseBody == "*" {
		md.ResponseBody = ""
	} else if responseBody != "" {
		md.ResponseBody = "." + camelCaseVars(responseBody)
	}
	return md
}

func buildMethodDesc(serverName string, m *protogen.Method, method, path string) *methodDesc {
	defer func() { methodSets[m.GoName]++ }()
	return &methodDesc{
		Name:    m.GoName,
		Num:     methodSets[m.GoName],
		Request: m.Input.GoIdent.GoName,
		Reply:   m.Output.GoIdent.GoName,
		Path:    path,
		Method:  method,
	}
}

func camelCaseVars(s string) string {
	subs := strings.Split(s, ".")
	vars := make([]string, 0, len(subs))
	for _, sub := range subs {
		vars = append(vars, camelCase(sub))
	}
	return strings.Join(vars, ".")
}

func camelCase(s string) string {
	if s == "" {
		return ""
	}
	t := make([]byte, 0, 32)
	i := 0
	if s[0] == '_' {
		t = append(t, 'X')
		i++
	}
	for ; i < len(s); i++ {
		c := s[i]
		if c == '_' && i+1 < len(s) && isASCIILower(s[i+1]) {
			continue // Skip the underscore in s.
		}
		if isASCIIDigit(c) {
			t = append(t, c)
			continue
		}
		if isASCIILower(c) {
			c ^= ' ' // Make it a capital letter.
		}
		t = append(t, c) // Guaranteed not lower case.
		for i+1 < len(s) && isASCIILower(s[i+1]) {
			i++
			t = append(t, s[i])
		}
	}
	return string(t)
}

func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func ReadModUrl() string {
	fileReader, err := os.OpenFile("../go.mod", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		logger.Infof("open file failed, err:%s", err)
		return ""
	}
	defer fileReader.Close()
	input := bufio.NewScanner(fileReader)
	for input.Scan() {
		lineText := input.Text()
		if strings.HasPrefix(lineText, "module") {
			lineText = strings.ReplaceAll(lineText, "module", "")
			lineText = strings.ReplaceAll(lineText, " ", "")
			return lineText
		}
	}
	return ""
}
