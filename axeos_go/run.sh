cd ../main/http_server/axe-os
npm run ng build
cd -
cp -a ../main/http_server/axe-os/dist/axe-os/ ./dist
go run .
