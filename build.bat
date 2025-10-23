 set GOOS=linux
 set GOARCH=amd64
 set filename=%date:~0,4%%date:~5,2%%date:~8,2%%time:~0,2%%time:~3,2%%time:~6,2%
 echo %filename%
 go env -w GOPROXY=https://goproxy.cn,direct
 go env -w GO111MODULE=auto
 go mod tidy
 go generate main.go
 go build -ldflags "-s -w -checklinkname=0" -o zbp%filename% -trimpath