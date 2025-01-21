# Set the result of the command 'git rev-list --max-parents=0 HEAD' to a variable called result
$result = git rev-list --max-parents=0 HEAD

# Get the first 7 characters of the result and set this to a variable called COMMIT
$COMMIT = $result.Substring(0,7)

# Remove directory dist\release
Remove-Item dist\release -Recurse -Force

# Create directory dist\release
New-Item -ItemType Directory -Force -Path dist\release

# Create directory .tmp
New-Item -ItemType Directory -Force -Path .tmp

# Create directory .tmp\xgen\v0.9\dist
New-Item -ItemType Directory -Force -Path .tmp\xgen\v0.9\dist

# Write 'XGEN v0.9' to a file called index.html in the directory .tmp\xgen\v0.9\dist
Set-Content -Path .tmp\xgen\v0.9\dist\index.html -Value 'XGEN v0.9'

# echo "yao init"
# & "./build_init.ps1"

# echo "build xgen"
# & "./build_xgen.ps1"

# Create directory .tmp\data\xgen
New-Item -ItemType Directory -Force -Path .tmp\data\xgen

# Copy ui directory to .tmp\data\ui
Copy-Item -Path ui -Destination .tmp\data\ui -Recurse

# Copy yao directory to .tmp\data\yao
Copy-Item -Path yao -Destination .tmp\data\yao -Recurse

# Copy .tmp\xgen\v0.9\dist directory to .tmp\data\xgen\v0.9
Copy-Item -Path .tmp\xgen\v0.9\dist -Destination .tmp\data\xgen\v0.9 -Recurse

# Copy ..\xgen-v1.0\packages\xgen\dist directory to .tmp\data\xgen\v1.0
Copy-Item -Path '..\xgen-v1.0\packages\xgen\dist' -Destination .tmp\data\xgen\v1.0 -Recurse

# Copy ..\xgen-v1.0\packages\setup\build directory to .tmp\data\xgen\setup
# Copy-Item -Path '..\xgen-v1.0\packages\setup\build' -Destination .tmp\data\xgen\setup -Recurse

Copy-Item -Path ".tmp\yao-init" -Destination ".tmp\data\init" -Recurse -Force

Copy-Item -Path "sui\libsui" -Destination ".tmp\data\libsui" -Recurse -Force

# Generate data/bindata.go with go-bindata
go-bindata -fs -pkg data -o data/bindata.go -prefix '.tmp/data/' .tmp/data/...

# Remove directory .tmp\data
Remove-Item .tmp\data -Recurse -Force

# Remove directory .tmp\xgen
Remove-Item .tmp\xgen -Recurse -Force

# Create directory dist
New-Item -ItemType Directory -Force -Path dist

# Remove yao.exe from GOPATH\bin
Remove-Item $env:GOPATH\bin\yao.exe -Force

# Build yao-debug.exe with GOARCH=amd64, GOOS=windows
Set-Item -Path Env:GOARCH -Value 'amd64'
Set-Item -Path Env:GOOS -Value 'windows'

echo "build yao"

go build -v -o dist\release\yao-debug.exe

# Move yao-debug.exe to GOPATH\bin\yao.exe
echo "Move yao"
Move-Item dist\release\yao-debug.exe $env:GOPATH\bin\yao.exe -Force