Remove-Item .tmp\yao-init -Recurse -Force
git clone --depth=1 https://github.com/wwsheng009/yao-init-0.10.3.git .tmp\yao-init
Remove-Item .tmp\yao-init\.git -Recurse -Force
Remove-Item .tmp\yao-init\.gitignore
Remove-Item .tmp\yao-init\LICENSE
Remove-Item .tmp\yao-init\README.md