# 在集成之前，需要修改cui的环境变量BASE=__yao_admin_root，如果是前端单独测试，设置BASE=yao，或是清空BASE设置

Set-Location ../cui-v1.0/packages/setup
$env:NODE_ENV="development"
pnpm install --no-frozen-lockfile
$env:NODE_ENV="production"
pnpm run build

# $BASE="__yao_admin_root"; Set-Content ../cui-v1.0/packages/cui/.env "BASE=$BASE"
Set-Location ../cui
Remove-Item dist -Recurse -Force
$env:NODE_ENV="development"
pnpm install --no-frozen-lockfile
$env:NODE_ENV="production"
pnpm run build

Set-Location ../../../yao