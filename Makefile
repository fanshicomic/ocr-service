deploy:
	gcloud builds submit . \
      --project=double-voice-460107-e4 \
      --substitutions=_REGION=asia-southeast1,_REPO_NAME=my-go-services,_SERVICE_NAME=ocr-service

local-build:
	export CGO_CFLAGS="-I/opt/homebrew/include";CGO_LDFLAGS="-L/opt/homebrew/lib";export LIBRARY_PATH="/opt/homebrew/lib";export CPATH="/opt/homebrew/include";go build