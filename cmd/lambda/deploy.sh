# build , zip and deploy

# build
echo "Building lambda function"
GOOS=linux GOARCH=arm64 go build -o bootstrap main.go

# zip
echo "Zipping lambda function"
zip go_lambda.zip bootstrap

# deploy
echo "Deploying lambda function"
aws lambda update-function-code --function-name pixify-transformer --zip-file fileb://go_lambda.zip

