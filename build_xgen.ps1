# 在集成之前，需要修改xgen的环境变量BASE=__yao_admin_root，如果是前端单独测试，设置BASE=yao，或是清空BASE设置

$BASE="__yao_admin_root"; Set-Content ../xgen-v1.0/packages/xgen/.env "BASE=$BASE"
Set-Location ../xgen-v1.0
pnpm install --no-frozen-lockfile
$env:NODE_ENV="production"
pnpm run build
$BASE="yao"; Set-Content ../xgen-v1.0/packages/xgen/.env "BASE=$BASE"
Set-Location ../yao