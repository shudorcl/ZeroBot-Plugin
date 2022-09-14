 set GOOS=linux
 set GOARCH=amd64
 set filename=%date:~0,4%%date:~5,2%%date:~8,2%%time:~0,2%%time:~3,2%%time:~6,2%
 go build -ldflags "-s -w" -o zbp%filename% -trimpath