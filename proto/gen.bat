@echo on
protoc --go_out=./ --vkit_out=./ --vkit_opt=--handlePath=../handler --swagger_out=./ --validate_out="lang=go:./"  .\sso\sso.proto
