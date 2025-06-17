deploy:
	gcloud builds submit . \
      --project=double-voice-460107-e4 \
      --substitutions=_REGION=asia-southeast1,_REPO_NAME=my-go-services,_SERVICE_NAME=ocr-service