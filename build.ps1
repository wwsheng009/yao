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

# Create directory .tmp\cui\v0.9\dist
New-Item -ItemType Directory -Force -Path .tmp\cui\v0.9\dist

# Write 'CUI v0.9' to a file called index.html in the directory .tmp\cui\v0.9\dist
Set-Content -Path .tmp\cui\v0.9\dist\index.html -Value 'CUI v0.9'

# echo "yao init"
# & "./build_init.ps1"

# echo "build cui"
# & "./build_xgen.ps1"

# Create directory .tmp\data\cui
New-Item -ItemType Directory -Force -Path .tmp\data\cui

# Copy ui directory to .tmp\data\ui
Copy-Item -Path ui -Destination .tmp\data\ui -Recurse

# Copy yao directory to .tmp\data\yao
Copy-Item -Path yao -Destination .tmp\data\yao -Recurse

# Copy .tmp\cui\v0.9\dist directory to .tmp\data\cui\v0.9
Copy-Item -Path .tmp\cui\v0.9\dist -Destination .tmp\data\cui\v0.9 -Recurse

# Copy ..\cui-v1.0\packages\cui\dist directory to .tmp\data\cui\v1.0
Copy-Item -Path '..\cui-v1.0\packages\cui\dist' -Destination .tmp\data\cui\v1.0 -Recurse

# Copy ..\cui-v1.0\packages\setup\build directory to .tmp\data\cui\setup
Copy-Item -Path '..\cui-v1.0\packages\setup\build' -Destination .tmp\data\cui\setup -Recurse

Copy-Item -Path "..\yao-init" -Destination ".tmp\data\init" -Recurse -Force

Remove-Item .tmp\data\init\.git -Recurse -Force
Remove-Item .tmp\data\init\node_modules -Recurse -Force
Remove-Item .tmp\data\init\.gitignore -Force
Remove-Item .tmp\data\init\LICENSE -Force
Remove-Item .tmp\data\init\README.md -Force

Copy-Item -Path "sui\libsui" -Destination ".tmp\data\libsui" -Recurse -Force
# 删除 .tmp/data 目录下所有的 .DS_Store 文件
Get-ChildItem -Path ".tmp/data" -Filter ".DS_Store" -Recurse -File | Remove-Item -Force

# Generate data/bindata.go with go-bindata
go-bindata -fs -pkg data -o data/bindata.go -prefix '.tmp/data/' .tmp/data/...

# Remove directory .tmp\data
Remove-Item .tmp\data -Recurse -Force

# Remove directory .tmp\cui
Remove-Item .tmp\cui -Recurse -Force

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