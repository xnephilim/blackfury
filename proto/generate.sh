cd proto
buf generate
cd ..

cp -r github.com/ingenuity-build/blackfury/* ./
rm -rf github.com

swagger-combine ./docs/config.json -o ./docs/swagger.yml
rm -rf tmp-swagger-gen
