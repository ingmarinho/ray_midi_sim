# 1) Set environment variables
$env:CGO_CPPFLAGS = '-O3 -march=native -DNDEBUG -flto'
$env:CGO_LDFLAGS  = '-flto'
$env:CGO_ENABLED = '0'       # only needed if CGO is not already enabled

# 2) Build with the desired flags
go build -ldflags="-s -w -H=windowsgui" -gcflags="-N -l"
