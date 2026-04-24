//go:build windows

package testhelpers

//go:generate go build -o ../build/test-args.exe -buildvcs=false ./args
//go:generate go build -o ../build/test-echo.exe -buildvcs=false ./echo
//go:generate go build -o ../build/test-crontab.exe -buildvcs=false ./crontab
//go:generate go build -o ../build/test-shell.exe -buildvcs=false ./shell
