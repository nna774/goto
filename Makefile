all: app

SAM := sam

app:
	go build

app-for-deploy:
	GOOS=linux go build

deploy: app-for-deploy
	$(SAM) package --region ap-northeast-1 --template-file template.yml --s3-bucket nana-lambda --output-template-file packaged-template.yml
	$(SAM) deploy --template-file packaged-template.yml --region ap-northeast-1 --stack-name goto
