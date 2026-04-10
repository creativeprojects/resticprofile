//go:build !windows

package testhelpers

//go:generate go build -o ../build/test-args -buildvcs=false ./args
//go:generate go build -o ../build/test-echo -buildvcs=false ./echo
//go:generate go build -o ../build/test-crontab -buildvcs=false ./crontab
