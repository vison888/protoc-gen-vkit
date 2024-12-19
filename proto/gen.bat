@echo on
protoc --go_out=./ --infore_out=./ --infore_opt=--handlePath=../handler --swagger_out=./ --validate_out="lang=go:./"  .\sso\sso.proto
