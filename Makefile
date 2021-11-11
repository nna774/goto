all: app

SAM := sam
REGION := ap-northeast-1
BUCKET := nana-lambda

STACK_NAME := goto

app:
	go build

app-for-deploy:
	GOOS=linux go build

deploy: app-for-deploy
	$(SAM) package --region $(REGION) --template-file template.yml --s3-bucket $(BUCKET) --output-template-file packaged-template.yml
	$(SAM) deploy --region $(REGION) --template-file packaged-template.yml --stack-name $(STACK_NAME)
