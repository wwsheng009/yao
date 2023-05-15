echo BASE=__yao_admin_root > ../xgen-v1.0/packages/xgen/.env
cd ../xgen-v1.0 && pnpm install --no-frozen-lockfile
set NODE_ENV=production
pnpm run build
cd ../
echo BASE=yao > ../xgen-v1.0/packages/xgen/.env
